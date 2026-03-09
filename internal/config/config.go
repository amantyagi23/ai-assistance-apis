package config

import (
	"os"
	"time"
)

type Config struct {
	Port            string
	APIPrefix       string
	OllamaHostPath  string
	MongodbURI      string
	DatabaseName    string
	DatabaseTimeOut time.Duration
	FrontendURL     string
	JWTSecret       string
	TokenHeader     string
}

func Load() *Config {
	return &Config{
		Port:            os.Getenv("PORT"),
		APIPrefix:       os.Getenv("API_PREFIX"),
		OllamaHostPath:  os.Getenv("OLLAMA_HOST_PATH"),
		MongodbURI:      os.Getenv("MONGO_URI"),
		DatabaseName:    os.Getenv("DATABASE_NAME"),
		DatabaseTimeOut: 10 * time.Second,
		FrontendURL:     os.Getenv("FRONTEND_URL"),
		JWTSecret:       os.Getenv("JWT_SECRET"),
	}
}
