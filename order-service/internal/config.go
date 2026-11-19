package internal

import "os"

type Config struct {
	Port        string
	DatabaseURL string
	AuthURL     string
	CatalogURL  string
}

func LoadConfig() Config {
	return Config{
		Port:        env("PORT", "8084"),
		DatabaseURL: env("DATABASE_URL", "postgres://orders:orders@localhost:5436/orders?sslmode=disable"),
		AuthURL:     env("AUTH_URL", "http://localhost:8081"),
		CatalogURL:  env("CATALOG_URL", "http://localhost:8082"),
	}
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
