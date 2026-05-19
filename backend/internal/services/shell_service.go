package services

import (
	"built-and-deploy/internal/repository"
	"built-and-deploy/pkg/logger"
)

// ShellService manages shell command execution operations across target servers.
//
// ShellService provides access to shell execution infrastructure:
//   - Shell server registration and management
//   - Command template creation and storage
//   - Execution command tracking and monitoring
//   - Sensitive credential encryption (SSH credentials, API keys)
//
// The service works with three main repository types:
//   - ShellServerRepository: Manages target servers for command execution
//   - ShellCommandRepository: Manages reusable command templates
//   - ShellCommandExecutionRepository: Tracks individual command executions
//
// Security:
//   - All sensitive credentials are encrypted at rest
//   - Encryption key is provided at service initialization
//   - Decryption happens on-demand for actual execution
//
// Usage:
//
//	service := NewShellService(serverRepo, commandRepo, execRepo, encryptionKey, log)
//	serverRepo := service.ShellServerRepo()
//	commandRepo := service.ShellCommandRepo()
//	execRepo := service.ShellCommandExecutionRepo()
type ShellService struct {
	serverRepo    repository.ShellServerRepository
	commandRepo   repository.ShellCommandRepository
	execRepo      repository.ShellCommandExecutionRepository
	encryptionKey string
	log           *logger.Logger
}

// NewShellService creates a new ShellService instance.
//
// Parameters:
//   - serverRepo: ShellServerRepository for server management
//   - commandRepo: ShellCommandRepository for command templates
//   - execRepo: ShellCommandExecutionRepository for execution tracking
//   - encryptionKey: Key for encrypting sensitive credentials
//   - log: Logger for structured logging
//
// Returns a configured ShellService ready for use.
//
// Example:
//
//	service := NewShellService(
//	    serverRepo,
//	    commandRepo,
//	    execRepo,
//	    config.EncryptionKey,
//	    logger.GetLogger(),
//	)
func NewShellService(
	serverRepo repository.ShellServerRepository,
	commandRepo repository.ShellCommandRepository,
	execRepo repository.ShellCommandExecutionRepository,
	encryptionKey string,
	log *logger.Logger,
) *ShellService {
	return &ShellService{
		serverRepo:    serverRepo,
		commandRepo:   commandRepo,
		execRepo:      execRepo,
		encryptionKey: encryptionKey,
		log:           log,
	}
}

// ShellServerRepo returns the underlying ShellServerRepository.
//
// Returns:
//   - repository.ShellServerRepository: The configured server repository
//
// Usage:
//
//	serverRepo := service.ShellServerRepo()
//	servers, err := serverRepo.List(ctx, 0, 100)
func (s *ShellService) ShellServerRepo() repository.ShellServerRepository {
	return s.serverRepo
}

// ShellCommandRepo returns the underlying ShellCommandRepository.
//
// Returns:
//   - repository.ShellCommandRepository: The configured command repository
//
// Usage:
//
//	commandRepo := service.ShellCommandRepo()
//	cmd, err := commandRepo.GetByID(ctx, commandID)
func (s *ShellService) ShellCommandRepo() repository.ShellCommandRepository {
	return s.commandRepo
}

// ShellCommandExecutionRepo returns the underlying ShellCommandExecutionRepository.
//
// Returns:
//   - repository.ShellCommandExecutionRepository: The configured execution repository
//
// Usage:
//
//	execRepo := service.ShellCommandExecutionRepo()
//	exec, err := execRepo.GetByID(ctx, executionID)
func (s *ShellService) ShellCommandExecutionRepo() repository.ShellCommandExecutionRepository {
	return s.execRepo
}
