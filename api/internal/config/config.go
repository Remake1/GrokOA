package config

import (
	"errors"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

const (
	defaultConfigPath = "config/local.yaml"
	defaultDotenvPath = ".env"
)

type Config struct {
	HTTP HTTP `yaml:"http"`
}

type HTTP struct {
	Host            string        `yaml:"host" env:"HTTP_HOST" env-default:"0.0.0.0"`
	Port            string        `yaml:"port" env:"HTTP_PORT" env-default:"8080"`
	ReadTimeout     time.Duration `yaml:"read_timeout" env:"HTTP_READ_TIMEOUT" env-default:"5s"`
	WriteTimeout    time.Duration `yaml:"write_timeout" env:"HTTP_WRITE_TIMEOUT" env-default:"10s"`
	IdleTimeout     time.Duration `yaml:"idle_timeout" env:"HTTP_IDLE_TIMEOUT" env-default:"60s"`
	ShutdownTimeout time.Duration `yaml:"shutdown_timeout" env:"HTTP_SHUTDOWN_TIMEOUT" env-default:"10s"`
}

func (h HTTP) Address() string {
	return net.JoinHostPort(h.Host, h.Port)
}

func Load() (Config, error) {
	dotenvPath := os.Getenv("DOTENV_PATH")
	if dotenvPath == "" {
		dotenvPath = defaultDotenvPath
	}

	if err := godotenv.Load(dotenvPath); err != nil && !errors.Is(err, os.ErrNotExist) {
		return Config{}, fmt.Errorf("load dotenv from %q: %w", dotenvPath, err)
	}

	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = defaultConfigPath
	}

	var cfg Config
	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		return Config{}, fmt.Errorf("read config from %q: %w", configPath, err)
	}

	return cfg, nil
}
