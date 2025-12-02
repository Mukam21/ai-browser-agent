package config

import (
	"os"
	"strconv"
)

type Config struct {
	OpenAIAPIKey string
	Model        string
	MaxSteps     int
	Debug        bool
	Headless     bool
	Timeout      int
	MaxTokens    int
}

func Load() *Config {
	maxSteps, _ := strconv.Atoi(getEnv("MAX_STEPS", "30"))
	timeout, _ := strconv.Atoi(getEnv("TIMEOUT", "60"))
	maxTokens, _ := strconv.Atoi(getEnv("MAX_TOKENS", "1000"))

	return &Config{
		OpenAIAPIKey: os.Getenv("OPENAI_API_KEY"),
		Model:        getEnv("MODEL", "gpt-4o-mini"),
		MaxSteps:     maxSteps,
		Debug:        getEnv("DEBUG", "true") == "true",
		Headless:     getEnv("HEADLESS", "false") == "true",
		Timeout:      timeout,
		MaxTokens:    maxTokens,
	}
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
