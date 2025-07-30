package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sksmith/go-base-ms/internal/api"
	"github.com/sksmith/go-base-ms/internal/config"
	"github.com/sksmith/go-base-ms/internal/db"
	"github.com/sksmith/go-base-ms/internal/health"
	"github.com/sksmith/go-base-ms/internal/kafka"
	"github.com/sksmith/go-base-ms/internal/logger"
	"github.com/sksmith/go-base-ms/internal/version"
)

// Build information set by GoReleaser
var (
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
	BuiltBy = "unknown"
)

func main() {
	// Set version info that will be injected by GoReleaser
	version.Version = Version
	version.Commit = Commit
	version.Date = Date
	version.BuiltBy = BuiltBy

	log := logger.New()

	versionInfo := version.Get()
	log.Info("go-base-ms starting",
		"version", versionInfo.Version,
		"commit", versionInfo.Commit,
		"built_at", versionInfo.Date,
		"built_by", versionInfo.BuiltBy)

	cfg, err := config.Load()
	if err != nil {
		log.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	log.Info("starting server", "port", cfg.Port)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	database, err := db.New(ctx, cfg.Database)
	if err != nil {
		log.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer database.Close()

	kafkaClient, err := kafka.New(cfg.Kafka, cfg.SchemaRegistry, log)
	if err != nil {
		log.Error("failed to connect to kafka", "error", err)
		os.Exit(1)
	}
	defer kafkaClient.Close()

	healthChecker := health.New(database, kafkaClient)

	router := api.NewRouter(log, healthChecker)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Info("server starting", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("server failed", "error", err)
			cancel()
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	select {
	case <-sigChan:
		log.Info("shutdown signal received")
	case <-ctx.Done():
		log.Info("context cancelled")
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error("server shutdown failed", "error", err)
	}

	log.Info("server stopped")
}
