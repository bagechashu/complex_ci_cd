package shell_commands

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"built-and-deploy/internal/models"
	"built-and-deploy/internal/repository"
	"built-and-deploy/pkg/logger"
	"built-and-deploy/pkg/responses"

	"github.com/go-chi/chi/v5"
)

// ListShellCommandsHandler handles GET /shell-commands request
func ListShellCommandsHandler(commandRepo repository.ShellCommandRepository, log *logger.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get pagination parameters
		pageStr := r.URL.Query().Get("page")
		pageSizeStr := r.URL.Query().Get("pageSize")
		serverIDStr := r.URL.Query().Get("serverID")

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
		commands, total, err := commandRepo.List(ctx, offset, pageSize)
		if err != nil {
			log.Error("Failed to list shell commands", "error", err)
			responses.InternalErrorResponse(w, "Failed to retrieve shell commands")
			return
		}

		// Filter by server if provided
		if serverIDStr != "" {
			if serverID, err := strconv.Atoi(serverIDStr); err == nil {
				filtered := []*models.ShellCommand{}
				for _, cmd := range commands {
					if cmd.ServerID == serverID {
						filtered = append(filtered, cmd)
					}
				}
				commands = filtered
			}
		}

		totalPages := (total + pageSize - 1) / pageSize

		data := map[string]interface{}{
			"data":       commands,
			"page":       page,
			"pageSize":   pageSize,
			"total":      total,
			"totalPages": totalPages,
		}

		responses.SuccessResponse(w, data)
	}
}

// GetShellCommandHandler handles GET /shell-commands/{id} request
func GetShellCommandHandler(commandRepo repository.ShellCommandRepository, log *logger.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(chi.URLParam(r, "id"))
		if err != nil {
			responses.BadRequestResponse(w, "Invalid command ID")
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
		defer cancel()

		command, err := commandRepo.GetByID(ctx, id)
		if err != nil {
			log.Error("Failed to get shell command", "id", id, "error", err)
			responses.NotFoundResponse(w, "Shell command not found")
			return
		}

		responses.SuccessResponse(w, command)
	}
}

// CreateShellCommandHandler handles POST /shell-commands request
func CreateShellCommandHandler(commandRepo repository.ShellCommandRepository, log *logger.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			ServerID    int    `json:"server_id"`
			Command     string `json:"command"`
			Description string `json:"description"`

			IsPublished bool `json:"is_published"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			responses.BadRequestResponse(w, "Invalid request body")
			return
		}

		// Validate required fields
		if req.ServerID == 0 || req.Command == "" {
			responses.BadRequestResponse(w, "Missing required fields: server_id, command")
			return
		}

		command := &models.ShellCommand{
			ServerID:    req.ServerID,
			Command:     req.Command,
			Description: req.Description,
			IsPublished: req.IsPublished,
		}

		ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
		defer cancel()

		if err := commandRepo.Create(ctx, command); err != nil {
			log.Error("Failed to create shell command", "error", err)
			responses.InternalErrorResponse(w, "Failed to create shell command")
			return
		}

		responses.CreatedResponse(w, command)
	}
}

// PublishShellCommandHandler handles POST /shell-commands/{id}/publish request
func PublishShellCommandHandler(commandRepo repository.ShellCommandRepository, log *logger.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(chi.URLParam(r, "id"))
		if err != nil {
			responses.BadRequestResponse(w, "Invalid command ID")
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
		defer cancel()

		command, err := commandRepo.GetByID(ctx, id)
		if err != nil {
			responses.NotFoundResponse(w, "Shell command not found")
			return
		}

		command.IsPublished = true
		if err := commandRepo.Update(ctx, command); err != nil {
			log.Error("Failed to publish shell command", "id", id, "error", err)
			responses.InternalErrorResponse(w, "Failed to publish shell command")
			return
		}

		responses.SuccessResponse(w, command)
	}
}

// UnpublishShellCommandHandler handles POST /shell-commands/{id}/unpublish request
func UnpublishShellCommandHandler(commandRepo repository.ShellCommandRepository, log *logger.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(chi.URLParam(r, "id"))
		if err != nil {
			responses.BadRequestResponse(w, "Invalid command ID")
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
		defer cancel()

		command, err := commandRepo.GetByID(ctx, id)
		if err != nil {
			responses.NotFoundResponse(w, "Shell command not found")
			return
		}

		command.IsPublished = false
		if err := commandRepo.Update(ctx, command); err != nil {
			log.Error("Failed to unpublish shell command", "id", id, "error", err)
			responses.InternalErrorResponse(w, "Failed to unpublish shell command")
			return
		}

		resp := map[string]interface{}{
			"code":    0,
			"message": "success",
			"data":    command,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}

// DeleteShellCommandHandler handles DELETE /shell-commands/{id} request
func DeleteShellCommandHandler(commandRepo repository.ShellCommandRepository, log *logger.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(chi.URLParam(r, "id"))
		if err != nil {
			responses.BadRequestResponse(w, "Invalid command ID")
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
		defer cancel()

		if err := commandRepo.Delete(ctx, id); err != nil {
			log.Error("Failed to delete shell command", "id", id, "error", err)
			responses.InternalErrorResponse(w, "Failed to delete shell command")
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

// UpdateShellCommandHandler handles PUT /shell-commands/{id} request
func UpdateShellCommandHandler(commandRepo repository.ShellCommandRepository, log *logger.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(chi.URLParam(r, "id"))
		if err != nil {
			responses.BadRequestResponse(w, "Invalid command ID")
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
		defer cancel()

		command, err := commandRepo.GetByID(ctx, id)
		if err != nil {
			responses.NotFoundResponse(w, "Shell command not found")
			return
		}

		var req struct {
			Command     string `json:"command"`
			Description string `json:"description"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			responses.BadRequestResponse(w, "Invalid request body")
			return
		}

		if req.Command != "" {
			command.Command = req.Command
		}
		if req.Description != "" {
			command.Description = req.Description
		}

		if err := commandRepo.Update(ctx, command); err != nil {
			log.Error("Failed to update shell command", "id", id, "error", err)
			responses.InternalErrorResponse(w, "Failed to update shell command")
			return
		}

		resp := map[string]interface{}{
			"code":    0,
			"message": "success",
			"data":    command,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}
