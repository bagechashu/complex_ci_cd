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
	"github.com/op/release-control/internal/deployers"
	"github.com/op/release-control/internal/handlers"
	"github.com/op/release-control/internal/repository"
	"github.com/op/release-control/internal/services"
	"github.com/op/release-control/pkg/logger"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()
	log := logger.NewLogger()

	log.Info("Starting Release Control Server")
	log.Info("Environment: %s", cfg.Environment)
	log.Info("Database: %s", cfg.DatabasePath)

	// Initialize database
	db, err := database.Init(cfg.DatabasePath)
	if err != nil {
		log.Error("Failed to initialize database: %v", err)
		os.Exit(1)
	}
	defer func() {
		if err := database.Close(); err != nil {
			log.Error("Failed to close database: %v", err)
		}
	}()

	// Insert initial data
	if err := database.InsertInitialData(db); err != nil {
		log.Error("Failed to insert initial data: %v", err)
	}

	// Get connection info
	info, err := database.GetConnectionInfo(db)
	if err == nil {
		log.Info("Database loaded: apps=%d, envs=%d, clusters=%d, releases=%d",
			info["applications"], info["environments"], info["clusters"], info["releases"])
	}

	// Initialize repositories
	releaseRepo := repository.NewReleaseRecordRepository(db)
	appRepo := repository.NewApplicationRepository(db)
	envRepo := repository.NewEnvironmentRepository(db)
	clusterRepo := repository.NewClusterRepository(db)
	targetRepo := repository.NewDeploymentTargetRepository(db)

	// Initialize deployer factory
	deployerFactory := deployers.NewDeployerFactory(log)

	// Initialize services
	releaseService := services.NewReleaseService(
		releaseRepo, appRepo, envRepo, clusterRepo, targetRepo,
		deployerFactory, log,
	)

	// Setup router
	router := chi.NewRouter()
	handlers.SetupRoutes(router, releaseService, appRepo, envRepo, clusterRepo, targetRepo, log)

	// Start server
	addr := fmt.Sprintf("%s:%d", cfg.ServerHost, cfg.ServerPort)
	log.Info("Server listening on %s", addr)

	server := &http.Server{
		Addr:    addr,
		Handler: router,
	}

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		log.Info("Received signal: %v", sig)
		if err := server.Close(); err != nil {
			log.Error("Error closing server: %v", err)
		}
	}()

	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		log.Error("Server error: %v", err)
		os.Exit(1)
	}

	log.Info("Server stopped")
}
