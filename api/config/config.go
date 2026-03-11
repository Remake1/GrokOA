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
	defaultConfigPath = "config/config.yaml"
	defaultDotenvPath = ".env"
	defaultTokenTTL   = 4 * time.Hour
)

type Config struct {
	HTTP       HTTP       `yaml:"http"`
	Logging    Logging    `yaml:"logging"`
	Auth       Auth       `yaml:"auth"`
	Screenshot Screenshot `yaml:"screenshot"`
	Room       Room       `yaml:"room"`
	AI         AI         `yaml:"ai"`
}

type HTTP struct {
	Host            string        `yaml:"host" env:"HTTP_HOST" env-default:"0.0.0.0"`
	Port            string        `yaml:"port" env:"HTTP_PORT" env-default:"8080"`
	ReadTimeout     time.Duration `yaml:"read_timeout" env:"HTTP_READ_TIMEOUT" env-default:"5s"`
	WriteTimeout    time.Duration `yaml:"write_timeout" env:"HTTP_WRITE_TIMEOUT" env-default:"10s"`
	IdleTimeout     time.Duration `yaml:"idle_timeout" env:"HTTP_IDLE_TIMEOUT" env-default:"60s"`
	ShutdownTimeout time.Duration `yaml:"shutdown_timeout" env:"HTTP_SHUTDOWN_TIMEOUT" env-default:"10s"`
}

type Logging struct {
	Level           string `yaml:"level" env:"LOG_LEVEL" env-default:"info"`
	Format          string `yaml:"format" env:"LOG_FORMAT" env-default:"console"`
	TimeFieldFormat string `yaml:"time_field_format" env:"LOG_TIME_FIELD_FORMAT" env-default:"2006-01-02T15:04:05Z07:00"`
	IncludeCaller   bool   `yaml:"include_caller" env:"LOG_INCLUDE_CALLER" env-default:"false"`
}

type Auth struct {
	AccessKey string        `yaml:"-" env:"ACCESS_KEY" env-required:"true"`
	JWTSecret string        `yaml:"-" env:"JWT_SECRET" env-required:"true"`
	TokenTTL  time.Duration `yaml:"token_ttl"`
	RateLimit AuthRateLimit `yaml:"rate_limit"`
}

type AuthRateLimit struct {
	MaxFailedAttempts int           `yaml:"max_failed_attempts" env:"AUTH_RATE_LIMIT_MAX_FAILED_ATTEMPTS" env-default:"5"`
	Window            time.Duration `yaml:"window" env:"AUTH_RATE_LIMIT_WINDOW" env-default:"1h"`
}

type Screenshot struct {
	Dir string `yaml:"dir" env:"SCREENSHOT_DIR" env-default:"./screenshots"`
}

type Room struct {
	GracePeriod time.Duration `yaml:"grace_period" env:"ROOM_GRACE_PERIOD" env-default:"30s"`
}

type AI struct {
	OpenAI OpenAI `yaml:"openai"`
	Gemini Gemini `yaml:"gemini"`
}

type OpenAI struct {
	APIKey string   `yaml:"-" env:"OPENAI_API_KEY"`
	Models []string `yaml:"models"`
}

type Gemini struct {
	APIKey string   `yaml:"-" env:"GEMINI_API_KEY"`
	Models []string `yaml:"models"`
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

	if cfg.Auth.TokenTTL <= 0 {
		cfg.Auth.TokenTTL = defaultTokenTTL
	}

	if cfg.Auth.RateLimit.MaxFailedAttempts <= 0 {
		cfg.Auth.RateLimit.MaxFailedAttempts = 5
	}

	if cfg.Auth.RateLimit.Window <= 0 {
		cfg.Auth.RateLimit.Window = time.Hour
	}

	return cfg, nil
}
