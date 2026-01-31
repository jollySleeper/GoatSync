package config

import (
	"os"
	"strconv"
	"strings"
)

// Config holds all configuration for the GoatSync server.
// Configuration is loaded from environment variables.
type Config struct {
	// Server
	Port    string
	Debug   bool
	GinMode string

	// Security
	EncryptionSecret string   // Required: used for challenge encryption
	AllowedOrigins   []string // CORS allowed origins
	AllowedHosts     []string // Allowed Host headers

	// Authentication
	ChallengeValidSeconds int // How long login challenges are valid (default: 300)

	// Storage
	ChunkStoragePath string // Root directory for encrypted chunk files

	// Database
	DatabaseURL string // PostgreSQL connection string
	DBHost      string // Database host (alternative to URL)
	DBPort      string // Database port
	DBUser      string // Database user
	DBPassword  string // Database password
	DBName      string // Database name
	DBSSLMode   string // SSL mode (disable, require, etc.)

	// Redis (optional, for WebSocket)
	RedisURL string
}

var cfg *Config

// Load loads configuration from environment variables.
// Call this once at application startup.
func Load() *Config {
	if cfg != nil {
		return cfg
	}

	c := &Config{
		// Server
		Port:    getEnv("PORT", "3735"),
		Debug:   getEnvBool("DEBUG", false),
		GinMode: getEnv("GIN_MODE", "release"),

		// Security
		EncryptionSecret: getEnv("ENCRYPTION_SECRET", ""),
		AllowedOrigins:   splitAndTrim(getEnv("ALLOWED_ORIGINS", "*")),
		AllowedHosts:     splitAndTrim(getEnv("ALLOWED_HOSTS", "*")),

		// Authentication
		ChallengeValidSeconds: getEnvInt("CHALLENGE_VALID_SECONDS", 300),

		// Storage
		ChunkStoragePath: getEnv("CHUNK_STORAGE_PATH", "./data/chunks"),

		// Database
		DatabaseURL: getEnv("DATABASE_URL", ""),
		DBHost:      getEnv("DB_HOST", "localhost"),
		DBPort:      getEnv("DB_PORT", "5432"),
		DBUser:      getEnv("DB_USER", "postgres"),
		DBPassword:  getEnv("DB_PASSWORD", ""),
		DBName:      getEnv("DB_NAME", "goatsync"),
		DBSSLMode:   getEnv("DB_SSLMODE", "disable"),

		// Redis
		RedisURL: getEnv("REDIS_URL", ""),
	}

	cfg = c
	return cfg
}

func Get() *Config {
	return Load()
}

func getEnv(key, dflt string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return dflt
}

func getEnvBool(key string, dflt bool) bool {
	if v := os.Getenv(key); v != "" {
		b, err := strconv.ParseBool(v)
		if err == nil {
			return b
		}
	}
	return dflt
}

func getEnvInt(key string, dflt int) int {
	if v := os.Getenv(key); v != "" {
		i, err := strconv.Atoi(v)
		if err == nil {
			return i
		}
	}
	return dflt
}

func splitAndTrim(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	if len(out) == 0 {
		return []string{"*"}
	}
	return out
}
