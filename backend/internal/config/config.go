package config

import (
	"log"
	"os"
	"strconv"

	"go.uber.org/zap"
)

type Config struct {
	AppName          string
	Env              string
	LogLevel         string
	ServerPort       int
	GRPCPort         int
	DatabaseURL      string
	RedisURL         string
	NATSURL          string
	JWTSecret        string
	JWTExpiry        string
	RateLimitRPS     int
	RateLimitBurst   int
	WSMaxConnections int
	WSPingInterval   int
	*zap.Logger
}

func Load() *Config {
	cfg := &Config{
		AppName:          getEnv("APP_NAME", "approval-flow"),
		Env:              getEnv("ENV", "development"),
		LogLevel:         getEnv("LOG_LEVEL", "debug"),
		ServerPort:       getEnvAsInt("SERVER_PORT", 8080),
		GRPCPort:         getEnvAsInt("GRPC_PORT", 9090),
		DatabaseURL:      getEnv("DATABASE_URL", "postgres://aeroxe:secret@localhost:5432/approval-flow?sslmode=disable"),
		RedisURL:         getEnv("REDIS_URL", "redis://localhost:6379"),
		NATSURL:          getEnv("NATS_URL", "nats://localhost:4222"),
		JWTSecret:        getEnv("JWT_SECRET", "your-secret-key"),
		JWTExpiry:        getEnv("JWT_EXPIRY", "24h"),
		RateLimitRPS:     getEnvAsInt("RATE_LIMIT_RPS", 100),
		RateLimitBurst:   getEnvAsInt("RATE_LIMIT_BURST", 200),
		WSMaxConnections: getEnvAsInt("WS_MAX_CONNECTIONS", 1000),
		WSPingInterval:   getEnvAsInt("WS_PING_INTERVAL", 30),
	}

	logger, err := NewLogger(cfg.Env, cfg.LogLevel)
	if err != nil {
		log.Fatalf("failed to init logger: %v", err)
	}
	cfg.Logger = logger

	return cfg
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func getEnvAsInt(key string, defaultVal int) int {
	if val := os.Getenv(key); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}
	return defaultVal
}
