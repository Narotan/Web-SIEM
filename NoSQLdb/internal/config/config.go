package config

import (
	"log"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Host string `env:"DB_HOST" env-default:""`
	Port string `env:"DB_PORT" env-default:"5140"`
}

func Load() *Config {
	var cfg Config

	if err := cleanenv.ReadConfig(".env", &cfg); err != nil {
		if err := cleanenv.ReadEnv(&cfg); err != nil {
			log.Fatalf("cannot read config: %s", err)
		}
	}

	return &cfg
}
