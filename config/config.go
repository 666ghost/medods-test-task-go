package config

import (
	"github.com/joho/godotenv"
	"log"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Env         string
	DbUser      string
	DbName      string
	DbPassword  string
	DbAddresses []string
	DbPort      string
	TokenSecret string
}

// New returns a new Config struct
func New() *Config {
	return &Config{
		Env:         getEnv("APP_ENV", "dev"),
		DbName:      getEnv("DB_NAME", "root"),
		DbUser:      getEnv("DB_USER", "root"),
		DbPassword:  getEnv("DB_PASSWORD", "password"),
		DbAddresses: getEnvAsSlice("DB_SET_ADDRESSES", []string{"mongo-0", "mongo-1", "mongo-2"}, ","),
		DbPort:      getEnv("DB_PORT", "27017"),

		TokenSecret: getEnv("TOKEN_SECRET", "secret"),
	}
}

func LoadFromFile() {
	if New().Env != "dev" {
		return
	}
	if err := godotenv.Load(); err != nil {
		log.Fatal("No .env file found")
	}
}

// Simple helper function to read an environment or return a default value
func getEnv(key string, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}

	return defaultVal
}

// Simple helper function to read an environment variable into integer or return a default value
func getEnvAsInt(name string, defaultVal int) int {
	valueStr := getEnv(name, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}

	return defaultVal
}

// Helper to read an environment variable into a bool or return default value
func getEnvAsBool(name string, defaultVal bool) bool {
	valStr := getEnv(name, "")
	if val, err := strconv.ParseBool(valStr); err == nil {
		return val
	}

	return defaultVal
}

// Helper to read an environment variable into a string slice or return default value
func getEnvAsSlice(name string, defaultVal []string, sep string) []string {
	valStr := getEnv(name, "")

	if valStr == "" {
		return defaultVal
	}

	val := strings.Split(valStr, sep)

	return val
}
