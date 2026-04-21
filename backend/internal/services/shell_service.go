package services

import (
	"built-and-deploy/internal/repository"
	"built-and-deploy/pkg/logger"
)

// ShellService manages shell command execution operations across target servers.
//
// ShellService provides access to shell execution infrastructure:
//   - Shell task definition and management
//   - Shell server registration and management
//   - Command template creation and storage
//   - Execution task tracking and monitoring
//   - Sensitive credential encryption (SSH credentials, API keys)
//
// The service works with four main repository types:
//   - ShellTaskRepository: Manages task definitions
//   - ShellServerRepository: Manages target servers for command execution
//   - ShellCommandRepository: Manages reusable command templates
//   - ShellTaskExecutionRepository: Tracks individual command executions
//
// Security:
//   - All sensitive credentials are encrypted at rest
//   - Encryption key is provided at service initialization
//   - Decryption happens on-demand for actual execution
//
// Usage:
//
//	service := NewShellService(taskRepo, serverRepo, commandRepo, execRepo, encryptionKey, log)
//	taskRepo := service.ShellTaskRepo()
//	serverRepo := service.ShellServerRepo()
//	commandRepo := service.ShellCommandRepo()
//	execRepo := service.ShellTaskExecutionRepo()
type ShellService struct {
	taskRepo      repository.ShellTaskRepository
	serverRepo    repository.ShellServerRepository
	commandRepo   repository.ShellCommandRepository
	execRepo      repository.ShellTaskExecutionRepository
	encryptionKey string
	log           *logger.Logger
}

// NewShellService creates a new ShellService instance.
//
// Parameters:
//   - taskRepo: ShellTaskRepository for task definition management
//   - serverRepo: ShellServerRepository for server management
//   - commandRepo: ShellCommandRepository for command templates
//   - execRepo: ShellTaskExecutionRepository for execution tracking
//   - encryptionKey: Key for encrypting sensitive credentials
//   - log: Logger for structured logging
//
// Returns a configured ShellService ready for use.
//
// Example:
//
//	service := NewShellService(
//	    taskRepo,
//	    serverRepo,
//	    commandRepo,
//	    execRepo,
//	    config.EncryptionKey,
//	    logger.GetLogger(),
//	)
func NewShellService(
	taskRepo repository.ShellTaskRepository,
	serverRepo repository.ShellServerRepository,
	commandRepo repository.ShellCommandRepository,
	execRepo repository.ShellTaskExecutionRepository,
	encryptionKey string,
	log *logger.Logger,
) *ShellService {
	return &ShellService{
		taskRepo:      taskRepo,
		serverRepo:    serverRepo,
		commandRepo:   commandRepo,
		execRepo:      execRepo,
		encryptionKey: encryptionKey,
		log:           log,
	}
}

// ShellTaskRepo returns the underlying ShellTaskRepository.
//
// Returns:
//   - repository.ShellTaskRepository: The configured task repository
//
// Usage:
//
//	taskRepo := service.ShellTaskRepo()
//	tasks, _, err := taskRepo.List(ctx, 0, 100)
func (s *ShellService) ShellTaskRepo() repository.ShellTaskRepository {
	return s.taskRepo
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

// ShellTaskExecutionRepo returns the underlying ShellTaskExecutionRepository.
//
// Returns:
//   - repository.ShellTaskExecutionRepository: The configured execution repository
//
// Usage:
//
//	execRepo := service.ShellTaskExecutionRepo()
//	exec, err := execRepo.GetByID(ctx, executionID)
func (s *ShellService) ShellTaskExecutionRepo() repository.ShellTaskExecutionRepository {
	return s.execRepo
}
