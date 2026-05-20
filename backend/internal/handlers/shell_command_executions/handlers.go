// Package shell_command_executions provides HTTP handlers for shell command execution history.
//
// This package handles all API endpoints related to shell command execution:
//   - Monitor shell command execution status
//   - Retrieve shell command execution output
//   - Execute shell commands directly
//
// All handlers coordinate with ShellService for accessing underlying repositories
// (ShellServerRepository, ShellCommandRepository, ShellCommandExecutionRepository).
package shell_command_executions

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"built-and-deploy/internal/models"
	"built-and-deploy/internal/services"
	"built-and-deploy/pkg/logger"
	"built-and-deploy/pkg/responses"

	"github.com/go-chi/chi/v5"
)

// ListExecutions handles GET /shell-command-executions request to retrieve all shell command executions.
//
// @Summary List Shell Command Executions
// @Description Retrieves a list of all shell command executions with pagination support
// @Tags ShellCommandExecutions
// @Produce json
// @Param page query int false "Page number (default: 1)"
// @Param pageSize query int false "Page size (default: 10)"
// @Param commandID query int false "Filter by command ID"
// @Success 200 {object} PaginatedResponse "List of shell command executions retrieved successfully"
// @Failure 400 {object} responses.ErrorResponse "Invalid parameters"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /shell-command-executions [get]
func ListExecutions(shellService *services.ShellService, log *logger.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get pagination parameters
		pageStr := r.URL.Query().Get("page")
		pageSizeStr := r.URL.Query().Get("pageSize")
		commandIDStr := r.URL.Query().Get("commandID")

		page := 1
		pageSize := 10
		commandID := 0

		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
		if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 {
			pageSize = ps
		}
		if cid, err := strconv.Atoi(commandIDStr); err == nil && cid > 0 {
			commandID = cid
		}

		ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
		defer cancel()

		offset := (page - 1) * pageSize
		execRepo := shellService.ShellCommandExecutionRepo()

		var executions []*models.ShellCommandExecution
		var total int
		var err error

		// Apply filters based on query parameters
		if commandID > 0 {
			executions, total, err = execRepo.ListByCommand(ctx, commandID, offset, pageSize)
		} else {
			executions, total, err = execRepo.List(ctx, offset, pageSize)
		}

		if err != nil {
			log.Error("Failed to list shell command executions", "error", err)
			responses.InternalErrorResponse(w, "Failed to retrieve executions")
			return
		}

		totalPages := (total + pageSize - 1) / pageSize

		data := map[string]interface{}{
			"data":       executions,
			"page":       page,
			"pageSize":   pageSize,
			"total":      total,
			"totalPages": totalPages,
		}

		responses.SuccessResponse(w, data)
	}
}

// GetExecution retrieves a single shell command execution by ID
//
// @Summary Get shell command execution
// @Description Get details of a single shell command execution
// @Tags Shell Command Executions
// @Param id path int true "Execution ID"
// @Success 200 {object} responses.SuccessResponseData[models.ShellCommandExecution]
// @Failure 400 {object} responses.ErrorResponse "Invalid execution ID"
// @Failure 404 {object} responses.ErrorResponse "Execution not found"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /shell-command-executions/{id} [get]
func GetExecution(shellService *services.ShellService, log *logger.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get ID from URL parameters
		idStr := chi.URLParam(r, "id")
		if idStr == "" {
			log.Error("Invalid execution ID parameter")
			responses.BadRequestResponse(w, "Invalid execution ID")
			return
		}

		id, err := strconv.Atoi(idStr)
		if err != nil || id <= 0 {
			log.Error("Invalid execution ID format", "id", idStr, "error", err)
			responses.BadRequestResponse(w, "Invalid execution ID")
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
		defer cancel()

		executionRepo := shellService.ShellCommandExecutionRepo()
		execution, err := executionRepo.GetByID(ctx, id)
		if err != nil {
			log.Error("Failed to get shell command execution", "id", id, "error", err)
			responses.NotFoundResponse(w, "Execution not found")
			return
		}

		if execution == nil {
			log.Warn("Shell command execution not found", "id", id)
			responses.NotFoundResponse(w, "Execution not found")
			return
		}

		responses.SuccessResponse(w, execution)
	}
}
