package config

import "os"

type Config struct {
	DatabaseURL      string
	GooglePlacesKey  string
	PythonServiceURL string
	Port             string
}

func New() *Config {
	return &Config{
		DatabaseURL:      os.Getenv("DATABASE_URL"),
		GooglePlacesKey:  os.Getenv("GOOGLE_PLACES_API_KEY"),
		PythonServiceURL: getEnv("PYTHON_SERVICE_URL", "http://localhost:8001"),
		Port:             getEnv("GO_SERVICE_PORT", "8080"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
