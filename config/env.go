package config

import (
	"os"

	"github.com/joho/godotenv"
)

func LoadEnv() error {
	return godotenv.Load()
}

func GetEnv(key string) string {
	return os.Getenv(key)
}
