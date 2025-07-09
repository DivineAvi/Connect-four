package config

import (
	"os"

	"github.com/joho/godotenv"
)

type DBConfig struct {
	DatabaseURL string
}

func LoadDBConfig() (*DBConfig, error) {
	godotenv.Load()

	return &DBConfig{
		DatabaseURL: os.Getenv("DATABASE_URL"),
	}, nil
}
