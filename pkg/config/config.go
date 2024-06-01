package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

var (
	ErrPathIsNotFile     = errors.New("path is not a file")
	ErrNoConfigFileFound = errors.New("no config file found")
)

type PathsProperties struct {
	Data       string `yaml:"data"`
	Archive    string `yaml:"archive"`
	Repository string `yaml:"repository"`
	Temp       string `yaml:"temp"`
}

type GitHubProperties struct {
	Token string `yaml:"token"`
}

type ConfigProperties struct {
	Paths  PathsProperties  `yaml:"paths"`
	GitHub GitHubProperties `yaml:"github"`
}

func LoadConfig() (*ConfigProperties, error) {
	file, err := LookupConfigFile("gitar")
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}
	cfg := new(ConfigProperties)
	err = yaml.Unmarshal(data, cfg)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

func ReadConfig(filename string) (*ConfigProperties, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	cfg := new(ConfigProperties)
	err = yaml.Unmarshal(data, cfg)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

func LookupConfigFile(name string) (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	paths := []string{
		fmt.Sprintf("%s.yaml", name),
		fmt.Sprintf("%s.yml", name),
		"config.yaml",
		"config.yml",
		filepath.Join(homeDir, fmt.Sprintf(".%s/config.yaml", name)),
		filepath.Join(homeDir, fmt.Sprintf(".%s.yaml", name)),
		filepath.Join(homeDir, fmt.Sprintf(".%s/config.yml", name)),
		filepath.Join(homeDir, fmt.Sprintf(".%s.yml", name)),
		fmt.Sprintf("/etc/%s/config.yaml", name),
		fmt.Sprintf("/etc/%s.yaml", name),
		fmt.Sprintf("/etc/%s/config.yml", name),
		fmt.Sprintf("/etc/%s.yml", name),
	}
	for _, path := range paths {
		exists, _ := fileExists(path)
		if exists {
			fullpath, err := filepath.Abs(path)
			if err != nil {
				return "", err
			}
			return fullpath, nil
		}
	}
	return "", ErrNoConfigFileFound
}

func fileExists(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	if info.IsDir() {
		return false, ErrPathIsNotFile
	}
	return true, nil
}
