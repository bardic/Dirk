// A generated module for LocalGameci functions
//
// This module has been generated via dagger init and serves as a reference to
// basic module structure as you get started with Dagger.
//
// Two functions have been pre-created. You can modify, delete, or add to them,
// as needed. They demonstrate usage of arguments and return types using simple
// echo and grep commands. The functions can be called from the dagger CLI or
// from one of the SDKs.
//
// The first line in this comment block is a short description line and the
// rest is a long description with more detail on the module's purpose or usage,
// if appropriate. All modules should have a short description.

package main

import (
	"context"
	"dagger/local-gameci/internal/dagger"
	"fmt"
	"strings"
)

type LocalGameci struct {
	Src                                        *dagger.Directory
	Ulf                                        *dagger.File
	User, Platform, BuildTarget, Os, BuildName string
	Pass, Serial                               *dagger.Secret
}

func (m *LocalGameci) Build(ctx context.Context, src *dagger.Directory, user, platform, buildTarget, os, buildName string, pass *dagger.Secret,
	// +optional
	serial *dagger.Secret,
	// +optional
	ulf *dagger.File,
) *dagger.Directory {
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

	libCache := dag.CacheVolume("lib")

	unityVersion, err := m.determineUnityProjectVersion()

	if err != nil {
		return nil
	}

	c := dag.Container().From("unityci/editor:" + os + "-" + unityVersion + "-" + platform + "-3.1.0")

	if ulf != nil {
		fmt.Println("Registering personal license")
		c = m.registerPersonalLicense(c)
	} else {
		fmt.Println("Registering serial license")
		c = m.registerSerialLicense(c)
	}

	c = c.WithDirectory("/src", m.Src).
		WithMountedCache("/src/Library/", libCache)

	c = m.build(c)
	c = m.returnLicense(c)

	return m.getBuildArtifact(c)
}

func (m *LocalGameci) determineUnityProjectVersion() (string, error) {
	s, err := m.Src.File("ProjectSettings/ProjectVersion.txt").Contents(marshalCtx)

	if err != nil {
		return "", err
	}

	v := strings.Split(strings.Split(s, "\n")[0], ": ")[1]

	return v, nil
}

func (m *LocalGameci) build(container *dagger.Container) *dagger.Container {
	cmd := append(m.baseCommand(),
		[]string{
			"-projectPath",
			"/src",
			"-buildTarget",
			m.BuildTarget,
			"-customBuildPath",
			"/src/Builds/",
			"-customBuildName",
			m.BuildName,
			"-customBuildTarget",
			m.BuildTarget,
			"-executeMethod",
			"BuildCommand.PerformBuild",
		}...,
	)

	return container.
		WithExec(cmd,
			dagger.ContainerWithExecOpts{
				Expect: dagger.ReturnTypeAny,
			},
		)
}

func (m *LocalGameci) getBuildArtifact(container *dagger.Container) *dagger.Directory {
	return container.
		Directory("/src/Builds")
}

func (m *LocalGameci) registerPersonalLicense(container *dagger.Container) *dagger.Container {
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

	return container.
		WithFile("/root/.local/share/unity3d/Unity/Unity_lic.ulf", m.Ulf).
		WithExec(cmd,
			dagger.ContainerWithExecOpts{
				Expect: dagger.ReturnTypeAny,
			},
		)
}

func (m *LocalGameci) registerSerialLicense(container *dagger.Container) *dagger.Container {
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

	return container.
		WithExec(cmd,
			dagger.ContainerWithExecOpts{
				Expect: dagger.ReturnTypeAny,
			},
		)
}

func (m *LocalGameci) returnLicense(container *dagger.Container) *dagger.Container {

	cmd := append(m.baseCommand(), []string{"-returnlicense"}...)
	return container.
		WithExec(cmd, dagger.ContainerWithExecOpts{
			Expect: dagger.ReturnTypeAny,
		})
}

func (m *LocalGameci) baseCommand() []string {
	return []string{
		"xvfb-run",
		"--auto-servernum",
		"--server-args='-screen 0 640x480x24'",
		"unity-editor",
		"-quit",
		"-batchmode",
		"-nographics",
	}
}
