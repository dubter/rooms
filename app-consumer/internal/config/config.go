package config

import (
	"flag"
	"fmt"
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
	"os"
)

type Config struct {
	Env      string `yaml:"env" env-default:"local"`
	Postgres PostgresConfig
	Redis    RedisConfig
	Kafka    KafkaConfig
}

type PostgresConfig struct {
	PostgresURL string `env:"POSTGRES_URL" env-required:"true"`
}

type RedisConfig struct {
	Addrs    []string `yaml:"addrs" env-required:"true"`
	Password string   `env:"REDIS_PASSWORD" env-required:"true"`
}

type KafkaConfig struct {
	BrokerList    []string `yaml:"brokers" env-required:"true"`
	Topic         string   `yaml:"topic" env-required:"true"`
	ConsumerGroup string   `yaml:"consumer_group" env-required:"true"`
}

func LoadConfig() (*Config, error) {
	envPath, configPath := fetchPaths()

	if envPath == "" {
		return nil, fmt.Errorf("'.env' file path is empty")
	}

	if configPath == "" {
		return nil, fmt.Errorf("config path is empty")
	}

	if err := godotenv.Load(envPath); err != nil {
		return nil, fmt.Errorf("no .env file found")
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
		return nil, fmt.Errorf("can not read config: %w", err)
	}

	return &cfg, nil
}

// fetchConfigPath fetches config path from command line flag or environment variable.
// Priority: flag > env > default.
// Default value is empty string.
func fetchPaths() (string, string) {
	var envPath, configPath string

	flag.StringVar(&envPath, "env", "", "path to '.env' file")
	flag.StringVar(&configPath, "config", "", "path to config file")
	flag.Parse()

	if envPath == "" {
		envPath = os.Getenv("ENV_PATH")
	}

	if configPath == "" {
		configPath = os.Getenv("CONFIG_PATH")
	}

	return envPath, configPath
}
