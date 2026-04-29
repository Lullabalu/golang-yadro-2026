package config

import (
	"log"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type HTTPConfig struct {
	Address string        `yaml:"address" env:"API_ADDRESS" env-default:"localhost:80"`
	Timeout time.Duration `yaml:"timeout" env:"API_TIMEOUT" env-default:"5s"`
}

type Config struct {
	LogLevel          string        `yaml:"log_level" env:"LOG_LEVEL" env-default:"DEBUG"`
	HTTPConfig        HTTPConfig    `yaml:"api_server"`
	WordsAddress      string        `yaml:"words_address" env:"WORDS_ADDRESS" env-default:"words:81"`
	UpdateAddress     string        `yaml:"update_address" env:"UPDATE_ADDRESS" env-default:"update:82"`
	SearchAddress     string        `yaml:"search_address" env:"SEARCH_ADDRESS" env-default:"search:83"`
	TokenTTL          time.Duration `env:"TOKEN_TTL" env-default:"2m"`
	SearchConcurrency int           `env:"SEARCH_CONCURRENCY" env-default:"10"`
	SearchRate        int           `env:"SEARCH_RATE" env-default:"100"`
}

func MustLoad(configPath string) Config {
	var cfg Config
	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		if err = cleanenv.ReadEnv(&cfg); err != nil {
			log.Fatalf("Cannot read env variables")
		}
	}
	return cfg
}
