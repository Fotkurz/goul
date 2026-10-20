package config

import (
	"os"
	"strings"
)

type Config struct {
	Port        string
	DBPath      string
	OllamaURL   string
	OllamaModel string
}

const (
	DefaultPort        = ":8080"
	DefaultDBPath      = "goul.db"
	DefaultOllamaURL   = "http://localhost:11434"
	DefaultOllamaModel = "qwen2.5:0.5b"
)

func Load() *Config {
	port := os.Getenv("GOUL_PORT")
	if port == "" {
		port = os.Getenv("PORT")
	}
	if port == "" {
		port = DefaultPort
	} else if !strings.HasPrefix(port, ":") {
		port = ":" + port
	}

	dbPath := os.Getenv("GOUL_DB_PATH")
	if dbPath == "" {
		dbPath = DefaultDBPath
	}

	ollamaURL := os.Getenv("OLLAMA_URL")
	if ollamaURL == "" {
		ollamaURL = DefaultOllamaURL
	}
	ollamaURL = strings.TrimRight(ollamaURL, "/")

	ollamaModel := os.Getenv("OLLAMA_MODEL")
	if ollamaModel == "" {
		ollamaModel = DefaultOllamaModel
	}

	return &Config{
		Port:        port,
		DBPath:      dbPath,
		OllamaURL:   ollamaURL,
		OllamaModel: ollamaModel,
	}
}
