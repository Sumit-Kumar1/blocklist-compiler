package config

import (
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type config struct {
	AppName string
	Config  struct {
		Name string
		Path string
	}
}

func LoadConfig() *config {
	var cfg config

	_ = godotenv.Load(".env")

	cfg.AppName = getEnv("APP_NAME", "blocklist-compiler")
	cfg.Config = struct {
		Name string
		Path string
	}{
		Name: getEnv("CONFIG_FILE", "config.json"),
		Path: getEnv("CONFIG_PATH", "./"),
	}

	return &cfg
}

func getEnv[T comparable](key string, defaultVal T) T {
	var (
		val  string
		resp T
	)

	if val = strings.TrimSpace(os.Getenv(key)); val == "" {
		return defaultVal
	}

	resp, ok := any(val).(T)
	if !ok {
		return defaultVal
	}

	return resp
}
