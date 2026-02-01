package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

const DefaultHTTPPort = ":8080"

type Config struct {
	HTTPPort string
}

func NewConfig() (Config, error) {
	cfg := Config{}

	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found")
	}

	httpPort := os.Getenv("HTTP_PORT")
	if httpPort == "" {
		cfg.HTTPPort = DefaultHTTPPort
	} else {
		// Ensure port starts with ':' if not already present
		if len(httpPort) > 0 && httpPort[0] != ':' {
			cfg.HTTPPort = ":" + httpPort
		} else {
			cfg.HTTPPort = httpPort
		}
	}
	return cfg, nil
}
