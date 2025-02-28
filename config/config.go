package config

import (
	"errors"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	WebhookURL string
}

var cfg *Config

func Init() error {
	if err := godotenv.Load(); err != nil {
		return err
	}

	cfg = &Config{
		WebhookURL: os.Getenv("WEBHOOK_URL"),
	}

	return validate()
}

func validate() error {
	if cfg.WebhookURL == "" {
		return errors.New("WEBHOOK_URL is not set in the .env file")
	}

	return nil
}

func Get() *Config {
	return cfg
}
