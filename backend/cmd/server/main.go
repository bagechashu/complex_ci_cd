// Package main provides the Release Control System server.
//
// The Release Control System is a comprehensive deployment management platform that
// enables teams to safely and systematically release applications across multiple
// environments and infrastructure platforms (Kubernetes, SSH, Ansible, Salt).
//
// Key features:
//   - Multi-environment deployment orchestration (Dev → Staging → Production)
//   - Support for multiple deployment targets (Kubernetes, SSH, Ansible, Salt)
//   - Comprehensive release event tracking and audit logging
//   - Shell command execution with approval workflows
//   - Workload-based deployment targeting
//   - Cluster and application configuration management
//
// Architecture:
//   - REST API endpoints for all operations
//   - SQLite database backend with encryption support
//   - Service layer for business logic
//   - Repository pattern for data access
//   - Context-based timeout and cancellation support
//
// See:
//   - Handler documentation in internal/handlers/
//   - Service documentation in internal/services/
//   - Repository documentation in internal/repository/
//   - Model documentation in internal/models/
package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"built-and-deploy/internal/config"
	"built-and-deploy/internal/database"
	"built-and-deploy/internal/repository"
	"built-and-deploy/internal/routes"
	"built-and-deploy/internal/services"
	"built-and-deploy/pkg/logger"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		// Use fmt before logger is initialized
		// Since LogConfig failed, we need to exit immediately
		os.Exit(1)
	}

	// Initialize global logger
	logger.InitLogger()
	log := logger.GetLogger()

	log.Info("Starting Release Control Server", "environment", cfg.Environment)

	// Initialize database connection
	db, err := database.Init(cfg.DatabasePath)
	if err != nil {
		log.Error("failed to initialize database", "error", err)
		os.Exit(1)
	}
	defer database.Close(db)

	encryptionKey := cfg.EncryptionKey

	// Repositories
	appRepo := repository.NewSQLiteApplicationRepository(db)
	clusterRepo := repository.NewSQLiteClusterRepository(db, encryptionKey)
	releaseRepo := repository.NewSQLiteReleaseRecordRepository(db)
	workloadRepo := repository.NewWorkloadTargetRepository(db)
	envRepo := repository.NewEnvironmentRepository(db)
	eventRepo := repository.NewSQLiteReleaseEventRepository(db)
	shellServerRepo := repository.NewSQLiteShellServerRepository(db, encryptionKey)
	shellCommandRepo := repository.NewSQLiteShellCommandRepository(db)
	shellTaskExecutionRepo := repository.NewSQLiteShellTaskExecutionRepository(db)
	shellTaskRepo := repository.NewSQLiteShellTaskRepository(db)
	appClusterConfigRepo := repository.NewSQLiteApplicationClusterConfigRepository(db)

	// Service Container with functional options
	container, err := services.NewServiceContainer(
		db, log,
		services.WithApplicationRepository(appRepo),
		services.WithClusterRepository(clusterRepo),
		services.WithReleaseRepository(releaseRepo),
		services.WithWorkloadRepository(workloadRepo),
		services.WithEnvironmentRepository(envRepo),
		services.WithEventRepository(eventRepo),
		services.WithShellServerRepository(shellServerRepo),
		services.WithShellCommandRepository(shellCommandRepo),
		services.WithShellTaskExecutionRepository(shellTaskExecutionRepo),
		services.WithShellTaskRepository(shellTaskRepo),
		services.WithApplicationClusterConfigRepository(appClusterConfigRepo),
	)
	if err != nil {
		log.Error("failed to create service container", "error", err)
		os.Exit(1)
	}

	// Create router
	router := routes.NewRouter(container, log)

	// Create HTTP server with timeouts
	addr := cfg.ServerAddr()
	server := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown channel
	done := make(chan struct{})

	// Listen for SIGTERM and SIGINT signals
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)
		<-sigChan

		log.Info("shutdown signal received, starting graceful shutdown")

		// Give requests 30 seconds to complete
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			log.Error("server shutdown error", "error", err)
		}
		close(done)
	}()

	// Start server
	log.Info("server listening", "address", addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Error("server error", "error", err)
		os.Exit(1)
	}

	// Wait for graceful shutdown to complete
	<-done
	log.Info("server stopped")
}
