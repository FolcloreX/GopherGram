package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	APIID    int
	APIHash  string
	Phone    string
	Password string
	ChatID   int64
}

func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		return nil, fmt.Errorf("erro ao carregar arquivo .env: %w", err)
	}

	apiID, err := strconv.Atoi(os.Getenv("API_ID"))
	if err != nil {
		return nil, fmt.Errorf("API_ID inválido no .env: %w", err)
	}

	ChatID, err := strconv.ParseInt(os.Getenv("CHAT_ID"), 10, 64)
	if err != nil {
		return nil, fmt.Errorf("CHAT_ID inválido: %w", err)
	}

	cfg := &Config{
		APIID:    apiID,
		APIHash:  os.Getenv("API_HASH"),
		Phone:    os.Getenv("PHONE_NUMBER"),
		Password: os.Getenv("PASSWORD"),
		ChatID:   ChatID,
	}

	if cfg.APIHash == "" || cfg.Phone == "" {
		return nil, fmt.Errorf("API_HASH e PHONE_NUMBER devem ser definidos no .env")
	}

	return cfg, nil
}
