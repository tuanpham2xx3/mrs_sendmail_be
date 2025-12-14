package config

import (
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Server    ServerConfig
	Redis     RedisConfig
	SMTP      SMTPConfig
	Security  SecurityConfig
	RateLimit RateLimitConfig
	Code      CodeConfig
}

type ServerConfig struct {
	Port string
	Host string
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

type SMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	FromName string
}

type SecurityConfig struct {
	APIKeys []string
}

type RateLimitConfig struct {
	EmailPerHour int
	IPPerHour    int
}

type CodeConfig struct {
	ExpireMinutes     int
	Length            int
	DefaultSystemName string
}

func Load() (*Config, error) {
	// Load .env file if exists
	_ = godotenv.Load()

	config := &Config{
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8200"),
			Host: getEnv("SERVER_HOST", "0.0.0.0"),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnv("REDIS_PORT", "6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvAsInt("REDIS_DB", 0),
		},
		SMTP: SMTPConfig{
			Host:     getEnv("SMTP_HOST", "smtp.gmail.com"),
			Port:     getEnvAsInt("SMTP_PORT", 587),
			Username: getEnv("SMTP_USERNAME", ""),
			Password: getEnv("SMTP_PASSWORD", ""),
			FromName: getEnv("SMTP_FROM_NAME", "Fix4Home System"),
		},
		Security: SecurityConfig{
			APIKeys: getEnvAsSlice("API_KEYS", []string{}),
		},
		RateLimit: RateLimitConfig{
			EmailPerHour: getEnvAsInt("RATE_LIMIT_EMAIL_PER_HOUR", 5),
			IPPerHour:    getEnvAsInt("RATE_LIMIT_IP_PER_HOUR", 30),
		},
		Code: CodeConfig{
			ExpireMinutes:     getEnvAsInt("CODE_EXPIRE_MINUTES", 30),
			Length:            getEnvAsInt("CODE_LENGTH", 6),
			DefaultSystemName: getEnv("DEFAULT_SYSTEM_NAME", "Fix4Home"),
		},
	}

	return config, nil
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

func getEnvAsSlice(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		return strings.Split(value, ",")
	}
	return defaultValue
}
