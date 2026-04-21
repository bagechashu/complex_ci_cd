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
	"encoding/json"
	"net/http"

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
		// TODO: Implement shell tasks listing
		responses.SuccessResponse(w, []interface{}{})
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
		var req map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			responses.BadRequestResponse(w, "Invalid request body")
			return
		}
		// TODO: Implement shell task creation
		responses.CreatedResponse(w, req)
	}
}
