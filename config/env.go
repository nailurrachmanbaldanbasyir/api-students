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

func GetEnvWithDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func GetEnvInt(key string, fallback int) int {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}
	var n int
	_, err := parseInt(val, &n)
	if err != nil {
		return fallback
	}
	return n
}

func parseInt(s string, out *int) (int, error) {
	var n int
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, os.ErrInvalid
		}
		n = n*10 + int(c-'0')
	}
	*out = n
	return n, nil
}
