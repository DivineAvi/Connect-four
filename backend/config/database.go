package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// DBConfig holds database connection parameters
type DBConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// LoadDBConfig loads database configuration from environment variables
func LoadDBConfig() (*DBConfig, error) {
	// Load .env file if it exists
	godotenv.Load()

	// Get database connection parameters from environment variables
	host := getEnv("DB_HOST", "localhost")
	portStr := getEnv("DB_PORT", "5432")
	user := getEnv("DB_USER", "postgres")
	password := getEnv("DB_PASSWORD", "")
	dbName := getEnv("DB_NAME", "connect4")
	sslMode := getEnv("DB_SSL_MODE", "disable")

	// Parse port number
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, fmt.Errorf("invalid DB_PORT: %v", err)
	}

	return &DBConfig{
		Host:     host,
		Port:     port,
		User:     user,
		Password: password,
		DBName:   dbName,
		SSLMode:  sslMode,
	}, nil
}

// ConnectionString returns a formatted PostgreSQL connection string
func (c *DBConfig) ConnectionString() string {
	// Check if DATABASE_URL environment variable is set
	if url := os.Getenv("DATABASE_URL"); url != "" {
		return url
	}

	// Otherwise, construct the connection string from individual parameters
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		c.User, c.Password, c.Host, c.Port, c.DBName, c.SSLMode)
}

// getEnv retrieves an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
