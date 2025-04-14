package devcontainerspec

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
)

// DevcontainerConfig represents the key fields from a devcontainer.json file
type DevcontainerConfig struct {
	Name              string   `json:"name,omitempty"`
	DockerFile        string   `json:"dockerFile,omitempty"`
	Image             string   `json:"image,omitempty"`
	Mounts            []string `json:"mounts,omitempty"`
	RunArgs           []string `json:"runArgs,omitempty"`
	PostStartCommand  string   `json:"postStartCommand,omitempty"`
	PostCreateCommand string   `json:"postCreateCommand,omitempty"`
}

type Devcontainer struct {
	Cwd    string
	Config DevcontainerConfig
	Hash   string
}

func ParseDevcontainer(path string) (Devcontainer, error) {
	devcontainer := Devcontainer{}
	devcontainer.Cwd = path
	logger.Debug().Msgf("Parsing devcontainer config from path: %s", path)
	config, err := parseDevcontainerConfig(path)
	if err != nil {
		return Devcontainer{}, err
	}
	logger.Debug().Msgf("Parsed devcontainer config: %+v", config)
	devcontainer.Config = config
	hash, err := calculateDevcontainerHash(path)
	if err != nil {
		return Devcontainer{}, err
	}
	logger.Debug().Msgf("Calculated devcontainer hash: %s", hash)
	devcontainer.Hash = hash
	return devcontainer, nil
}

// parseDevcontainerConfig reads a devcontainer.json file from the specified path,
// cleans JSON5 features like comments, applies regex replacements,
// and extracts the key configuration elements
func parseDevcontainerConfig(path string) (DevcontainerConfig, error) {
	devPath := filepath.Join(path, ".devcontainer", "devcontainer.json")
	data, err := os.ReadFile(devPath)
	if err != nil {
		return DevcontainerConfig{}, err
	}

	// Convert to string for preprocessing
	jsonStr := string(data)

	// Remove single line comments (//...)
	re := regexp.MustCompile(`//.*`)
	jsonStr = re.ReplaceAllString(jsonStr, "")

	// Remove multi-line comments (/* ... */)
	re = regexp.MustCompile(`/\*[\s\S]*?\*/`)
	jsonStr = re.ReplaceAllString(jsonStr, "")

	// Remove trailing commas in objects and arrays
	re = regexp.MustCompile(`,\s*}`)
	jsonStr = re.ReplaceAllString(jsonStr, "}")
	re = regexp.MustCompile(`,\s*]`)
	jsonStr = re.ReplaceAllString(jsonStr, "]")

	// Replace environment variable references with their values
	re = regexp.MustCompile(`\${localEnv:(.+?)}`)
	matches := re.FindAllStringSubmatch(jsonStr, -1)
	for _, match := range matches {
		envVar := match[1]
		envValue := os.Getenv(envVar)
		if envValue == "" {
			return DevcontainerConfig{}, fmt.Errorf("environment variable %s not set", envVar)
		}
		jsonStr = re.ReplaceAllString(jsonStr, envValue)
	}

	// replace workplace reference
	re = regexp.MustCompile(`\${localWorkspaceFolder}`)
	jsonStr = re.ReplaceAllString(jsonStr, path)
	re = regexp.MustCompile(`\${containerWorkspaceFolder}`)
	containerWorkspaceFolder := filepath.Join("/workspaces", filepath.Base(path))
	jsonStr = re.ReplaceAllString(jsonStr, containerWorkspaceFolder)

	// Parse the cleaned JSON
	var config DevcontainerConfig
	if err := json.Unmarshal([]byte(jsonStr), &config); err != nil {
		return DevcontainerConfig{}, fmt.Errorf("failed to unmarshal devcontainer config: %w", err)
	}

	return config, nil
}

// CalculateDevcontainerHash generates a unique hash based on:
// - The current working directory
// - The entire DevcontainerConfig content
// - The content of the Dockerfile if specified in the config
// This hash can be used to identify unique development environment configurations
func calculateDevcontainerHash(path string) (string, error) {
	// Parse devcontainer config
	config, err := parseDevcontainerConfig(path)
	if err != nil {
		return "", fmt.Errorf("failed to parse devcontainer config: %w", err)
	}

	// Create a hash builder
	h := sha256.New()

	// Add current working directory to hash
	io.WriteString(h, path)

	// Add config content to hash
	configJSON, err := json.Marshal(config)
	if err != nil {
		return "", fmt.Errorf("failed to marshal config: %w", err)
	}
	h.Write(configJSON)

	// If Dockerfile is specified, add its content to hash
	if config.DockerFile != "" {
		// Resolve Dockerfile path relative to .devcontainer directory
		dockerfilePath := config.DockerFile
		if !filepath.IsAbs(dockerfilePath) {
			dockerfilePath = filepath.Join(path, ".devcontainer", dockerfilePath)
		}

		// Read Dockerfile content
		dockerfileContent, err := os.ReadFile(dockerfilePath)
		if err != nil {
			return "", fmt.Errorf("failed to read Dockerfile: %w", err)
		}
		h.Write(dockerfileContent)
	}

	// Generate the final hash
	hashBytes := h.Sum(nil)
	return hex.EncodeToString(hashBytes), nil
}
