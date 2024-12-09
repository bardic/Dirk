// Dirk is a Dagger implementation of the GameCI.
// This allows you to build and test Unity projects locally and in CI.
package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/bardic/Dirk/internal/dagger"
)

type Dirk struct {
	Src                                             *dagger.Directory
	Ulf, ServiceConfig, JunitTransform              *dagger.File
	User, Platform, BuildTarget, Os, BuildName      string
	TestingingPlatform, UnityVersion, GameCIVersion string
	Pass, Serial                                    *dagger.Secret
}

func (d *Dirk) EnvTest(ctx context.Context,
	gameSrc *dagger.Directory,
) (*dagger.Directory, error) {
	fmt.Printf("EnvTest	\n")

	var f, s *dagger.File

	f = gameSrc.File("./unity.env")
	s = gameSrc.File("./unity_secrets.env")

	env := Env{}
	env.Host(ctx, f)

	gameSrc = gameSrc.WithoutDirectory(".git")
	gameSrc = gameSrc.WithoutDirectory(".dagger")
	gameSrc = gameSrc.WithoutDirectory(".vscode")
	gameSrc = gameSrc.WithoutFiles([]string{".gitignore", ".gitmodules", ".DS_Store", "dagger.json", "go.work", "LICENSE", "README.md"})

	d.Src = gameSrc
	d.Os = os.Getenv("OS")
	d.Platform = os.Getenv("PLATFORM")
	d.GameCIVersion = os.Getenv("GAMECI_VERSION")
	d.BuildTarget = os.Getenv("BUILD_TARGET")
	d.BuildName = os.Getenv("BUILD_NAME")

	d.Ulf = gameSrc.File(os.Getenv("ULF"))

	var err error
	d.UnityVersion, err = d.determineUnityProjectVersion()

	if err != nil {
		return nil, err
	}

	c := d.createBaseImage()

	c, _ = env.Container(ctx, s, c, true)

	libCache := dag.CacheVolume("lib")

	c = d.register(c)

	c = c.WithDirectory("/src", d.Src).
		WithMountedCache("/src/Library/", libCache)

	c = d.build(c)
	c = d.returnLicense(c)

	err = d.checkForError()

	if err != nil {
		return nil, err
	}

	return d.getBuildArtifact(c), nil
}

/*
Build takes a source directory and builds the Unity project within it.
Usage:

	Build(src, user, platform, buildTarget, os, buildName, pass, serial, ulf, serviceConfig)
		src: *dagger.Directory
		user: string
		platform: string
		buildTarget: string
		os: string
		buildName: string
		pass: *dagger.Secret
		// +optional
		serial: *dagger.Secret
		// +optional
		ulf: *dagger.File
		// +optional
		serviceConfig: *dagger.File

Returns:

	*dagger.Directory

Example:

	// Build unity project with a personal license targeting Windows Mono on Ubuntu
	dagger call test --src="./example/game" \
		--ulf="./Unity_v6000.x.ulf" \
		--build-target="StandaloneWindows64" \
		--build-name="demo" \
		--platform="windows-mono" \
		--os="ubuntu" \
		--user=env:USER \
		--pass=env:PASS \
		export ./builds

	// Build unity project with a User and Serail targeting Windows Mono on Ubuntu
	dagger call test --src="./example/game" \
		--build-target="StandaloneWindows64" \
		--build-name="demo" \
		--platform="windows-mono" \
		--os="ubuntu" \
		--user=env:USER \
		--pass=env:PASS \
		--serial=env:SERIAL \
		export ./builds

	// Build unity project with Service Config (float license) targeting Windows Mono on Ubuntu
	dagger call test --src="./example/game" \
		--build-target="StandaloneWindows64" \
		--build-name="demo" \
		--platform="windows-mono" \
		--os="ubuntu" \
		--user=env:USER \
		--pass=env:PASS \
		--service-config="./service-config.json" \
		export ./builds
*/
func (d *Dirk) Build(
	src *dagger.Directory,
	user, platform, buildTarget, os, buildName string,
	pass *dagger.Secret,
	// +optional
	serial *dagger.Secret,
	// +optional
	ulf *dagger.File,
	// +optional
	serviceConfig *dagger.File,
) *dagger.Directory {
	c := d.configureContainer(src, user, platform, buildTarget, os, buildName, pass, serial, ulf, serviceConfig)
	c = d.build(c)
	c = d.returnLicense(c)

	err := d.checkForError()

	if err != nil {
		return nil
	}

	return d.getBuildArtifact(c)
}

/*
Test takes a source directory and tests the Unity project within it.
Usage:

	Test(src, user, platform, buildTarget, os, buildName, testingingPlatform, pass, junitTransform, serial, ulf, serviceConfig)
		src: *dagger.Directory
		user: string
		platform: string
		buildTarget: string
		os: string
		buildName: string
		testingingPlatform: string
		pass: *dagger.Secret
		// +optional
		junitTransform: *dagger.File
		// +optional
		serial: *dagger.Secret
		// +optional
		ulf: *dagger.File
		// +optional
		serviceConfig: *dagger.File

Returns:

	*dagger.Directory

Example:

	// Test unity project with a personal license targeting Windows Mono on Ubuntu
	dagger call test \
		--src="./example/game" \
		--user=env:USER \
		--platform="windows-mono" \
		--build-target="StandaloneWindows64" \
		--os="ubuntu" \
		--build-name="demo" \
		--testinging-platform="editor" \
		--pass=env:PASS \
		--junitTransform="/nunit-transforms/nunit3-junit.xslt" \
		--ulf="./Unity_v6000.x.ulf" \
		export ./results

*/

func (d *Dirk) Test(
	src *dagger.Directory,
	user string,
	platform string,
	buildTarget string,
	os string,
	buildName string,
	testingingPlatform string,
	pass *dagger.Secret,
	// +optional
	junitTransform *dagger.File,
	// +optional
	serial *dagger.Secret,
	// +optional
	ulf *dagger.File,
	// +optional
	serviceConfig *dagger.File,
) *dagger.Directory {
	d.Src = src
	d.User = user
	d.Platform = platform
	d.BuildTarget = buildTarget
	d.Os = os
	d.BuildName = buildName
	d.TestingingPlatform = testingingPlatform
	d.Pass = pass
	d.JunitTransform = junitTransform
	d.Serial = serial
	d.Ulf = ulf
	d.ServiceConfig = serviceConfig

	c := d.configureContainer(src, user, platform, buildTarget, os, buildName, pass, serial, ulf, serviceConfig)
	c.WithFile("/nunit-transforms/nunit3-junit.xslt", junitTransform)

	c = d.test(c)

	if junitTransform != nil {
		f := c.File("/results/" + d.TestingingPlatform + "-results.xml")
		jf := d.convertTestsToJUNIT(f, junitTransform)

		c = c.WithFile("/results/"+d.TestingingPlatform+"-junit-results.xml", jf)
	}

	c = d.returnLicense(c)

	err := d.checkForError()

	if err != nil {
		return nil
	}

	return d.getTestResults(c)
}

func (d *Dirk) configureContainer(src *dagger.Directory,
	user, platform, buildTarget, os, buildName string,
	pass *dagger.Secret,
	// +optional
	serial *dagger.Secret,
	// +optional
	ulf *dagger.File,
	// +optional
	serviceConfig *dagger.File,
) *dagger.Container {
	src = src.WithoutDirectory(".git")
	src = src.WithoutDirectory(".dagger")
	src = src.WithoutDirectory(".vscode")
	src = src.WithoutFiles([]string{".gitignore", ".gitmodules", ".DS_Store", "dagger.json", "go.work", "LICENSE", "README.md"})

	d.Src = src
	d.Ulf = ulf
	d.User = user
	d.Platform = platform
	d.BuildTarget = buildTarget
	d.Os = os
	d.BuildName = buildName
	d.Pass = pass
	d.Serial = serial
	d.ServiceConfig = serviceConfig

	var err error
	d.UnityVersion, err = d.determineUnityProjectVersion()

	if err != nil {
		return nil
	}

	c := d.createBaseImage()
	c.WithEnvVariable("CACHEBUSTER", time.Now().String())

	libCache := dag.CacheVolume("lib")

	c = d.register(c)

	c = c.WithDirectory("/src", d.Src).
		WithMountedCache("/src/Library/", libCache)

	return c
}

func (d *Dirk) determineUnityProjectVersion() (string, error) {
	s, err := d.Src.File("ProjectSettings/ProjectVersion.txt").Contents(marshalCtx)

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
	s, err := d.Serial.Plaintext(marshalCtx)

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
	return dag.Container().From("unityci/editor:" + d.Os + "-" + d.UnityVersion + "-" + d.Platform + "-" + d.GameCIVersion)
}
