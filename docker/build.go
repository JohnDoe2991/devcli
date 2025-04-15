package docker

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	devcontainerspec "github.com/johndoe2991/devcli/devcontainer_spec"
)

func Build(devc devcontainerspec.Devcontainer) error {
	// first we check if we have to build or to pull a image
	if devc.Config.Image != "" {
		// pull the image
		logger.Debug().Str("image", devc.Config.Image).Msg("pulling image")
		if err := pullImage(devc.Config.Image); err != nil {
			return err
		}
	} else if devc.Config.DockerFile != "" {
		// the image has to be build, check if the image already exists
		imageName := devc.GetImageName()
		exists, err := checkImageExists(imageName)
		logger.Debug().Str("imageName", imageName).Bool("exists", exists).Msg("checking if image exists")
		if err != nil {
			return err
		}
		if !exists {
			// build the image
			logger.Debug().Str("dockerfile", devc.Config.DockerFile).Msg("building image")
			if err := buildImage(devc); err != nil {
				return err
			}
		}
	} else {
		return fmt.Errorf("no image or dockerfile specified")
	}
	return nil
}

func pullImage(imagepath string) error {
	// run docker and pull the image
	err := exec.Command("docker", "pull", imagepath).Run()
	if err != nil {
		return err
	}
	return nil
}

func buildImage(devc devcontainerspec.Devcontainer) error {
	// run docker and build the image
	dockerFilePath := filepath.Join("./.devcontainer", devc.Config.DockerFile)
	dockerFileBase := filepath.Dir(dockerFilePath)
	cmd := exec.Command("docker", "build", "-f", dockerFilePath, "-t", devc.GetImageName(), dockerFileBase)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

func checkImageExists(hash string) (bool, error) {
	// check if the image exists
	cmd := exec.Command("docker", "images", "-q", hash)
	output, err := cmd.Output()
	if err != nil {
		return false, err
	}
	return len(output) > 0, nil
}
