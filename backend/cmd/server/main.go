package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"built-and-deploy/internal/config"
	"built-and-deploy/internal/database"
	"built-and-deploy/internal/handlers"
	"built-and-deploy/internal/repository"
	"built-and-deploy/pkg/logger"

	"github.com/go-chi/chi/v5"
)

// Use air for live reload during development
func main() {
	cfg := config.LoadConfig()
	
	// Initialize global logger
	logger.InitLogger()
	log := logger.GetLogger()

	log.Info("Starting Release Control Server")
	log.Info("Configuration loaded", "environment", cfg.Environment, "database", cfg.DatabasePath)

	sqlDB, err := database.Init(cfg.DatabasePath)
	if err != nil {
		log.Error("Failed to initialize database", "error", err)
		os.Exit(1)
	}
	defer database.Close()

	log.Info("Database initialized successfully")

	router := chi.NewRouter()
	
	// Create repositories
	appRepo := repository.NewSQLiteApplicationRepository(sqlDB)
	clusterRepo := repository.NewSQLiteClusterRepository(sqlDB, cfg.EncryptionKey)
	releaseRepo := repository.NewSQLiteReleaseRecordRepository(sqlDB)
	workloadTargetRepo := repository.NewWorkloadTargetRepository(sqlDB)
	
	// Shell-related repositories
	shellServerRepo := repository.NewSQLiteShellServerRepository(sqlDB, cfg.EncryptionKey)
	shellCommandRepo := repository.NewSQLiteShellCommandRepository(sqlDB)
	shellTaskRepo := repository.NewSQLiteShellTaskRepository(sqlDB)
	shellExecutionRepo := repository.NewSQLiteShellTaskExecutionRepository(sqlDB)
	
	// Setup routes
	handlers.SetupRoutes(router, appRepo, clusterRepo, releaseRepo, workloadTargetRepo, shellServerRepo, shellCommandRepo, shellTaskRepo, shellExecutionRepo, log)

	addr := fmt.Sprintf("%s:%d", cfg.ServerHost, cfg.ServerPort)
	log.Info("Server listening", "address", addr)

	server := &http.Server{
		Addr:    addr,
		Handler: router,
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		log.Info("Received signal", "signal", sig.String())
		server.Close()
	}()

	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		log.Error("Server error", "error", err)
		os.Exit(1)
	}

	log.Info("Server stopped")
}
