package config

import (
	"log"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL      string
	RedisURL         string
	TelegramBotToken string
	ServerPort       string
}

func Load() *Config {
	// Coba load file .env dari directory saat ini atau dari parent directory (root project)
	if err := godotenv.Load(); err != nil {
		if err := godotenv.Load(filepath.Join("..", ".env")); err != nil {
			godotenv.Load(filepath.Join("..", "..", ".env"))
		}
	}

	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	dbURL := os.Getenv("DATABASE_URL")
	redisURL := os.Getenv("REDIS_URL")

	if dbURL == "" {
		log.Println("Notice: DATABASE_URL not set in environment or .env")
	}

	return &Config{
		DatabaseURL:      dbURL,
		RedisURL:         redisURL,
		TelegramBotToken: os.Getenv("TELEGRAM_BOT_TOKEN"),
		ServerPort:       port,
	}
}
