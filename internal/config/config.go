package config

import (
	"log/slog"
	"os"
	"path/filepath"
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
	// CacheDir keeps each remote source's last good download, used when a
	// later fetch of that source fails.
	CacheDir string
	// MaxSourceBytes rejects a source whose body is larger than this.
	MaxSourceBytes int64

	AghAPI  string
	AghUser string
	AghPass string
}

// Load reads .env when present and falls back to defaults for missing keys.
func Load() *Config {
	_ = godotenv.Load(".env")

	outputFile := getEnv("OUTPUT_FILE", "output.txt")

	return &Config{
		ConfigName:     getEnv("CONFIG_FILE", "config.json"),
		ConfigPath:     getEnv("CONFIG_PATH", "./"),
		OutputFile:     outputFile,
		MinRules:       getInt("MIN_RULES", 1000),
		Interval:       getDuration("INTERVAL", 24*time.Hour),
		ListenAddr:     getEnv("LISTEN_ADDR", ""),
		CacheDir:       getEnv("CACHE_DIR", filepath.Join(filepath.Dir(outputFile), ".cache")),
		MaxSourceBytes: int64(getInt("MAX_SOURCE_MB", 100)) << 20,
		AghAPI:         getEnv("AGH_API", ""),
		AghUser:        getEnv("AGH_USER", ""),
		AghPass:        getEnv("AGH_PASS", ""),
	}
}

func getEnv(key, defaultVal string) string {
	if val := strings.TrimSpace(os.Getenv(key)); val != "" {
		return val
	}

	return defaultVal
}

func getInt(key string, defaultVal int) int {
	raw := getEnv(key, "")
	if raw == "" {
		return defaultVal
	}

	val, err := strconv.Atoi(raw)
	if err != nil {
		warnDefault(key, raw, defaultVal)

		return defaultVal
	}

	return val
}

func getDuration(key string, defaultVal time.Duration) time.Duration {
	raw := getEnv(key, "")
	if raw == "" {
		return defaultVal
	}

	val, err := time.ParseDuration(raw)
	if err != nil {
		warnDefault(key, raw, defaultVal)

		return defaultVal
	}

	return val
}

// warnDefault logs a value that failed to parse, so a typo such as INTERVAL=24
// (no unit) does not silently fall back to the default.
func warnDefault(key, raw string, defaultVal any) {
	slog.Warn("ignoring unparsable env value, using default",
		slog.String("key", key), slog.String("value", raw), slog.Any("default", defaultVal))
}
