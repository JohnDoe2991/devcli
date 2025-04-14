package main

import (
	"os"

	devcontainerspec "github.com/johndoe2991/devcli/devcontainer_spec"
	"github.com/johndoe2991/devcli/docker"
	"github.com/johndoe2991/devcli/logging"
)

func main() {
	// setup logger
	logger := logging.GetLogger("main")
	logging.SetLevelFromString("debug")

	cwd, err := os.Getwd()
	if err != nil {
		logger.Fatal().Err(err).Msg("could not get current working directory")
	}
	logger.Debug().Str("cwd", cwd).Msg("current working directory")
	// get devcontainer setup
	devc, err := devcontainerspec.ParseDevcontainer(cwd)
	if err != nil {
		logger.Fatal().Err(err).Msg("could not get devcontainer setup")
	}
	err = docker.Run(devc)
	if err != nil {
		logger.Fatal().Err(err).Msg("could not run devcontainer")
	}
}
