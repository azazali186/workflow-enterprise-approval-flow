package config

import (
	"os"
	"path/filepath"

	"gopkg.in/natefinch/lumberjack.v2"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func NewLogger(env, level string, cfg *Config) (*zap.Logger, error) {
	var zapLevel zapcore.Level
	switch level {
	case "debug":
		zapLevel = zap.DebugLevel
	case "info":
		zapLevel = zap.InfoLevel
	case "warn":
		zapLevel = zap.WarnLevel
	case "error":
		zapLevel = zap.ErrorLevel
	default:
		zapLevel = zap.InfoLevel
	}

	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "message",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	consoleEncoder := zapcore.NewConsoleEncoder(encoderConfig)
	fileEncoder := zapcore.NewJSONEncoder(encoderConfig)

	logDir := filepath.Dir(cfg.LogFilePath)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, err
	}

	lumberjackLogger := &lumberjack.Logger{
		Filename:   cfg.LogFilePath,
		MaxSize:    cfg.LogMaxSize,
		MaxBackups: cfg.LogMaxBackups,
		MaxAge:     cfg.LogMaxAge,
		Compress:   cfg.LogCompress,
	}

	fileWriteSyncer := zapcore.AddSync(lumberjackLogger)
	consoleWriteSyncer := zapcore.AddSync(os.Stdout)

	core := zapcore.NewTee(
		zapcore.NewCore(fileEncoder, fileWriteSyncer, zapLevel),
		zapcore.NewCore(consoleEncoder, consoleWriteSyncer, zapLevel),
	)

	logger := zap.New(core)
	if env == "development" {
		logger = logger.WithOptions(zap.AddStacktrace(zap.ErrorLevel))
	}

	return logger, nil
}

func NewNopLogger() *zap.Logger {
	return zap.NewNop()
}

func GetLogLevelFromEnv() string {
	return os.Getenv("LOG_LEVEL")
}
