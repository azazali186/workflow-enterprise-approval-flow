package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aeroxe/approval-flow/internal/config"
	"github.com/aeroxe/approval-flow/internal/server"
	"go.uber.org/zap"
)

func main() {
	cfg := config.Load()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		cfg.Info("received shutdown signal")
		cancel()
	}()

	srv, err := server.NewServer(cfg)
	if err != nil {
		cfg.Fatal("failed to create server", zap.Error(err))
	}

	if err := srv.Migrate(); err != nil {
		cfg.Fatal("failed to run auto-migration", zap.Error(err))
	}

	go func() {
		if err := srv.Start(); err != nil && err != http.ErrServerClosed {
			cfg.Fatal("server error", zap.Error(err))
		}
	}()

	cfg.Info(fmt.Sprintf("server started on port %d", cfg.ServerPort))

	<-ctx.Done()
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		cfg.Error("server shutdown error", zap.Error(err))
	}

	cfg.Info("server stopped")
}
