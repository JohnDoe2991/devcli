package docker

import (
	"os/exec"
	"strings"

	devcontainerspec "github.com/johndoe2991/devcli/devcontainer_spec"
)

// Delete a single Image.
func CleanImage(imageName string) error {
	exists, err := checkImageExists(imageName)
	if err != nil {
		return err
	}
	if !exists {
		// image does not exists, return
		return nil
	}
	logger.Debug().Str("imageName", imageName).Msg("delete image")
	cmd := exec.Command("docker", "image", "rm", imageName)
	err = cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

// Delete a single Container. If the container is running, it will be stopped.
func CleanContainer(containerName string) error {
	exists, err := checkContainerExists(containerName)
	if err != nil {
		return err
	}
	if !exists {
		// image does not exists, return
		return nil
	}
	isRunning, err := checkContainerRunning(containerName)
	if err != nil {
		return err
	}
	if isRunning {
		err = stopContainer(containerName)
		if err != nil {
			return err
		}
	}
	logger.Debug().Str("imageName", containerName).Msg("delete container")
	cmd := exec.Command("docker", "container", "rm", containerName)
	err = cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

// Delete all images corresponding to the devcontainer config.
func CleanAllImageVersions(devc devcontainerspec.Devcontainer) error {
	logger.Debug().Str("suffix", devc.GetDevcNameSuffix()).Msg("clean all images with suffix")
	images, err := listImage()
	if err != nil {
		return err
	}
	baseName := devc.GetDevcNameSuffix()
	for _, image := range images {
		if strings.HasSuffix(image, baseName) {
			err := CleanImage(image)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// Delete all containers corresponding to the devcontainer config.
func CleanAllContainerVersions(devc devcontainerspec.Devcontainer) error {
	logger.Debug().Str("suffix", devc.GetDevcNameSuffix()).Msg("clean all containers with suffix")
	containers, err := listContainers()
	if err != nil {
		return err
	}
	baseName := devc.GetDevcNameSuffix()
	for _, container := range containers {
		if strings.HasSuffix(container, baseName) {
			err := CleanContainer(container)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// Delete all devcli images.
func CleanAllImages() error {
	logger.Debug().Msg("clean all images")
	images, err := listImage()
	if err != nil {
		return err
	}
	for _, image := range images {
		if image == "" {
			continue
		}
		err := CleanImage(image)
		if err != nil {
			return err
		}
	}
	return nil
}

// Delete all devcli containers.
func CleanAllContainers() error {
	logger.Debug().Msg("clean all containers")
	containers, err := listContainers()
	if err != nil {
		return err
	}
	for _, container := range containers {
		if container == "" {
			continue
		}
		err := CleanContainer(container)
		if err != nil {
			return err
		}
	}
	return nil
}
