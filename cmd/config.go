package cmd

import "github.com/caarlos0/env/v11"

type Config struct {
	Port        string `env:"PORT" envDefault:"8080"`
	DatabaseURL string `env:"DATABASE_URL"`
}

func LoadConfig() Config {
	var c Config
	_ = env.Parse(&c)
	return c
}
