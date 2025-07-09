package config

import (
	"os"

	"github.com/joho/godotenv"
)

type ServerConfig struct {
	Port string
}

func LoadServerConfig() (*ServerConfig, error) {
	godotenv.Load()
	port := os.Getenv("PORT")
	return &ServerConfig{Port: port}, nil
}
