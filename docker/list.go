package docker

import (
	"os/exec"
	"strings"
)

func listImage() ([]string, error) {
	cmd := exec.Command("docker", "images", "--filter", "reference=devcli_*", "--format", "{{.Repository}}")
	output, err := cmd.Output()
	if err != nil {
		return []string{}, err
	}
	retval := strings.Split(string(output), "\n")
	return retval, nil
}

func listContainers() ([]string, error) {
	cmd := exec.Command("docker", "ps", "-a", "--filter", "name=devcli_*", "--format", "{{.Names}}")
	output, err := cmd.Output()
	if err != nil {
		return []string{}, err
	}
	retval := strings.Split(string(output), "\n")
	return retval, nil
}
