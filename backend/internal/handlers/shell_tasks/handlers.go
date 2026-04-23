// Package shell_tasks provides HTTP handlers for shell command execution management.
//
// This package handles all API endpoints related to shell command execution:
//   - List shell tasks
//   - Create new shell tasks
//   - Retrieve shell task details
//   - Monitor shell task execution status
//   - Retrieve shell task output
//
// All handlers coordinate with ShellService for accessing underlying repositories
// (ShellServerRepository, ShellCommandRepository, ShellTaskExecutionRepository).
package shell_tasks

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"built-and-deploy/internal/models"
	"built-and-deploy/internal/services"
	"built-and-deploy/pkg/logger"
	"built-and-deploy/pkg/responses"
)

// List handles GET /shell-tasks request to retrieve all shell tasks.
//
// @Summary List Shell Tasks
// @Description Retrieves a list of all shell tasks in the system
// @Tags ShellTasks
// @Produce json
// @Success 200 {array} interface{} "List of shell tasks"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /shell-tasks [get]
//
// Returns:
//   - Array of shell execution task records
//   - Each task includes execution ID, status, start time, end time
//   - Empty array if no tasks exist
//
// Task status values:
//   - "pending": Task queued but not yet started
//   - "running": Task currently executing
//   - "completed": Task finished successfully
//   - "failed": Task failed with error
//   - "cancelled": Task was cancelled before completion
//
// Example response (200):
//
//	[
//	  {
//	    "id": 1,
//	    "task_id": "task-123",
//	    "server_id": 1,
//	    "command_id": 5,
//	    "status": "completed",
//	    "started_at": "2026-04-21T10:00:00Z",
//	    "completed_at": "2026-04-21T10:00:05Z"
//	  },
//	  {
//	    "id": 2,
//	    "task_id": "task-124",
//	    "server_id": 2,
//	    "command_id": 6,
//	    "status": "running",
//	    "started_at": "2026-04-21T10:01:00Z"
//	  }
//	]
func List(shellService *services.ShellService, log *logger.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get pagination parameters
		pageStr := r.URL.Query().Get("page")
		pageSizeStr := r.URL.Query().Get("pageSize")

		page := 1
		pageSize := 10

		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
		if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 {
			pageSize = ps
		}

		ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
		defer cancel()

		offset := (page - 1) * pageSize
		taskRepo := shellService.ShellTaskRepo()
		tasks, total, err := taskRepo.List(ctx, offset, pageSize)
		if err != nil {
			log.Error("Failed to list shell tasks", "error", err)
			responses.InternalErrorResponse(w, "Failed to retrieve shell tasks")
			return
		}

		totalPages := (total + pageSize - 1) / pageSize

		resp := map[string]interface{}{
			"code":       0,
			"message":    "success",
			"data":       tasks,
			"page":       page,
			"pageSize":   pageSize,
			"total":      total,
			"totalPages": totalPages,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}

// Create handles POST /shell-tasks request to create a new shell task.
//
// @Summary Create Shell Task
// @Description Creates a new shell command execution task on a target server
// @Tags ShellTasks
// @Accept json
// @Produce json
// @Param request body CreateShellTaskRequest true "Create Shell Task Request"
// @Success 201 {object} interface{} "Shell task created successfully"
// @Failure 400 {object} responses.ErrorResponse "Invalid request body or missing required fields"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /shell-tasks [post]
//
// Request body:
//
//	{
//	  "server_id": 1,
//	  "command": "ls -la /var/log",
//	  "timeout": 30,
//	  "environment": {
//	    "PATH": "/usr/local/bin:/usr/bin"
//	  }
//	}
//
// Response (201):
//
//	{
//	  "id": 1,
//	  "task_id": "task-125",
//	  "server_id": 1,
//	  "status": "pending",
//	  "created_at": "2026-04-21T10:02:00Z"
//	}
//
// Notes:
//   - Task creation is asynchronous - task will execute in background
//   - Use /shell-tasks/{id}/status endpoint to monitor execution
//   - Command timeout is in seconds (default 30s if not specified)
//   - Sensitive command output may be redacted from logs
func Create(shellService *services.ShellService, log *logger.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Name             string `json:"name"`
			Description      string `json:"description"`
			CommandID        int    `json:"command_id"`
			ServerIDs        []int  `json:"server_ids"`
			ExecutionMethod  string `json:"execution_method"`
			RequiresApproval bool   `json:"requires_approval"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			responses.BadRequestResponse(w, "Invalid request body")
			return
		}

		// Validate required fields
		if req.Name == "" || req.CommandID == 0 || len(req.ServerIDs) == 0 {
			responses.BadRequestResponse(w, "Missing required fields: name, command_id, server_ids")
			return
		}

		// Set default execution method
		if req.ExecutionMethod == "" {
			req.ExecutionMethod = "serial"
		}

		task := &models.ShellTask{
			Name:             req.Name,
			Description:      req.Description,
			CommandID:        req.CommandID,
			ServerIDs:        req.ServerIDs,
			ExecutionMethod:  req.ExecutionMethod,
			RequiresApproval: req.RequiresApproval,
		}

		// Validate model
		if err := task.Validate(); err != nil {
			responses.BadRequestResponse(w, err.Error())
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
		defer cancel()

		taskRepo := shellService.ShellTaskRepo()
		if err := taskRepo.Create(ctx, task); err != nil {
			log.Error("Failed to create shell task", "error", err)
			responses.InternalErrorResponse(w, "Failed to create shell task")
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		resp := map[string]interface{}{
			"code":    0,
			"message": "created",
			"data":    task,
		}
		json.NewEncoder(w).Encode(resp)
	}
}

// Get handles GET /shell-tasks/{id} request to retrieve a specific shell task.
func Get(shellService *services.ShellService, log *logger.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			responses.BadRequestResponse(w, "Invalid task ID")
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
		defer cancel()

		taskRepo := shellService.ShellTaskRepo()
		task, err := taskRepo.GetByID(ctx, id)
		if err != nil {
			log.Error("Failed to get shell task", "id", id, "error", err)
			responses.NotFoundResponse(w, "Shell task not found")
			return
		}

		w.Header().Set("Content-Type", "application/json")
		resp := map[string]interface{}{
			"code":    0,
			"message": "success",
			"data":    task,
		}
		json.NewEncoder(w).Encode(resp)
	}
}

// Update handles PUT /shell-tasks/{id} request to update a shell task.
func Update(shellService *services.ShellService, log *logger.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			responses.BadRequestResponse(w, "Invalid task ID")
			return
		}

		var req struct {
			Name             string `json:"name"`
			Description      string `json:"description"`
			CommandID        int    `json:"command_id"`
			ServerIDs        []int  `json:"server_ids"`
			ExecutionMethod  string `json:"execution_method"`
			RequiresApproval bool   `json:"requires_approval"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			responses.BadRequestResponse(w, "Invalid request body")
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
		defer cancel()

		taskRepo := shellService.ShellTaskRepo()
		task, err := taskRepo.GetByID(ctx, id)
		if err != nil {
			responses.NotFoundResponse(w, "Shell task not found")
			return
		}

		// Update fields if provided
		if req.Name != "" {
			task.Name = req.Name
		}
		if req.Description != "" {
			task.Description = req.Description
		}
		if req.CommandID > 0 {
			task.CommandID = req.CommandID
		}
		if len(req.ServerIDs) > 0 {
			task.ServerIDs = req.ServerIDs
		}
		if req.ExecutionMethod != "" {
			task.ExecutionMethod = req.ExecutionMethod
		}
		task.RequiresApproval = req.RequiresApproval

		// Validate model
		if err := task.Validate(); err != nil {
			responses.BadRequestResponse(w, err.Error())
			return
		}

		if err := taskRepo.Update(ctx, task); err != nil {
			log.Error("Failed to update shell task", "id", id, "error", err)
			responses.InternalErrorResponse(w, "Failed to update shell task")
			return
		}

		w.Header().Set("Content-Type", "application/json")
		resp := map[string]interface{}{
			"code":    0,
			"message": "success",
			"data":    task,
		}
		json.NewEncoder(w).Encode(resp)
	}
}

// Delete handles DELETE /shell-tasks/{id} request to delete a shell task.
func Delete(shellService *services.ShellService, log *logger.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			responses.BadRequestResponse(w, "Invalid task ID")
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
		defer cancel()

		taskRepo := shellService.ShellTaskRepo()
		if err := taskRepo.Delete(ctx, id); err != nil {
			log.Error("Failed to delete shell task", "id", id, "error", err)
			responses.InternalErrorResponse(w, "Failed to delete shell task")
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
