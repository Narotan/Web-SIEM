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
		DBAddr:     getEnvFirst([]string{"NOSQL_DB_ADDR", "DB_SOCKET"}, "localhost:9090"),
		DBName:     getEnvFirst([]string{"NOSQL_DB_NAME", "DB_NAME"}, "siem_events"),
		ServerPort: getEnvFirst([]string{"SERVER_PORT"}, "8080"),
		WebUser:    getEnvFirst([]string{"WEB_USER", "USER"}, "admin"),
		WebPass:    getEnvFirst([]string{"WEB_PASSWORD", "PASSWORD"}, "admin"),
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

// getEnvFirst returns the first non-empty env value from the provided keys.
func getEnvFirst(keys []string, defaultValue string) string {
	for _, key := range keys {
		if value := os.Getenv(key); value != "" {
			return value
		}
	}
	return defaultValue
}
