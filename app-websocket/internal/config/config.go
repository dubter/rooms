package config

import (
	"flag"
	"fmt"
	"github.com/joho/godotenv"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env      string `yaml:"env" env-default:"local"`
	Auth     AuthConfig
	Http     HTTPConfig
	Chat     ChatConfig
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

type AuthConfig struct {
	AccessTokenTTL  time.Duration `yaml:"access_token_ttl" env-default:"15m"`
	RefreshTokenTTL time.Duration `yaml:"refresh_token_ttl" env-default:"720h"`
	PasswordSalt    string        `env:"PASSWORD_SALT" env-required:"true"`
	JWTSigningKey   string        `env:"JWT_SIGNING_KEY" env-required:"true"`
}

type HTTPConfig struct {
	Port            string        `yaml:"port" env-default:"8080"`
	ReadTimeout     time.Duration `yaml:"read_timeout" env-default:"10s"`
	WriteTimeout    time.Duration `yaml:"write_timeout" env-default:"10s"`
	ShutdownTimeout time.Duration `yaml:"shutdown_timeout" env-default:"10s"`
	Limiter         Limiter
	TLS             TLSConfig `yaml:"tls"`
}

type TLSConfig struct {
	CertFilePath string `yaml:"cert"`
	KeyFilePath  string `yaml:"key"`
}

type ChatConfig struct {
	CountMessagesGet int `yaml:"count_messages_get" env-default:"10"`
}

type Limiter struct {
	RPS   int           `yaml:"rps" env-default:"10"`
	Burst int           `yaml:"burst" env-default:"20"`
	TTL   time.Duration `yaml:"ttl" env-default:"10m"`
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
