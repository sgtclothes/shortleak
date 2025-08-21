package config

import (
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Env      string
	Database string
	User     string
	Password string
	Host     string
	Dialect  string
	Port     string
}

var LogFatalf = log.Fatalf

func LoadConfig() Config {
	err := godotenv.Load()
	if err != nil {
		log.Println("⚠️ .env file not found:", err)
	}

	env := os.Getenv("NODE_ENV")
	if env == "" {
		env = "development"
	}

	suffix := "_" + strings.ToUpper(env)

	cfg := Config{
		Env:      env,
		Database: getEnv("DB_DATABASE"+suffix, ""),
		User:     getEnv("DB_USERNAME"+suffix, ""),
		Password: getEnv("DB_PASSWORD"+suffix, ""),
		Host:     getEnv("DB_HOST"+suffix, ""),
		Dialect:  getEnv("DB_DIALECT"+suffix, "postgres"),
		Port:     getEnv("DB_PORT"+suffix, "5432"),
	}

	if cfg.Database == "" {
		LogFatalf("❌ Database config for %s not found", env)
	}

	return cfg
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

func toUpper(s string) string {
	if len(s) == 0 {
		return s
	}
	return string([]rune(s)[0]-32) + s[1:]
}
