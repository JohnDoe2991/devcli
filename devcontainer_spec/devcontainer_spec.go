package devcontainerspec

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	NamePrefix string = "devcli"
)

type RegistryAlias struct {
	Original string `json:"original"`
	Alias    string `json:"alias"`
}

// DevcontainerConfig represents the key fields from a devcontainer.json file
type DevcontainerConfig struct {
	Name               string
	DockerFileContent  string
	Context            string
	Image              string
	Mounts             []string
	RunArgs            []string
	PostStartCommands  []string
	PostCreateCommands []string
	RegistryAliases    []RegistryAlias
}

type DevcontainerJson struct {
	Name       string `json:"name,omitempty"`
	DockerFile string `json:"dockerFile,omitempty"`
	Build      struct {
		Dockerfile string `json:"dockerfile,omitempty"`
		Context    string `json:"context,omitempty"`
	} `json:"build"`
	Image             string   `json:"image,omitempty"`
	Mounts            []string `json:"mounts,omitempty"`
	RunArgs           []string `json:"runArgs,omitempty"`
	PostStartCommand  string   `json:"postStartCommand,omitempty"`
	PostCreateCommand string   `json:"postCreateCommand,omitempty"`
	Customizations    struct {
		Devcli struct {
			RegistryAliases []RegistryAlias `json:"registryAliases"`
		} `json:"devcli"`
	} `json:"customizations"`
}

type Devcontainer struct {
	Cwd    string
	Config DevcontainerConfig
	Hash   string
}

func ParseDevcontainer(path string) (Devcontainer, error) {
	devc := Devcontainer{}
	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		return Devcontainer{}, err
	}
	homeConfigDir := filepath.Join(userConfigDir, "devcli")
	homeConfigDevcontainer := filepath.Join(homeConfigDir, ".devcontainer", "devcontainer.json")
	if _, err := os.Stat(homeConfigDevcontainer); err == nil {
		logger.Debug().Msgf("Parsing devcontainer json from path: %s", homeConfigDir)
		globalData, err := parseDevcontainerJson(homeConfigDir)
		if err != nil {
			return Devcontainer{}, err
		}
		logger.Debug().Msgf("Parsed devcontainer config: %+v", globalData)
		devc.Merge(globalData)
	}
	devc.Cwd = path
	logger.Debug().Msgf("Parsing devcontainer json from path: %s", path)
	projectData, err := parseDevcontainerJson(path)
	if err != nil {
		return Devcontainer{}, err
	}
	logger.Debug().Msgf("Parsed devcontainer config: %+v", projectData)
	devc.Merge(projectData)
	hash, err := calculateDevcontainerHash(devc)
	if err != nil {
		return Devcontainer{}, err
	}
	logger.Debug().Msgf("Calculated devcontainer hash: %s", hash)
	devc.Hash = hash
	logger.Debug().Msgf("Final devcontainer: %+v", devc)
	return devc, nil
}

// parseDevcontainerConfig reads a devcontainer.json file from the specified path,
// cleans JSON5 features like comments, applies regex replacements,
// and extracts the key configuration elements
func parseDevcontainerJson(path string) (DevcontainerJson, error) {
	devPath := filepath.Join(path, ".devcontainer", "devcontainer.json")
	data, err := os.ReadFile(devPath)
	if err != nil {
		return DevcontainerJson{}, err
	}

	// Convert to string for preprocessing
	jsonStr := string(data)

	// Remove single line comments (//...); either directly at the start of the line, or a whitespace character followed by "//"
	re := regexp.MustCompile(`(?m)((^)|([ \t]+))//.*`)
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
		envValue, exists := os.LookupEnv(envVar)
		if !exists {
			logger.Warn().Str("env", envVar).Msg("environment variable is not set")
		}
		logger.Debug().Str("env", envVar).Str("value", envValue).Msg("set env variable")
		jsonStr = strings.ReplaceAll(jsonStr, match[0], envValue)
	}

	// replace workplace reference
	re = regexp.MustCompile(`\${localWorkspaceFolder}`)
	jsonStr = re.ReplaceAllString(jsonStr, path)
	re = regexp.MustCompile(`\${localWorkspaceFolderBasename}`)
	localWorkspaceFolderBasename := filepath.Base(path)
	jsonStr = re.ReplaceAllString(jsonStr, localWorkspaceFolderBasename)
	re = regexp.MustCompile(`\${containerWorkspaceFolder}`)
	containerWorkspaceFolder := filepath.Join("/workspaces", localWorkspaceFolderBasename)
	jsonStr = re.ReplaceAllString(jsonStr, containerWorkspaceFolder)

	// Parse the cleaned JSON
	var jsonData DevcontainerJson
	if err := json.Unmarshal([]byte(jsonStr), &jsonData); err != nil {
		return DevcontainerJson{}, fmt.Errorf("failed to unmarshal devcontainer config: %w, %s", err, jsonStr)
	}

	return jsonData, nil
}

// CalculateDevcontainerHash generates a unique hash based on:
// - The current working directory
// - The entire DevcontainerConfig content
// - The content of the Dockerfile if specified in the config
// This hash can be used to identify unique development environment configurations
func calculateDevcontainerHash(devc Devcontainer) (string, error) {
	// Create a hash builder
	h := sha256.New()

	// Add current working directory to hash
	io.WriteString(h, devc.Cwd)

	// Add config content to hash
	configJSON, err := json.Marshal(devc.Config)
	if err != nil {
		return "", fmt.Errorf("failed to marshal config: %w", err)
	}
	h.Write(configJSON)

	// If Dockerfile is specified, add its content to hash
	if devc.Config.DockerFileContent != "" {
		h.Write([]byte(devc.Config.DockerFileContent))
	}

	// Generate the final hash
	hashBytes := h.Sum(nil)
	return hex.EncodeToString(hashBytes), nil
}

func (devc Devcontainer) GetImageName() string {
	if devc.Config.Image != "" {
		return devc.Config.Image
	} else if devc.Config.DockerFileContent != "" {
		return devc.GetDevcNamePrefix() + devc.Hash[0:7]
	} else {
		return ""
	}
}

func (devc Devcontainer) GetContainerName() string {
	return devc.GetDevcNamePrefix() + devc.Hash[0:7]
}

func (devc Devcontainer) GetDevcNamePrefix() string {
	return NamePrefix + "_" + strings.ToLower(filepath.Base(devc.Cwd)) + "_"
}

// Merge a DevcontainerJson into the devcontainer config.
// Single arguments get overwritten, arrays get appended.
func (devc *Devcontainer) Merge(devj DevcontainerJson) error {
	devc.Config.Name = devj.Name
	// the dockerfile can be defined either with "dockerFile" directly or with "build.dockerfile"
	// we check both and give the "build.dockerfile" priority
	if devj.DockerFile != "" {
		content, err := os.ReadFile(path.Join(devc.Cwd, ".devcontainer", devj.DockerFile))
		if err != nil {
			return err
		}
		devc.Config.DockerFileContent = string(content)
	}
	if devj.Build.Dockerfile != "" {
		content, err := os.ReadFile(path.Join(devc.Cwd, ".devcontainer", devj.Build.Dockerfile))
		if err != nil {
			return err
		}
		devc.Config.DockerFileContent = string(content)
	}
	devc.Config.Context = devj.Build.Context
	devc.Config.Image = devj.Image
	devc.Config.Mounts = append(devc.Config.Mounts, devj.Mounts...)
	devc.Config.RunArgs = append(devc.Config.RunArgs, devj.RunArgs...)
	devc.Config.PostStartCommands = append(devc.Config.PostStartCommands, devj.PostStartCommand)
	devc.Config.PostCreateCommands = append(devc.Config.PostCreateCommands, devj.PostCreateCommand)
	devc.Config.RegistryAliases = append(devc.Config.RegistryAliases, devj.Customizations.Devcli.RegistryAliases...)
	devc.ApplyRegistryAliases()
	return nil
}

func (devc *Devcontainer) ApplyRegistryAliases() {
	for _, alias := range devc.Config.RegistryAliases {
		logger.Debug().Str("original", alias.Original).Str("alias", alias.Alias).Msg("apply registry alias")
		devc.Config.Image = strings.Replace(devc.Config.Image, alias.Original, alias.Alias, -1)
		devc.Config.DockerFileContent = strings.Replace(devc.Config.DockerFileContent, alias.Original, alias.Alias, -1)
	}
	logger.Debug().Msgf("new Image: %s", devc.Config.Image)
	logger.Debug().Str("DockerfileContent", devc.Config.DockerFileContent).Msg("new Dockerfile")
}
