//go:build mage
// +build mage

package build

import (
	"os"
	"os/exec"
)

var (
	examplEnv = `DIRK_BUILD_NAME=
DIRK_BUILD_TARGET=
DIRK_GAMECI_VERSION=
DIRK_PLATFORM=
DIRK_OS=
DIRK_USER=
DIRK_ULF=./License/Unity_v6000.x.ulf
DIRK_JUNIT_TRANSFORM=./nunit-transforms/nunit3-junit.xslt`

	secretExampleEnv = `DIRK_USER=
DIRK_PASS=
DIRK_SERIAL=`
)

func Config() error {
	err := os.WriteFile("dirk.env", []byte(examplEnv), 0644)

	if err != nil {
		return err
	}

	err = os.WriteFile("dirk.secret.env", []byte(secretExampleEnv), 0644)

	if err != nil {
		return err
	}

	return nil
}

func Build(src, gameciVersion, buildOs, targetPlatform string) error {
	cmd := exec.Command(
		"dagger",
		"call",
		"build",
		"--game-src="+src,
		"--gameci-version="+gameciVersion,
		"--target-os="+buildOs,
		"--platform="+targetPlatform,
		"--path=./builds",
	)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	return err
}
