// LocalGameci is a Dagger implementation of the GameCI.
// This allows you to build and test Unity projects locally and in CI.
package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/bardic/local_gameci/internal/dagger"
)

type LocalGameci struct {
	Src                                             *dagger.Directory
	Ulf, ServiceConfig, JunitTransform              *dagger.File
	User, Platform, BuildTarget, Os, BuildName      string
	TestingingPlatform, UnityVersion, GameCIVersion string
	Pass, Serial                                    *dagger.Secret
}

func (m *LocalGameci) EnvTest(ctx context.Context,
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

	m.Src = gameSrc
	m.Os = os.Getenv("OS")
	m.Platform = os.Getenv("PLATFORM")
	m.GameCIVersion = os.Getenv("GAMECI_VERSION")
	m.BuildTarget = os.Getenv("BUILD_TARGET")
	m.BuildName = os.Getenv("BUILD_NAME")

	m.Ulf = gameSrc.File(os.Getenv("ULF"))

	var err error
	m.UnityVersion, err = m.determineUnityProjectVersion()

	if err != nil {
		return nil, err
	}

	c := m.createBaseImage()

	c, _ = env.Container(ctx, s, c, true)

	libCache := dag.CacheVolume("lib")

	c = m.register(c)

	c = c.WithDirectory("/src", m.Src).
		WithMountedCache("/src/Library/", libCache)

	c = m.build(c)
	c = m.returnLicense(c)

	err = m.checkForError()

	if err != nil {
		return nil, err
	}

	return m.getBuildArtifact(c), nil
}

func (m *LocalGameci) configureContainerViaEnv(c *dagger.Container) {

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
func (m *LocalGameci) Build(
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
	c := m.configureContainer(src, user, platform, buildTarget, os, buildName, pass, serial, ulf, serviceConfig)
	c = m.build(c)
	c = m.returnLicense(c)

	err := m.checkForError()

	if err != nil {
		return nil
	}

	return m.getBuildArtifact(c)
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

func (m *LocalGameci) Test(
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
	m.Src = src
	m.User = user
	m.Platform = platform
	m.BuildTarget = buildTarget
	m.Os = os
	m.BuildName = buildName
	m.TestingingPlatform = testingingPlatform
	m.Pass = pass
	m.JunitTransform = junitTransform
	m.Serial = serial
	m.Ulf = ulf
	m.ServiceConfig = serviceConfig

	c := m.configureContainer(src, user, platform, buildTarget, os, buildName, pass, serial, ulf, serviceConfig)
	c.WithFile("/nunit-transforms/nunit3-junit.xslt", junitTransform)

	c = m.test(c)

	if junitTransform != nil {
		f := c.File("/results/" + m.TestingingPlatform + "-results.xml")
		jf := m.convertTestsToJUNIT(f, junitTransform)

		c = c.WithFile("/results/"+m.TestingingPlatform+"-junit-results.xml", jf)
	}

	c = m.returnLicense(c)

	err := m.checkForError()

	if err != nil {
		return nil
	}

	return m.getTestResults(c)
}

func (m *LocalGameci) configureContainer(src *dagger.Directory,
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

	m.Src = src
	m.Ulf = ulf
	m.User = user
	m.Platform = platform
	m.BuildTarget = buildTarget
	m.Os = os
	m.BuildName = buildName
	m.Pass = pass
	m.Serial = serial
	m.ServiceConfig = serviceConfig

	var err error
	m.UnityVersion, err = m.determineUnityProjectVersion()

	if err != nil {
		return nil
	}

	c := m.createBaseImage()
	c.WithEnvVariable("CACHEBUSTER", time.Now().String())

	libCache := dag.CacheVolume("lib")

	c = m.register(c)

	c = c.WithDirectory("/src", m.Src).
		WithMountedCache("/src/Library/", libCache)

	return c
}

func (m *LocalGameci) determineUnityProjectVersion() (string, error) {
	s, err := m.Src.File("ProjectSettings/ProjectVersion.txt").Contents(marshalCtx)

	if err != nil {
		return "", err
	}

	v := strings.Split(strings.Split(s, "\n")[0], ": ")[1]

	return v, nil
}

func (m *LocalGameci) build(c *dagger.Container) *dagger.Container {
	cmd := append(m.baseCommand(),
		[]string{
			"-projectPath",
			"/src",
			"-buildTarget",
			m.BuildTarget,
			"-customBuildPath",
			"/builds/",
			"-customBuildName",
			m.BuildName,
			"-customBuildTarget",
			m.BuildTarget,
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

func (m *LocalGameci) test(c *dagger.Container) *dagger.Container {
	cmd := append(m.baseCommand(),
		[]string{
			"-runTests",
			"-testResults",
			"/results/" + m.TestingingPlatform + "-results.xml",
			"-debugCodeOptimization",
			"-enableCodeCoverage",
			"-coverageResultsPath",
			"/results/" + m.TestingingPlatform + "-coverage/",
			"-coverageHistoryPath",
			"/results/" + m.TestingingPlatform + "-coverage-history/",
			"-testPlatform",
			m.TestingingPlatform,
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

func (m *LocalGameci) getBuildArtifact(c *dagger.Container) *dagger.Directory {
	return c.
		Directory("/builds")
}

func (m *LocalGameci) getTestResults(c *dagger.Container) *dagger.Directory {
	return c.
		Directory("/results")
}

func (m *LocalGameci) register(c *dagger.Container) *dagger.Container {
	if m.Ulf != nil {
		fmt.Println("Registering personal license")
		c = m.registerPersonalLicense(c)
	}

	if m.Serial != nil {
		fmt.Println("Registering serial license")
		c = m.registerSerialLicense(c)
	}

	if m.ServiceConfig != nil {
		fmt.Println("Registering license server")
		c = m.registerLicenseServer(c)
	}

	return c
}

func (m *LocalGameci) registerPersonalLicense(c *dagger.Container) *dagger.Container {

	cmd := append(m.baseCommand(),
		[]string{
			"-username",
			"echo ${USER}",
			"-password",
			"echo ${PASS}",
		}...,
	)

	return c.
		WithFile("/root/.local/share/unity3d/Unity/Unity_lic.ulf", m.Ulf).
		WithExec(cmd,
			dagger.ContainerWithExecOpts{
				Expect: dagger.ReturnTypeAny,
			},
		)
}

func (m *LocalGameci) registerSerialLicense(c *dagger.Container) *dagger.Container {
	s, err := m.Serial.Plaintext(marshalCtx)

	if err != nil {
		return nil
	}

	cmd := append(m.baseCommand(),
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

func (m *LocalGameci) registerLicenseServer(c *dagger.Container) *dagger.Container {
	return c.WithFile("/usr/share/unity3d/config/services-config.json", m.ServiceConfig).
		WithExec([]string{
			"sh",
			"-c",
			"/opt/unity/Editor/Data/Resources/Licensing/Client/Unity.Licensing.Client --acquire-floating",
		})
}

func (m *LocalGameci) returnLicense(c *dagger.Container) *dagger.Container {

	cmd := append(m.baseCommand(), []string{"-returnlicense"}...)
	return c.
		WithExec(cmd, dagger.ContainerWithExecOpts{
			Expect: dagger.ReturnTypeAny,
		})
}

func (m *LocalGameci) checkForError() error {
	return nil
}

func (m *LocalGameci) baseCommand() []string {
	return []string{
		"xvfb-run",
		"--auto-servernum",
		"--server-args='-screen 0 640x480x24'",
		"unity-editor",
		"-nographics",
	}
}

func (m *LocalGameci) convertTestsToJUNIT(f, transform *dagger.File) *dagger.File {
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
		WithFile("/results/"+m.TestingingPlatform+"-results.xml", f).
		WithFile("/nunit-transforms/nunit3-junit.xslt", transform).
		WithExec([]string{
			"sh",
			"-c",
			"saxonb-xslt -s /results/" + m.TestingingPlatform + "-results.xml -xsl /nunit-transforms/nunit3-junit.xslt > /results/" + m.TestingingPlatform + "-junit-results.xml",
		}).
		File("/results/" + m.TestingingPlatform + "-junit-results.xml")
}

func (m *LocalGameci) createBaseImage() *dagger.Container {
	return dag.Container().From("unityci/editor:" + m.Os + "-" + m.UnityVersion + "-" + m.Platform + "-" + m.GameCIVersion)
}
