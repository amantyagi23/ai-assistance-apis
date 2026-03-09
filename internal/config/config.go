package config

import "os"

type Config struct {
	Port           string
	APIPrefix      string
	OllamaHostPath string
}

func APPConfig() Config {
	return Config{
		Port:           os.Getenv("PORT"),
		APIPrefix:      os.Getenv("API_PREFIX"),
		OllamaHostPath: os.Getenv("OLLAMA_HOST_PATH"),
	}
}
