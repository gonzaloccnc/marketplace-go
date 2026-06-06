package env

import (
	"log/slog"
	"os"

	"github.com/joho/godotenv"
)

func init() {
	err := godotenv.Load()

	if err != nil {
		slog.Error("Error loading .env file", "error", err)
	}
}

func GetOrDefault(key, defaultValue string) string {
	value, exist := os.LookupEnv(key)

	if !exist {
		return defaultValue
	}

	return value
}
