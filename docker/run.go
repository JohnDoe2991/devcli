package docker

import (
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"time"

	devcontainerspec "github.com/johndoe2991/devcli/devcontainer_spec"
)

func Run(devc devcontainerspec.Devcontainer) error {
	containerName := devc.GetContainerName()
	// first check if a container is already running
	running, err := checkContainerRunning(containerName)
	if err != nil {
		return err
	}
	if running {
		// container is running, nothing to do
	} else {
		// check if a container already exists
		exists, err := checkContainerExists(containerName)
		if err != nil {
			return err
		}
		logger.Debug().Str("container", containerName).Bool("exists", exists).Msg("checking if container exists")
		if exists {
			// container exists, we start it
			if err := startContainer(containerName); err != nil {
				return err
			}
		} else {
			// container does not exist, we create it
			if err := Build(devc); err != nil {
				return err
			}
			if err := createAndStartContainer(devc); err != nil {
				return err
			}
			// first run, so we have to exec postCreateCommand
			time.Sleep(2 * time.Second) // wait for the container to be ready
			if devc.Config.PostCreateCommand != "" {
				// exec into the container
				if err := execCommand(containerName, false, false, []string{"sh", "-c", devc.Config.PostCreateCommand}); err != nil {
					return err
				}
			}
		}
	}
	// exec into the container
	time.Sleep(1 * time.Second) // wait for the container to be ready
	if devc.Config.PostStartCommand != "" {
		// exec into the container
		logger.Debug().Str("container", containerName).Str("postStartCommand", devc.Config.PostStartCommand).Msg("executing postStartCommand")
		if err := execCommand(containerName, false, false, []string{"sh", "-c", devc.Config.PostStartCommand}); err != nil {
			return err
		}
	}
	logger.Debug().Str("container", containerName).Msg("exec into container")
	if err := execCommand(containerName, true, true, []string{"/bin/bash"}); err != nil {
		return err
	}
	return nil
}

func checkContainerRunning(containerName string) (bool, error) {
	cmd := exec.Command("docker", "ps", "-q", "-f", "name="+containerName)
	output, err := cmd.Output()
	if err != nil {
		return false, err
	}
	return len(output) > 0, nil
}

func checkContainerExists(containerName string) (bool, error) {
	cmd := exec.Command("docker", "ps", "-aq", "-f", "name="+containerName)
	output, err := cmd.Output()
	if err != nil {
		return false, err
	}
	return len(output) > 0, nil
}

func startContainer(containerName string) error {
	// start the container
	cmd := exec.Command("docker", "start", containerName)

	logger.Debug().Str("container", containerName).Msg("starting container")
	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

func stopContainer(containerName string) error {
	// stop the container
	cmd := exec.Command("docker", "stop", containerName)

	logger.Debug().Str("container", containerName).Msg("stopping container")
	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

func createAndStartContainer(devc devcontainerspec.Devcontainer) error {
	// run the container
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	args := []string{"run", "-d", "--name", devc.GetContainerName(), "--volume", cwd + ":/workspaces/" + filepath.Base(cwd)}
	for _, mount := range devc.Config.Mounts {
		args = append(args, "--mount", mount)
	}
	args = append(args, devc.Config.RunArgs...)
	imageName := devc.GetImageName()
	//we keep the container running with a sleep so we can exec into it later
	args = append(args, imageName, "/bin/bash", "-c", "while true; do sleep 5; done;")
	cmd := exec.Command("docker", args...)
	logger.Debug().Str("image", imageName).Strs("args", args).Msg("running image")
	err = cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

func execCommand(containerName string, interactive bool, asUser bool, args []string) error {
	// exec into the container
	cmdargs := []string{"exec"}
	if interactive {
		cmdargs = append(cmdargs, "-it")
	}
	if asUser {
		currentUser, err := user.Current()
		if err != nil {
			return err
		}
		cmdargs = append(cmdargs, "--user", currentUser.Uid+":"+currentUser.Gid)
	}
	cmdargs = append(cmdargs, containerName)
	cmdargs = append(cmdargs, args...)
	cmd := exec.Command("docker", cmdargs...)
	if interactive {
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	logger.Debug().Str("container", containerName).Bool("interactive", interactive).Bool("asUser", asUser).Strs("args", args).Msg("executing command in container")
	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}
