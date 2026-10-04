package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Port             string
	GinMode          string
	DBHost           string
	DBPort           string
	DBUser           string
	DBPassword       string
	DBName           string
	DBSSLMode        string
	RedisAddr        string
	RedisPassword    string
	RedisDB          int
	FrequencyCap     int
	CacheTTLSeconds  int
}

var App *Config

func Load() {
	// Load .env if exists (ignore error in production)
	_ = godotenv.Load()

	redisDB, _ := strconv.Atoi(getEnv("REDIS_DB", "0"))
	freqCap, _ := strconv.Atoi(getEnv("FREQUENCY_CAP", "3"))
	cacheTTL, _ := strconv.Atoi(getEnv("CACHE_TTL_SECONDS", "60"))

	App = &Config{
		Port:            getEnv("PORT", "8080"),
		GinMode:         getEnv("GIN_MODE", "debug"),
		DBHost:          getEnv("DB_HOST", "localhost"),
		DBPort:          getEnv("DB_PORT", "5432"),
		DBUser:          getEnv("DB_USER", "postgres"),
		DBPassword:      getEnv("DB_PASSWORD", ""),
		DBName:          getEnv("DB_NAME", "adserver"),
		DBSSLMode:       getEnv("DB_SSLMODE", "disable"),
		RedisAddr:       getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword:   getEnv("REDIS_PASSWORD", ""),
		RedisDB:         redisDB,
		FrequencyCap:    freqCap,
		CacheTTLSeconds: cacheTTL,
	}

	log.Println("✅ Config loaded successfully")
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
