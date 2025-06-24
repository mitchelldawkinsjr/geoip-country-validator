package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port     int
	DBPath   string
	LogLevel string
}

func Load() *Config {
	return &Config{
		Port:     getEnvAsInt("PORT", 8080),
		DBPath:   getEnv("GEOIP_DB_PATH", "./GeoLite2-Country.mmdb"),
		LogLevel: getEnv("LOG_LEVEL", "info"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
