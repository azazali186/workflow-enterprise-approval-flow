package config

import (
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"
)

// Build-time version information (injected via -ldflags).
var (
	// Version is the application version. Override with:
	//   -ldflags "-X github.com/aeroxe/approval-flow/internal/config.Version=1.0.0"
	Version = "dev"
	// BuildTime is the build timestamp. Override with:
	//   -ldflags "-X github.com/aeroxe/approval-flow/internal/config.BuildTime=2026-08-08T00:00:00Z"
	BuildTime = "unknown"
)

// DefaultDatabaseURL is the development-only default; production deployments
// must override it via DATABASE_URL.
const DefaultDatabaseURL = "postgres://aeroxe:secret@localhost:5432/approval-flow?sslmode=disable"

type Config struct {
	AppName        string
	Env            string
	LogLevel       string
	ServerPort     int
	GRPCPort       int
	DatabaseURL    string
	RedisURL       string
	NATSURL        string
	JWTSecret      string
	JWTExpiry      string
	RateLimitRPS   int
	RateLimitBurst int
	// RateLimitWindow is the fixed window (seconds) for the per-IP rate limit.
	// RateLimitRPS requests are allowed per window per IP.
	RateLimitWindow  int
	WSMaxConnections int
	WSPingInterval   int
	LogFilePath      string
	LogMaxSize       int
	LogMaxBackups    int
	LogMaxAge        int
	LogCompress      bool
	// AdminEmail/AdminPassword bootstrap the initial administrator account at
	// startup (created only if they do not already exist).
	AdminEmail    string
	AdminPassword string
	// MigrationsPath points to the golang-migrate SQL directory. If no .sql
	// files are found there, AutoMigrate is used as a development fallback.
	MigrationsPath string
	// CORSAllowedOrigins is a comma-separated allow-list of origins. An empty
	// value means "*" (development default).
	CORSAllowedOrigins []string
	// SwaggerHost overrides the host advertised in the Swagger docs.
	SwaggerHost string
	// SMTP settings for email notifications. If SMTPHost is empty, email
	// delivery is skipped (in-app notifications still work).
	SMTPHost     string
	SMTPPort     int
	SMTPUser     string
	SMTPPassword string
	SMTPFrom     string
	*zap.Logger
}

func Load() *Config {
	cfg := &Config{
		AppName:          getEnv("APP_NAME", "approval-flow"),
		Env:              getEnv("ENV", "development"),
		LogLevel:         getEnv("LOG_LEVEL", "debug"),
		ServerPort:       getEnvAsInt("SERVER_PORT", 8080),
		GRPCPort:         getEnvAsInt("GRPC_PORT", 9090),
		DatabaseURL:      getEnv("DATABASE_URL", DefaultDatabaseURL),
		RedisURL:         getEnv("REDIS_URL", "redis://localhost:6379"),
		NATSURL:          getEnv("NATS_URL", "nats://localhost:4222"),
		JWTSecret:        getEnv("JWT_SECRET", ""),
		JWTExpiry:        getEnv("JWT_EXPIRY", "24h"),
		RateLimitRPS:     getEnvAsInt("RATE_LIMIT_RPS", 100),
		RateLimitBurst:   getEnvAsInt("RATE_LIMIT_BURST", 200),
		RateLimitWindow:  getEnvAsInt("RATE_LIMIT_WINDOW", 60),
		WSMaxConnections: getEnvAsInt("WS_MAX_CONNECTIONS", 1000),
		WSPingInterval:   getEnvAsInt("WS_PING_INTERVAL", 30),
		LogFilePath:      getEnvOrEmpty("LOG_FILE_PATH"),
		LogMaxSize:       getEnvAsInt("LOG_MAX_SIZE", 100),
		LogMaxBackups:    getEnvAsInt("LOG_MAX_BACKUPS", 10),
		LogMaxAge:        getEnvAsInt("LOG_MAX_AGE", 30),
		LogCompress:      getEnvAsBool("LOG_COMPRESS", true),
		AdminEmail:       getEnv("ADMIN_EMAIL", ""),
		AdminPassword:    getEnv("ADMIN_PASSWORD", ""),
		MigrationsPath:   getEnv("MIGRATIONS_PATH", "./migrations"),
		SwaggerHost:      getEnv("SWAGGER_HOST", ""),
		SMTPHost:         getEnv("SMTP_HOST", ""),
		SMTPPort:         getEnvAsInt("SMTP_PORT", 587),
		SMTPUser:         getEnv("SMTP_USER", ""),
		SMTPPassword:     getEnv("SMTP_PASSWORD", ""),
		SMTPFrom:         getEnv("SMTP_FROM", "no-reply@approval-flow.local"),
	}

	// Parse comma-separated CORS allow-list
	if raw := os.Getenv("CORS_ALLOWED_ORIGINS"); raw != "" {
		for _, origin := range strings.Split(raw, ",") {
			origin = strings.TrimSpace(origin)
			if origin != "" {
				cfg.CORSAllowedOrigins = append(cfg.CORSAllowedOrigins, origin)
			}
		}
	}
	if len(cfg.CORSAllowedOrigins) == 0 {
		cfg.CORSAllowedOrigins = []string{"*"}
	}

	// Default log file path for local development only.
	if cfg.LogFilePath == "" {
		if cfg.Env == "production" {
			// Containers should log to stdout for docker logs / log shippers.
			cfg.LogFilePath = ""
		} else {
			cfg.LogFilePath = "logs/approval-flow.log"
		}
	}

	cfg.validate()

	logger, err := NewLogger(cfg.Env, cfg.LogLevel, cfg)
	if err != nil {
		log.Fatalf("failed to init logger: %v", err)
	}
	cfg.Logger = logger

	return cfg
}

// validate enforces fail-fast checks for required and production-only settings.
func (c *Config) validate() {
	if c.JWTSecret == "" {
		log.Fatal("JWT_SECRET environment variable is required")
	}

	if c.Env == "production" {
		if len(c.JWTSecret) < 32 {
			log.Fatal("JWT_SECRET must be at least 32 characters in production")
		}
		if looksLikePlaceholder(c.JWTSecret) || looksLikePlaceholder(c.AdminPassword) {
			log.Fatal("JWT_SECRET and ADMIN_PASSWORD must not be placeholder values in production (e.g. CHANGE_ME or your-secret)")
		}
		if c.DatabaseURL == DefaultDatabaseURL {
			log.Fatal("DATABASE_URL must be explicitly configured in production")
		}
		if c.AdminEmail == "" || c.AdminPassword == "" {
			log.Fatal("ADMIN_EMAIL and ADMIN_PASSWORD are required in production to bootstrap the administrator account")
		}
	}
}

// looksLikePlaceholder detects common placeholder values that must never be
// used in production (a guard against misapplied templates and example envs).
func looksLikePlaceholder(value string) bool {
	lower := strings.ToLower(value)
	for _, marker := range []string{"change_me", "your-secret", "change-me", "changeme"} {
		if strings.Contains(lower, marker) {
			return true
		}
	}
	return false
}

// JWTExpiryDuration returns the parsed JWT expiry duration (default 24h).
func (c *Config) JWTExpiryDuration() time.Duration {
	if d, err := time.ParseDuration(c.JWTExpiry); err == nil && d > 0 {
		return d
	}
	return 24 * time.Hour
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

// getEnvOrEmpty returns the raw env value, allowing it to be explicitly empty
// (used to disable file logging in containers).
func getEnvOrEmpty(key string) string {
	return os.Getenv(key)
}

func getEnvAsInt(key string, defaultVal int) int {
	if val := os.Getenv(key); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}
	return defaultVal
}

func getEnvAsBool(key string, defaultVal bool) bool {
	if val := os.Getenv(key); val != "" {
		if b, err := strconv.ParseBool(val); err == nil {
			return b
		}
	}
	return defaultVal
}
