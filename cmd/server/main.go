package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/alexghr/alexghr.me/internal/config"
	"github.com/alexghr/alexghr.me/internal/state"
	"github.com/alexghr/alexghr.me/internal/web"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	cfg, err := config.FromEnv()
	if err != nil {
		logger.Error("config issue", "error", err)
		os.Exit(1)
	}

	if err := run(logger, cfg); err != nil {
		logger.Error("server stopped", "error", err)
		os.Exit(1)
	}
}

func run(logger *slog.Logger, cfg config.Config) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	db, err := state.Open(ctx, filepath.Join(cfg.StateDir, "db.sqlite3"), 1)
	if err != nil {
		return err
	}

	defer func() {
		if err := db.Close(); err != nil {
			logger.Error("close database", "error", err)
		}
	}()

	if err := db.Migrate(ctx); err != nil {
		return fmt.Errorf("migrate db: %w", err)
	}

	h, err := web.NewHTTPHandler(db, logger, cfg.Debug, cfg.RateLimitBurst, cfg.RateLimitRefill, cfg.AdminPassword)
	if err != nil {
		return fmt.Errorf("http handler: %w", err)
	}
	server := &http.Server{
		Handler:           h,
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       time.Minute,
		MaxHeaderBytes:    http.DefaultMaxHeaderBytes,
	}

	if !cfg.Debug {
		server.WriteTimeout = 15 * time.Second
	}

	listener, err := net.Listen("tcp", cfg.ListenAddr)
	if err != nil {
		return fmt.Errorf("listen web UI: %w", err)
	}

	serveErr := make(chan error, 1)
	go func() {
		serveErr <- server.Serve(listener)
	}()

	logger.Info("website listening", "version", web.Version, "url", "http://"+listener.Addr().String())

	select {
	case err := <-serveErr:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("serve web UI: %w", err)
		}
		return nil
	case <-ctx.Done():
		logger.Info("shutting down website")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shut down web UI: %w", err)
	}

	if err := <-serveErr; err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("serve web UI: %w", err)
	}

	return nil
}
