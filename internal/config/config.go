package config

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the application.
type Config struct {
	ServerPort    string
	AuthUser      string
	AuthPass      string
	DatabaseURL   string
	GithubToken   string
	SyncInterval  time.Duration
}

// LoadConfig loads configuration from environment variables.
// It attempts to load a .env file first for local development.
func LoadConfig() (*Config, error) {
	// Attempt to load .env file, ignore error if it doesn't exist
	_ = godotenv.Load() 

	port := getEnv("SERVER_PORT", "8080")
	authUser := getEnv("AUTH_USER", "") // Require AUTH_USER
	authPass := getEnv("AUTH_PASS", "") // Require AUTH_PASS
	dbURL := getEnv("DATABASE_URL", "") // Require DATABASE_URL
	githubToken := getEnv("GITHUB_TOKEN", "") // Require GITHUB_TOKEN
	syncIntervalStr := getEnv("SYNC_INTERVAL", "1h") // Default to 1 hour

	if authUser == "" || authPass == "" {
		log.Fatal("AUTH_USER and AUTH_PASS environment variables are required")
	}
	if dbURL == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}
	if githubToken == "" {
		log.Fatal("GITHUB_TOKEN environment variable is required")
	}

	syncInterval, err := time.ParseDuration(syncIntervalStr)
	if err != nil {
		log.Printf("Warning: Invalid SYNC_INTERVAL format '%s'. Using default 1h. Error: %v", syncIntervalStr, err)
		syncInterval = time.Hour // Default to 1 hour on parse error
	}

	cfg := &Config{
		ServerPort:    port,
		AuthUser:      authUser,
		AuthPass:      authPass,
		DatabaseURL:   dbURL,
		GithubToken:   githubToken,
		SyncInterval:  syncInterval,
	}

	log.Println("Configuration loaded successfully.")
	// Avoid logging sensitive info like tokens or passwords in production
	log.Printf("Server Port: %s", cfg.ServerPort)
	log.Printf("Sync Interval: %s", cfg.SyncInterval.String())

	return cfg, nil
}

// getEnv retrieves an environment variable or returns a default value.
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

// getEnvAsInt retrieves an environment variable as an integer or returns a default value.
// Kept for potential future use, though not currently needed for this config.
func getEnvAsInt(key string, fallback int) int {
	valueStr := getEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return fallback
}
