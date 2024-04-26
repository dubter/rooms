package config

import (
	"errors"
	"flag"
	"fmt"
	"github.com/ilyakaznacheev/cleanenv"
	"os"
)

type Config struct {
	Env     string `yaml:"env" env-default:"local"`
	Address string `yaml:"address" env-default:"localhost"`
	Port    string `yaml:"port" env-default:"80"`
}

func LoadConfig() (*Config, error) {
	configPath := fetchConfigPath()
	if configPath == "" {
		return nil, errors.New("config path is empty")
	}

	return LoadPath(configPath)
}

func LoadPath(configPath string) (*Config, error) {
	// check if file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("config file does not exist: " + configPath)
	}

	var cfg Config
	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		return nil, fmt.Errorf("cannot read config: %w", err)
	}

	return &cfg, nil
}

// fetchConfigPath fetches config path from command line flag or environment variable.
// Priority: flag > env > default.
// Default value is empty string.
func fetchConfigPath() string {
	var res string

	flag.StringVar(&res, "config", "", "path to config file")
	flag.Parse()

	if res == "" {
		res = os.Getenv("CONFIG_PATH")
	}

	return res
}
