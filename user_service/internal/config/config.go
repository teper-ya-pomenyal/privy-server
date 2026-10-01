package config

import (
	"log"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Port           string
	UserDB         *UserDBConfig
	UserCacheDB    *UserDBCacheConfig
	TTLRefresh     time.Duration
	TTLAccess      time.Duration
	TTLStream      time.Duration
	PrivateKeyPath string
}

func LoadConfig() *Config {

	ttlRefresh := getIntEnvOrDefault("TTL_REFRESH_TIME", 604800)
	ttlAccess := getIntEnvOrDefault("TTL_ACCESS_TIME", 900)
	ttlStream := getIntEnvOrDefault("TTL_STREAM_TIME", 600)

	port := getEnvOrDefault("PORT", "50051")
	privateKey := getEnvOrDefault("PRIVATE_KEY_PATH", "keys/private.pem")
	return &Config{
		Port:           port,
		UserDB:         NewUserDBConfig(),
		UserCacheDB:    NewUserDBCacheConfig(),
		TTLRefresh:     time.Duration(ttlRefresh) * time.Second,
		TTLAccess:      time.Duration(ttlAccess) * time.Second,
		TTLStream:      time.Duration(ttlStream) * time.Second,
		PrivateKeyPath: privateKey,
	}
}

func getEnvOrDefault(key, defaultVal string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Printf("environment variable %s is not set, using default", key)
		return defaultVal
	}
	return v
}

// getDurationEnvOrDefault expects Go duration format, e.g. "5s", "30m", "1h".
func getDurationEnvOrDefault(key string, defaultVal time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return defaultVal
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		log.Fatalf("failed to parse variable %s as duration: %s", key, err)
	}
	return d
}

func getIntEnvOrDefault(key string, defaultVal int) int {
	v := os.Getenv(key)
	if v == "" {
		return defaultVal
	}
	intParam, err := strconv.Atoi(v)
	if err != nil {
		log.Fatalf("failed to convert variable %s to a number: %s", key, err)
	}
	return intParam
}
