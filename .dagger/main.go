package main

import (
	"context"
	"dagger/local-gameci/internal/dagger"
	"fmt"
	"strings"
	"time"
)

type LocalGameci struct {
	Src                                        *dagger.Directory
	Ulf                                        *dagger.File
	User, Platform, BuildTarget, Os, BuildName string
	Pass, Serial                               *dagger.Secret
}

//dagger call build

func (m *LocalGameci) Build(ctx context.Context,
	src *dagger.Directory,
	user, platform, buildTarget, os, buildName string,
	pass *dagger.Secret,
	// +optional
	serial *dagger.Secret,
	// +optional
	ulf *dagger.File,
) *dagger.Directory {
	c := m.createBaseContainer(src, user, platform, buildTarget, os, buildName, pass, serial, ulf)

	c = m.build(c)
	c = m.returnLicense(c)

	err := m.checkForError()

	if err != nil {
		return nil
	}

	return m.getBuildArtifact(c)
}

//dagger call test

func (m *LocalGameci) Test(ctx context.Context,
	src *dagger.Directory,
	user, platform, buildTarget, os, buildName, testingingPlatform string,
	pass *dagger.Secret,
	// +optional
	junit bool,
	// +optional
	serial *dagger.Secret,
	// +optional
	ulf *dagger.File,
) *dagger.Container {
	c := m.createBaseContainer(src, user, platform, buildTarget, os, buildName, pass, serial, ulf)

	c = m.test(c, testingingPlatform)

	if junit {
		c = m.convertTestsToJUNIT(c)
	}

	c = m.returnLicense(c)

	err := m.checkForError()

	if err != nil {
		return nil
	}

	return c //m.getBuildArtifact(c)
}

func (m *LocalGameci) createBaseContainer(src *dagger.Directory, user, platform, buildTarget, os, buildName string, pass *dagger.Secret,
	// +optional
	serial *dagger.Secret,
	// +optional
	ulf *dagger.File) *dagger.Container {
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

	unityVersion, err := m.determineUnityProjectVersion()

	if err != nil {
		return nil
	}

	c := dag.Container().From("unityci/editor:" + os + "-" + unityVersion + "-" + platform + "-3.1.0")
	c.WithEnvVariable("CACHEBUSTER", time.Now().String())

	libCache := dag.CacheVolume("lib")

	c = m.register(c, serial, ulf)

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

func (m *LocalGameci) test(c *dagger.Container, testingingPlatform string) *dagger.Container {
	cmd := append(m.baseCommand(),
		[]string{
			"-runTests",
			"-testResults",
			"/results/" + testingingPlatform + "-results.xml",
			"-debugCodeOptimization",
			"-enableCodeCoverage",
			"-coverageResultsPath",
			"/results/" + testingingPlatform + "-coverage/",
			"-coverageHistoryPath",
			"/results/" + testingingPlatform + "-coverage-history/",
			"-testPlatform",
			testingingPlatform,
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

func (m *LocalGameci) register(c *dagger.Container, serial *dagger.Secret, ulf *dagger.File) *dagger.Container {
	if ulf != nil {
		fmt.Println("Registering personal license")
		c = m.registerPersonalLicense(c)
	}

	if serial != nil {
		fmt.Println("Registering serial license")
		c = m.registerSerialLicense(c)
	}

	return c
}

func (m *LocalGameci) registerPersonalLicense(c *dagger.Container) *dagger.Container {
	p, err := m.Pass.Plaintext(marshalCtx)

	if err != nil {
		return nil
	}

	cmd := append(m.baseCommand(),
		[]string{
			"-username",
			m.User,
			"-password",
			p,
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

	p, err := m.Pass.Plaintext(marshalCtx)

	if err != nil {
		return nil
	}

	cmd := append(m.baseCommand(),
		[]string{
			"-username",
			m.User,
			"-password",
			p,
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
		"-projectPath",
		"/src",
	}
}

func (m *LocalGameci) convertTestsToJUNIT(c *dagger.File) *dagger.File {

	return dag.Container().From("openjdk").
		WithExec([]string{
			"apt-get",
			"update",
			"&&",
			"apt-get,",
			"install",
			"-,y",
			"libsaxonb-java",
		}).
		WithExec([]string{
			"saxonb-xslt",
			"-s",
			"/results/" + m.Platform + "-results.xml",
			"-xsl",
			"nunit-transforms/nunit3-junit.xslt",
			">",
			"/results/" + m.Platform + "-junit-results.xml",
		}).File("/results/" + m.Platform + "-junit-results.xml")
}
