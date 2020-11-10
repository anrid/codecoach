package config

import (
	"os"
	"path"
	"runtime"
	"time"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
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
	GithubAccessToken  string
	TokenExpires       time.Duration
}

// New ...
func New() *Config {
	// Load .env files.
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "development"
	}

	root := ""
	if env == "test" {
		// Get project root dir.
		_, filename, _, _ := runtime.Caller(0)
		root = path.Join(path.Dir(filename), "..", "..")
	}

	envFile := func(n string) error {
		return godotenv.Load(path.Join(root, n))
	}

	_ = envFile(".env." + env + ".local")
	if env != "test" {
		_ = envFile(".env.local")
	}
	err := envFile(".env")
	if err != nil {
		zap.S().Fatalw("error loading .env file", "error", err)
	}

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
		GithubAccessToken:  mustEnv("GITHUB_ACCESS_TOKEN"),
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
