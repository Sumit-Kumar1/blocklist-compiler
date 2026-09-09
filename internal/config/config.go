package config

import (
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// Config holds the runtime settings resolved from the environment.
type Config struct {
	ConfigName string
	ConfigPath string
	// OutputFile is the published list. It is replaced atomically.
	OutputFile string
	// MinRules is the sanity floor: a compiled producing fewer rules is discarded
	// and the previously published list is left in place.
	MinRules int
	// Interval is the delay between compile cycles. Zero means compile once and exit.
	Interval time.Duration
	// ListenAddr serves the published list. Empty disables the server.
	ListenAddr string

	AghAPI  string
	AghUser string
	AghPass string
}

// Load reads .env when present and falls back to defaults for missing keys.
func Load() *Config {
	_ = godotenv.Load(".env")

	return &Config{
		ConfigName: getEnv("CONFIG_FILE", "config.json"),
		ConfigPath: getEnv("CONFIG_PATH", "./"),
		OutputFile: getEnv("OUTPUT_FILE", "output.txt"),
		MinRules:   getInt("MIN_RULES", 1000),
		Interval:   getDuration("INTERVAL", 24*time.Hour),
		ListenAddr: getEnv("LISTEN_ADDR", ""),
		AghAPI:     getEnv("AGH_API", ""),
		AghUser:    getEnv("AGH_USER", ""),
		AghPass:    getEnv("AGH_PASS", ""),
	}
}

func getEnv(key, defaultVal string) string {
	if val := strings.TrimSpace(os.Getenv(key)); val != "" {
		return val
	}

	return defaultVal
}

func getInt(key string, defaultVal int) int {
	val, err := strconv.Atoi(getEnv(key, ""))
	if err != nil {
		return defaultVal
	}

	return val
}

func getDuration(key string, defaultVal time.Duration) time.Duration {
	val, err := time.ParseDuration(getEnv(key, ""))
	if err != nil {
		return defaultVal
	}

	return val
}
