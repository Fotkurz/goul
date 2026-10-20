package config

import (
	"os"
	"testing"
)

func TestLoadDefaults(t *testing.T) {
	os.Unsetenv("GOUL_PORT")
	os.Unsetenv("PORT")
	os.Unsetenv("GOUL_DB_PATH")
	os.Unsetenv("OLLAMA_URL")
	os.Unsetenv("OLLAMA_MODEL")

	cfg := Load()

	if cfg.Port != DefaultPort {
		t.Errorf("expected Port %s, got %s", DefaultPort, cfg.Port)
	}
	if cfg.DBPath != DefaultDBPath {
		t.Errorf("expected DBPath %s, got %s", DefaultDBPath, cfg.DBPath)
	}
	if cfg.OllamaURL != DefaultOllamaURL {
		t.Errorf("expected OllamaURL %s, got %s", DefaultOllamaURL, cfg.OllamaURL)
	}
	if cfg.OllamaModel != DefaultOllamaModel {
		t.Errorf("expected OllamaModel %s, got %s", DefaultOllamaModel, cfg.OllamaModel)
	}
}

func TestLoadFromEnv(t *testing.T) {
	t.Setenv("GOUL_PORT", "9090")
	t.Setenv("GOUL_DB_PATH", "/tmp/custom.db")
	t.Setenv("OLLAMA_URL", "http://homelab:11434/")
	t.Setenv("OLLAMA_MODEL", "llama3.2:1b")

	cfg := Load()

	if cfg.Port != ":9090" {
		t.Errorf("expected Port :9090, got %s", cfg.Port)
	}
	if cfg.DBPath != "/tmp/custom.db" {
		t.Errorf("expected DBPath /tmp/custom.db, got %s", cfg.DBPath)
	}
	if cfg.OllamaURL != "http://homelab:11434" {
		t.Errorf("expected OllamaURL without trailing slash, got %s", cfg.OllamaURL)
	}
	if cfg.OllamaModel != "llama3.2:1b" {
		t.Errorf("expected OllamaModel llama3.2:1b, got %s", cfg.OllamaModel)
	}
}
