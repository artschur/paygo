package config

import "os"

type Config struct {
	DatabaseURL string
	Port        string
}

func LoadConfig() Config {
	DB_URL := os.Getenv("DATABASE_URL")
	if DB_URL == "" {
		return Config{}
	}

	APP_PORT := os.Getenv("APP_PORT")
	if APP_PORT == "" {
		return Config{}
	}
	return Config{
		DatabaseURL: DB_URL,
		Port:        APP_PORT,
	}
}
