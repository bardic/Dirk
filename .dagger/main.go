// Dirk leverages Dagger.io and GameCI to offer you a platform agnostic unity
// build pipeline. This pipeline can run on any platform that supports Docker.
package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/bardic/Dirk/internal/dagger"
)

// Dirk
type Dirk struct {
	BuildName          string            // Unity Build Name
	BuildTarget        string            // Unity Build Target
	GameciVersion      string            // GameCI Version
	JunitTransform     *dagger.File      // Junit Transform Path
	Os                 string            // GameCI base OS
	Pass               *dagger.Secret    // Unity Account Password
	Platform           string            // Unity Build Target Platform
	Serial             *dagger.Secret    // Unity Serial
	ServiceConfig      *dagger.File      // Unity Service Config for Licesning Server
	Src                *dagger.Directory // Source directory of the Unity project
	TestingingPlatform string            //If should test as editor or playback
	Ulf                *dagger.File      // Unity Personal License File
	UnityVersion       string            // Unity Version that GameCI should use
	User               string            // Unity Account Username

}

// Build the things
func (d *Dirk) Build(
	ctx context.Context,
	gameSrc *dagger.Directory,
	// +optional
	buildName string,
	// +optional
	buildTarget string,
	// +optional
	gameciVersion string,
	// +optional
	isTest bool,
	// +optional
	junitTransform *dagger.File,
	// +optional
	pass *dagger.Secret,
	// +optional
	platform string,
	// +optional
	serial *dagger.Secret,
	// +optional
	serviceConfig *dagger.File,
	// +optional
	targetOs string,
	// +optional
	testingingPlatform string,
	// +optional
	ulf *dagger.File,
	// +optional
	unityVersion string,
	// +optional
	user string,
) (*dagger.Directory, error) {
	gameSrc = gameSrc.WithoutDirectory(".git")
	gameSrc = gameSrc.WithoutDirectory(".dagger")
	gameSrc = gameSrc.WithoutDirectory(".vscode")
	gameSrc = gameSrc.WithoutFiles([]string{".gitignore", ".gitmodules", ".DS_Store", "dagger.json", "go.work", "LICENSE", "README.md"})

	d.Src = gameSrc

	var err error
	d.UnityVersion, err = d.determineUnityProjectVersion()

	if err != nil {
		return nil, err
	}

	var f, s *dagger.File

	if isTest {
		f = gameSrc.File("./unity_test.env")
	} else {
		f = gameSrc.File("./unity.env")
	}

	if f != nil {
		NewHostEnv(context.Background(), f)
	} else {
		return nil, errors.New("unity.env file is required")
	}

	d.BuildName = os.Getenv("DIRK_BUILD_NAME")
	d.BuildTarget = os.Getenv("DIRK_BUILD_TARGET")
	d.GameciVersion = os.Getenv("DIRK_GAMECI_VERSION")
	d.TestingingPlatform = os.Getenv("DIRK_TESTING_PLATFORM")

	d.Os = os.Getenv("DIRK_OS")

	if _, b := os.LookupEnv("DIRK_JUNIT_TRANSFORM"); b {
		d.JunitTransform = gameSrc.File(os.Getenv("DIRK_JUNIT_TRANSFORM"))
	}

	if _, b := os.LookupEnv("DIRK_PASS"); b {
		d.Pass = dag.Secret(os.Getenv("DIRK_PASS"))
	}
	d.Platform = os.Getenv("DIRK_PLATFORM")

	if _, b := os.LookupEnv("DIRK_SERIAL"); b {
		d.Serial = dag.Secret(os.Getenv("DIRK_SERIAL"))
	}

	if _, b := os.LookupEnv("DIRK_SERVICE_CONFIG"); b {
		d.ServiceConfig = gameSrc.File(os.Getenv("DIRK_SERVICE_CONFIG"))
	}

	if _, b := os.LookupEnv("DIRK_ULF"); b {
		d.Ulf = gameSrc.File(os.Getenv("DIRK_ULF"))
	}

	if _, b := os.LookupEnv("DIRK_UNITY_VERSION"); b {
		d.UnityVersion = os.Getenv("DIRK_UNITY_VERSION")
	}

	d.User = os.Getenv("DIRK_USER")

	if buildName != "" {
		d.BuildName = buildName
	}

	if buildTarget != "" {
		d.BuildTarget = buildTarget
	}

	if gameciVersion != "" {
		d.GameciVersion = gameciVersion
	}

	if pass != nil {
		d.Pass = pass
	}

	if platform != "" {
		d.Platform = platform
	}

	if serial != nil {
		d.Serial = serial
	}

	if serviceConfig != nil {
		d.ServiceConfig = serviceConfig
	}

	if targetOs != "" {
		d.Os = targetOs
	}

	if ulf != nil {
		d.Ulf = ulf
	}

	if unityVersion != "" {
		d.UnityVersion = unityVersion
	}

	if user != "" {
		d.User = user
	}

	err = d.validateConfig(isTest)

	if err != nil {
		return nil, err
	}

	c := d.createBaseImage()

	if isTest {
		s = gameSrc.File("./unity_test_secrets.env")
	} else {
		s = gameSrc.File("./unity_secrets.env")
	}

	if s != nil {
		c, _ = NewContainerEnv(ctx, s, c, true)
	}

	c = d.register(c)

	c = c.
		WithMountedCache("/src/Library/", dag.CacheVolume("lib")).
		WithDirectory("/src", d.Src)

	if isTest {
		c = d.test(c)

		if junitTransform != nil {
			f := c.File("/results/" + d.TestingingPlatform + "-results.xml")
			jf := d.convertTestsToJUNIT(f, junitTransform)

			c.WithFile("/nunit-transforms/nunit3-junit.xslt", junitTransform)
			c = c.WithFile("/results/"+d.TestingingPlatform+"-junit-results.xml", jf)
		}
	} else {

		c = d.build(c)
	}

	c = d.returnLicense(c)

	err = d.checkForError()

	if err != nil {
		return nil, err
	}

	if isTest {
		return d.getTestResults(c), nil
	}

	return d.getBuildArtifact(c), nil
}

func (d *Dirk) determineUnityProjectVersion() (string, error) {
	ctx := context.Background()
	s, err := d.Src.File("ProjectSettings/ProjectVersion.txt").Contents(ctx)

	if err != nil {
		return "", err
	}

	v := strings.Split(strings.Split(s, "\n")[0], ": ")[1]

	return v, nil
}

func (d *Dirk) build(c *dagger.Container) *dagger.Container {
	cmd := append(d.baseCommand(),
		[]string{
			"-projectPath",
			"/src",
			"-buildTarget",
			d.BuildTarget,
			"-customBuildPath",
			"/builds/",
			"-customBuildName",
			d.BuildName,
			"-customBuildTarget",
			d.BuildTarget,
			"-quit",
			"-executeMethod",
			"BuildCommand.PerformBuild",
			"-logFile",
			"/builds/unity.log",
		}...,
	)

	return c.
		WithExec(cmd,
			dagger.ContainerWithExecOpts{
				Expect: dagger.ReturnTypeAny,
			},
		)
}

func (d *Dirk) test(c *dagger.Container) *dagger.Container {
	cmd := append(d.baseCommand(),
		[]string{
			"-projectPath",
			"/src",
			"-runTests",
			"-testResults",
			"/results/" + d.TestingingPlatform + "-results.xml",
			"-debugCodeOptimization",
			"-enableCodeCoverage",
			"-coverageResultsPath",
			"/results/" + d.TestingingPlatform + "-coverage/",
			"-coverageHistoryPath",
			"/results/" + d.TestingingPlatform + "-coverage-history/",
			"-testPlatform",
			d.TestingingPlatform,
			"-coverageOptions",
			"'generateAdditionalMetrics;generateHtmlReport;generateHtmlReportHistory;generateBadgeReport;verbosity:verbose'",
			"-logFile",
			"/results/unity.log",
		}...)

	return c.
		WithExec(cmd,
			dagger.ContainerWithExecOpts{
				Expect: dagger.ReturnTypeAny,
			},
		)
}

func (d *Dirk) getBuildArtifact(c *dagger.Container) *dagger.Directory {
	return c.
		Directory("/builds")
}

func (d *Dirk) getTestResults(c *dagger.Container) *dagger.Directory {
	return c.
		Directory("/results")
}

func (d *Dirk) register(c *dagger.Container) *dagger.Container {
	if d.Ulf != nil {
		fmt.Println("Registering personal license")
		c = d.registerPersonalLicense(c)
	}

	if d.Serial != nil {
		fmt.Println("Registering serial license")
		c = d.registerSerialLicense(c)
	}

	if d.ServiceConfig != nil {
		fmt.Println("Registering license server")
		c = d.registerLicenseServer(c)
	}

	return c
}

func (d *Dirk) registerPersonalLicense(c *dagger.Container) *dagger.Container {

	cmd := append(d.baseCommand(),
		[]string{
			"-username",
			"echo ${USER}",
			"-password",
			"echo ${PASS}",
		}...,
	)

	return c.
		WithFile("/root/.local/share/unity3d/Unity/Unity_lic.ulf", d.Ulf).
		WithExec(cmd,
			dagger.ContainerWithExecOpts{
				Expect: dagger.ReturnTypeAny,
			},
		)
}

func (d *Dirk) registerSerialLicense(c *dagger.Container) *dagger.Container {
	ctx := context.Background()
	s, err := d.Serial.Plaintext(ctx)

	if err != nil {
		return nil
	}

	cmd := append(d.baseCommand(),
		[]string{
			"-username",
			"echo ${USER}",
			"-password",
			"echo ${PASS}",
			"-serial",
			s,
		}...,
	)

	return c.
		WithExec(cmd,
			dagger.ContainerWithExecOpts{
				Expect: dagger.ReturnTypeAny,
			},
		)
}

func (d *Dirk) registerLicenseServer(c *dagger.Container) *dagger.Container {
	return c.WithFile("/usr/share/unity3d/config/services-config.json", d.ServiceConfig).
		WithExec([]string{
			"sh",
			"-c",
			"/opt/unity/Editor/Data/Resources/Licensing/Client/Unity.Licensing.Client --acquire-floating",
		})
}

func (d *Dirk) returnLicense(c *dagger.Container) *dagger.Container {

	cmd := append(d.baseCommand(), []string{"-returnlicense"}...)
	return c.
		WithExec(cmd, dagger.ContainerWithExecOpts{
			Expect: dagger.ReturnTypeAny,
		})
}

func (d *Dirk) checkForError() error {
	return nil
}

func (d *Dirk) baseCommand() []string {
	return []string{
		"xvfb-run",
		"--auto-servernum",
		"--server-args='-screen 0 640x480x24'",
		"unity-editor",
		"-nographics",
	}
}

func (d *Dirk) convertTestsToJUNIT(f, transform *dagger.File) *dagger.File {
	return dag.Container().From("eclipse-temurin").
		WithExec([]string{
			"apt-get",
			"update",
		}).
		WithExec([]string{
			"apt-get",
			"install",
			"-y",
			"libsaxonb-java",
		}).
		WithFile("/results/"+d.TestingingPlatform+"-results.xml", f).
		WithFile("/nunit-transforms/nunit3-junit.xslt", transform).
		WithExec([]string{
			"sh",
			"-c",
			"saxonb-xslt -s /results/" + d.TestingingPlatform + "-results.xml -xsl /nunit-transforms/nunit3-junit.xslt > /results/" + d.TestingingPlatform + "-junit-results.xml",
		}).
		File("/results/" + d.TestingingPlatform + "-junit-results.xml")
}

func (d *Dirk) createBaseImage() *dagger.Container {
	return dag.Container().From("unityci/editor:" + d.Os + "-" + d.UnityVersion + "-" + d.Platform + "-" + d.GameciVersion)
}

func (d *Dirk) validateConfig(isTest bool) error {

	if d.User == "" {
		return errors.New("unity username is required")
	}

	if (d.Serial == nil && d.Pass == nil) && d.Ulf == nil {
		return errors.New("unity serial, user and pass or ulf required")
	}

	if d.Os == "" {
		return errors.New("target os is required")
	}

	if d.Platform == "" {
		return errors.New("platform is required")
	}

	if d.UnityVersion == "" {
		return errors.New("unity version is required")
	}

	if isTest {
		if d.TestingingPlatform == "" {
			return errors.New("testing platform is required")
		}
	} else {
		if d.BuildTarget == "" {
			return errors.New("build target is required")
		}

		if d.BuildName == "" {
			return errors.New("build name is required")
		}
	}

	return nil
}
