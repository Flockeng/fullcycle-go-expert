package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	RateLimitIP             int
	RateLimitIPBlockTime    time.Duration
	RateLimitTokenDefault   int
	RateLimitTokenBlockTime time.Duration
	RedisHost               string
	RedisPort               string
	RedisPassword           string
	RedisDB                 int
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{}

	cfg.RateLimitIP = getEnvAsInt("RATE_LIMIT_IP", 10)
	cfg.RateLimitIPBlockTime = time.Duration(getEnvAsInt("RATE_LIMIT_IP_BLOCK_TIME", 300)) * time.Second
	cfg.RateLimitTokenDefault = getEnvAsInt("RATE_LIMIT_TOKEN_DEFAULT", 100)
	cfg.RateLimitTokenBlockTime = time.Duration(getEnvAsInt("RATE_LIMIT_TOKEN_BLOCK_TIME", 300)) * time.Second

	cfg.RedisHost = getEnvAsString("REDIS_HOST", "localhost")
	cfg.RedisPort = getEnvAsString("REDIS_PORT", "6379")
	cfg.RedisPassword = getEnvAsString("REDIS_PASSWORD", "")
	cfg.RedisDB = getEnvAsInt("REDIS_DB", 0)

	return cfg, nil
}

func getEnvAsString(key string, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}

	return value
}

func (c *Config) RedisAddr() string {
	return fmt.Sprintf("%s:%s", c.RedisHost, c.RedisPort)
}
