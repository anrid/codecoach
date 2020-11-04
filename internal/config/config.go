package config

import (
	"os"
	"time"
)

// Config ...
type Config struct {
	Host               string
	DBHost             string
	DBPort             string
	DBUser             string
	DBPass             string
	DBName             string
	GithubClientID     string
	GithubClientSecret string
	GithubRedirectURI  string
	TokenExpires       time.Duration
}

// New ...
func New() *Config {
	return &Config{
		Host:               mustEnv("HOST"),
		DBHost:             mustEnv("DB_HOST"),
		DBPort:             mustEnv("DB_PORT"),
		DBUser:             mustEnv("DB_USER"),
		DBPass:             mustEnv("DB_PASS"),
		DBName:             mustEnv("DB_NAME"),
		GithubClientID:     mustEnv("GITHUB_CLIENT_ID"),
		GithubClientSecret: mustEnv("GITHUB_CLIENT_SECRET"),
		GithubRedirectURI:  mustEnv("GITHUB_REDIRECT_URI"),
		TokenExpires:       24 * 7 * time.Hour,
	}
}

func mustEnv(env string) string {
	v := os.Getenv(env)
	if v == "" {
		panic("missing env " + env)
	}
	return v
}
