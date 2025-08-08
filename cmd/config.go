package cmd

import "github.com/caarlos0/env/v11"

type Config struct {
	Port               string `env:"PORT" envDefault:"8080"`
	DatabaseURL        string `env:"DATABASE_URL"`
	GCloudClientID     string `env:"GCLOUD_CLIENT_ID"`
	GCloudClientSecret string `env:"GCLOUD_CLIENT_SECRET"`
	GCloudProjectID    string `env:"GCLOUD_PROJECT_ID"`
	BaseURL            string `env:"BASE_URL" envDefault:"http://localhost:8080"`
}

func LoadConfig() Config {
	var c Config
	_ = env.Parse(&c)
	return c
}
