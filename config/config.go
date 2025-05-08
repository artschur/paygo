package config

import (
	"log"
	"os"
)

type Config struct {
	DatabaseURL string
	Port        string
}

func LoadConfig() Config {
	DB_URL := os.Getenv("DATABASE_URL")

	APP_PORT := os.Getenv("APP_PORT")

	if DB_URL == "" || APP_PORT == "" {
		log.Println("Missing required environment variables (DATABASE_URL, APP_PORT)")
		log.Println("Shutting down server...")
		os.Exit(1) // Exit the program with error code
	}

	return Config{
		DatabaseURL: DB_URL,
		Port:        APP_PORT,
	}
}
