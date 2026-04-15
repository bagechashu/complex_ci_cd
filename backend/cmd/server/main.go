package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-chi/chi/v5"
	"github.com/op/release-control/internal/config"
	"github.com/op/release-control/internal/database"
	"github.com/op/release-control/internal/handlers"
	"github.com/op/release-control/internal/repository"
	"github.com/op/release-control/pkg/logger"
)

// Use air for live reload during development
func main() {
	cfg := config.LoadConfig()
	log := logger.NewLogger()

	log.Info("Starting Release Control Server")
	log.Info("Environment: %s", cfg.Environment)
	log.Info("Database: %s", cfg.DatabasePath)

	sqlDB, err := database.Init(cfg.DatabasePath)
	if err != nil {
		log.Error("Failed to initialize database: %v", err)
		os.Exit(1)
	}
	defer database.Close()

	log.Info("✅ Database initialized successfully")

	router := chi.NewRouter()
	
	// Create repositories
	appRepo := repository.NewSQLiteApplicationRepository(sqlDB)
	clusterRepo := repository.NewSQLiteClusterRepository(sqlDB)
	releaseRepo := repository.NewSQLiteReleaseRecordRepository(sqlDB)
	deploymentTargetRepo := repository.NewDeploymentTargetRepository(sqlDB)
	
	// Setup routes
	handlers.SetupRoutes(router, appRepo, clusterRepo, releaseRepo, deploymentTargetRepo, log)

	addr := fmt.Sprintf("%s:%d", cfg.ServerHost, cfg.ServerPort)
	log.Info("Server listening on %s", addr)

	server := &http.Server{
		Addr:    addr,
		Handler: router,
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		log.Info("Received signal: %v", sig)
		server.Close()
	}()

	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		log.Error("Server error: %v", err)
		os.Exit(1)
	}

	log.Info("Server stopped")
}
