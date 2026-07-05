package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

var discoveryFiles = []string{".release-clerk.yaml", ".release-clerk.yml"}

func Discover() (Config, error) {
	for _, file := range discoveryFiles {
		if _, err := os.Stat(file); err == nil {
			return Load(file)
		}
	}

	return Config{}, fmt.Errorf("no config file found in current directory")
}

func Load(path string) (Config, error) {
	f, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}

	var c Config

	if err := yaml.Unmarshal(f, &c); err != nil {
		return Config{}, err
	}

	return c, nil
}
