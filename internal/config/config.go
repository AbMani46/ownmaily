package config

import (
	"errors"
	"os"
)

type Config struct {
	DBUrl           string
	AppSecret       string
	InstallationURL string
	Port            string
}

func Load() (*Config, error) {
	dbUrl := os.Getenv("DB_URL")
	appSecret := os.Getenv("APP_SECRET")

	if dbUrl == "" {
		return nil, errors.New("DB_URL is required")
	}
	if appSecret == "" {
		return nil, errors.New("APP_SECRET is required")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "4400"
	}

	installURL := os.Getenv("INSTALLATION_URL")
	if installURL == "" {
		installURL = "http://localhost:4400"
	}

	return &Config{
		DBUrl:           dbUrl,
		AppSecret:       appSecret,
		InstallationURL: installURL,
		Port:            port,
	}, nil
}
