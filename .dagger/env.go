package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/bardic/Dirk/internal/dagger"
)

// End struct
type Env struct{}

func NewHostEnv(ctx context.Context, f *dagger.File) error {
	Env := &Env{}
	return Env.host(ctx, f)
}

func NewContainerEnv(ctx context.Context, f *dagger.File, c *dagger.Container, isSecrets bool) (*dagger.Container, error) {
	Env := &Env{}
	return Env.container(ctx, f, c, isSecrets)
}

// host
func (e *Env) host(ctx context.Context, f *dagger.File) error {

	envs, err := f.Contents(ctx)

	if err != nil {
		return err
	}

	envPair := strings.Split(envs, "\n")

	for _, v := range envPair {

		envVals := strings.SplitN(v, "=", 2)
		err := os.Setenv(envVals[0], envVals[1])

		if err != nil {
			return err
		}
	}

	return nil
}

// container
func (e *Env) container(ctx context.Context, f *dagger.File, c *dagger.Container,
	// +optional
	isSecrets bool,
) (*dagger.Container, error) {
	envs, err := f.Contents(ctx)

	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
	}

	envPair := strings.Split(envs, "\n")

	for _, v := range envPair {
		envVals := strings.SplitN(v, "=", 2)

		if isSecrets {
			fmt.Println("Secret found")
			c.WithSecretVariable(envVals[0], dag.SetSecret(envVals[0], envVals[1]))

		} else {
			fmt.Println("Env found")
			c = c.WithEnvVariable(envVals[0], envVals[1])
		}
	}

	return c, nil
}
