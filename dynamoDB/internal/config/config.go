package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port           string
	Env            string
	DynamoDBTable  string
	JWTSecret      string
	AWSRegion      string
	AWSEndpoint    string // Empty in production, http://localhost:4566 in local
	RateLimitRPS   float64
	RateLimitBurst int
}

func Load() *Config {
	env := getEnv("ENV", "local")

	var defaultEndpoint string
	if env == "local" {
		defaultEndpoint = "http://localhost:4566"
	}

	rps, err := strconv.ParseFloat(getEnv("RATE_LIMIT_RPS", "10.0"), 64)
	if err != nil {
		rps = 10.0
	}

	burst, err := strconv.Atoi(getEnv("RATE_LIMIT_BURST", "20"))
	if err != nil {
		burst = 20
	}

	return &Config{
		Port:           getEnv("PORT", "8080"),
		Env:            env,
		DynamoDBTable:  getEnv("DYNAMODB_TABLE", "goals_tasks_app"),
		JWTSecret:      getEnv("JWT_SECRET", "goals-tasks-api-secret-key-very-secure"),
		AWSRegion:      getEnv("AWS_REGION", "us-east-1"),
		AWSEndpoint:    getEnv("AWS_ENDPOINT", defaultEndpoint),
		RateLimitRPS:   rps,
		RateLimitBurst: burst,
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
