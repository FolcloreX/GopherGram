package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	APIID            int
	APIHash          string
	Phone            string
	Password         string
	ChatID           int64
	Logo             string
	PostGroupID      int64
	PostGroupTopicID int
}

func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		return nil, fmt.Errorf("erro ao carregar arquivo .env: %w", err)
	}

	apiID, err := strconv.Atoi(os.Getenv("API_ID"))
	if err != nil {
		return nil, fmt.Errorf("API_ID inválido no .env: %w", err)
	}

	// Opcional
	var chatID int64
	if val := os.Getenv("ORIGIN_CHAT_ID"); val != "" {
		chatID, _ = strconv.ParseInt(val, 10, 64)
	}

	// Opcional
	var postGroupID int64
	if val := os.Getenv("POST_GROUP_ID"); val != "" {
		postGroupID, _ = strconv.ParseInt(val, 10, 64)
	}

	// Opcional
	var postTopicID int
	if val := os.Getenv("POST_GROUP_TOPIC_ID"); val != "" {
		postTopicID, _ = strconv.Atoi(val)
	}

	logo := os.Getenv("LOGO")

	cfg := &Config{
		APIID:            apiID,
		APIHash:          os.Getenv("API_HASH"),
		Phone:            os.Getenv("PHONE_NUMBER"),
		Password:         os.Getenv("PASSWORD"),
		ChatID:           chatID,
		Logo:             logo,
		PostGroupID:      postGroupID,
		PostGroupTopicID: postTopicID,
	}

	if cfg.APIHash == "" || cfg.Phone == "" {
		return nil, fmt.Errorf("API_HASH e PHONE_NUMBER devem ser definidos no .env")
	}

	return cfg, nil
}
