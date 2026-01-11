package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DBAddr     string
	DBName     string
	ServerPort string
	WebUser    string
	WebPass    string
}

var cfg *Config

func init() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	cfg = &Config{
		DBAddr:     getEnvFirst([]string{"DB_SOCKET"}, "localhost:9090"),
		DBName:     getEnvFirst([]string{"DB_NAME"}, "siem_events"),
		ServerPort: getEnvFirst([]string{"SERVER_PORT"}, "8080"),
		WebUser:    getEnvFirst([]string{"WEB_USER"}, "admin"),
		WebPass:    getEnvFirst([]string{"WEB_PASSWORD"}, "admin"),
	}
}

func GetConfig() *Config {
	return cfg
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvFirst(keys []string, defaultValue string) string {
	for _, key := range keys {
		if value := os.Getenv(key); value != "" {
			return value
		}
	}
	return defaultValue
}
