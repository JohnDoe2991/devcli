package main

import (
	"os"

	"github.com/alexflint/go-arg"
	devcontainerspec "github.com/johndoe2991/devcli/devcontainer_spec"
	"github.com/johndoe2991/devcli/docker"
	"github.com/johndoe2991/devcli/logging"
)

type CleanCmd struct {
	All    bool `arg:"--all" help:"delete all devcontainer and image versions for the current working directory"`
	Global bool `arg:"--global" help:"delete all devcontainers and images created by devcli"`
}

type Args struct {
	Debug bool      `arg:"-d,--debug" help:"activate debug outputs"`
	Logs  bool      `arg:"-l, --log" help:"create a log file in the current directory with all outputs"`
	Clean *CleanCmd `arg:"subcommand:clean" help:"delete image and container"`
}

func (Args) Version() string {
	return "0.1.0"
}

func main() {
	var args Args
	arg.MustParse(&args)

	// setup logger
	if args.Logs {
		logging.WriteToLogFile(true)
	}
	if args.Debug {
		logging.SetLevelFromString("debug")
	}
	logger := logging.GetLogger("main")

	cwd, err := os.Getwd()
	if err != nil {
		logger.Fatal().Err(err).Msg("could not get current working directory")
	}
	logger.Debug().Str("cwd", cwd).Msg("current working directory")

	switch {
	default:
		// default command without anything; start devcontainer
		// get devcontainer setup
		devc, err := devcontainerspec.ParseDevcontainer(cwd)
		if err != nil {
			logger.Fatal().Err(err).Msg("could not get devcontainer setup")
		}
		err = docker.Run(devc)
		if err != nil {
			logger.Fatal().Err(err).Msg("could not run devcontainer")
		}
	case args.Clean != nil:
		if args.Clean.Global {
			err := docker.CleanAllContainers()
			if err != nil {
				logger.Fatal().Err(err).Msg("could not clean all containers")
			}
			err = docker.CleanAllImages()
			if err != nil {
				logger.Fatal().Err(err).Msg("could not clean all images")
			}
			return
		}
		devc, err := devcontainerspec.ParseDevcontainer(cwd)
		if err != nil {
			logger.Fatal().Err(err).Msg("could not get devcontainer setup")
		}
		if args.Clean.All {
			err := docker.CleanAllContainerVersions(devc)
			if err != nil {
				logger.Fatal().Err(err).Str("basename", devc.GetDevcNamePrefix()).Msg("could not delete all containers for this working directory")
			}
			err = docker.CleanAllImageVersions(devc)
			if err != nil {
				logger.Fatal().Err(err).Str("basename", devc.GetDevcNamePrefix()).Msg("could not delete all images for this working directory")
			}
		} else {
			err := docker.CleanContainer(devc.GetContainerName())
			if err != nil {
				logger.Fatal().Err(err).Str("container name", devc.GetContainerName()).Msg("could not delete container")
			}
			err = docker.CleanImage(devc.GetImageName())
			if err != nil {
				logger.Fatal().Err(err).Str("image name", devc.GetImageName()).Msg("could not delete image")
			}
		}
	}
}
