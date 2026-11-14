package config

import "os"

type Config struct {
	Port        string
	DatabaseURL string
	AllowedOrigin string
}

func Load() Config {
	port := os.Getenv("PORT")
	if port == "" { port = "8080" }
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" { databaseURL = "postgres://mapup:mapup@localhost:5432/mapup?sslmode=disable" }
	origin := os.Getenv("ALLOWED_ORIGIN")
	if origin == "" { origin = "http://localhost:5173" }
	return Config{Port: port, DatabaseURL: databaseURL, AllowedOrigin: origin}
}
