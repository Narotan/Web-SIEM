package config

import (
	"log"
	"sync"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	ServerPort string `env:"DB_PORT" env-default:"8080"`
	WebUser    string `env:"USER" env-required:"true"`
	WebPass    string `env:"PASSWORD" env-required:"true"`
	DBAddr     string `env:"DB_SOCKET" env-default:"localhost:5140"`
	DBName     string `env:"DB_NAME" env-default:"security_events"`
}

var (
	cfg  *Config
	once sync.Once
)

// GetConfig загружает настройки один раз за время работы приложения
func GetConfig() *Config {
	once.Do(func() {
		cfg = &Config{}
		// Читаем .env файл и заполняем структуру
		if err := cleanenv.ReadConfig(".env", cfg); err != nil {
			// Если .env нет, пытаемся прочитать просто переменные окружения
			if err := cleanenv.ReadEnv(cfg); err != nil {
				log.Fatalf("failed to load configuration: %v", err)
			}
		}
	})
	return cfg
}
