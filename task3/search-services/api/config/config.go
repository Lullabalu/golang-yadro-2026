package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"log"
	"time"
)

type Config struct {
	LogLevel     string     `yaml:"log_level" env:"LOG_LEVEL"`
	WordsAddress string     `yaml:"words_address" env:"WORDS_ADDRESS"`
	HTTPServer   HttpServer `yaml:"http_server"`
}

type HttpServer struct {
	Address string        `yaml:"address" env:"HTTP_SERVER_ADDRESS"`
	Timeout time.Duration `yaml:"timeout" env:"HTTP_SERVER_TIMEOUT"`
}

func MustLoad(configPath string) Config {
	var cfg Config
	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("cannot read config %q: %s", configPath, err)
	}
	return cfg
}
