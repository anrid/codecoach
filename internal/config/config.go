package config

import (
	"os"
	"time"
)

// Config ...
type Config struct {
	Host                string
	DBHost              string
	DBPort              string
	DBUser              string
	DBPass              string
	DBName              string
	GithubOAuthClientID string
	TokenExpires        time.Duration
}

// New ...
func New() *Config {
	return &Config{
		Host:                mustEnv("HOST"),
		DBHost:              mustEnv("DB_HOST"),
		DBPort:              mustEnv("DB_PORT"),
		DBUser:              mustEnv("DB_USER"),
		DBPass:              mustEnv("DB_PASS"),
		DBName:              mustEnv("DB_NAME"),
		GithubOAuthClientID: mustEnv("GITHUB_OAUTH_CLIENT_ID"),
		TokenExpires:        24 * 7 * time.Hour,
	}
}

func mustEnv(env string) string {
	v := os.Getenv(env)
	if v == "" {
		panic("missing env " + env)
	}
	return v
}
