package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Port               string
	DBPath             string
	BaseURL            string
	JWTSecret          string
	RateLimitPerMinute int
}

func LoadConfig() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, using environment variables")
	}

	rateLimitStr := getEnv("RATE_LIMIT_PER_MINUTE", "10")
	rateLimit := 10
	if val, err := strconv.Atoi(rateLimitStr); err == nil {
		rateLimit = val
	}

	return &Config{
		Port:               getEnv("PORT", "8080"),
		DBPath:             getEnv("DB_PATH", "urlshortener.db"),
		BaseURL:            getEnv("BASE_URL", "http://localhost:8080"),
		JWTSecret:          getEnv("JWT_SECRET", "supersecret"),
		RateLimitPerMinute: rateLimit,
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
