package config

import (
	"log"
	"os"
)

type Config struct {
	TrackStoragePath      string
	CatalogServiceAddress string
	Port                  string
}

func NewConfig() *Config {
	return &Config{
		TrackStoragePath:      getEnv("TRACK_STORAGE_PATH", "data/tracks"),
		CatalogServiceAddress: getEnv("CATALOG_SERVICE_ADDRESS", "localhost:50053"),
		Port:                  getEnv("STREAMING_SERVICE_PORT", "50054"),
	}
}

func getEnv(key, def string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Printf("environment variable %s is not set, using default", key)
		return def
	}
	return v
}
