package shell_servers

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

// ListShellServersHandler handles GET /shell-servers request
func ListShellServersHandler(serverRepo repository.ShellServerRepository, log *logger.Logger) http.HandlerFunc {
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
		servers, total, err := serverRepo.List(ctx, offset, pageSize)
		if err != nil {
			log.Error("Failed to list shell servers", "error", err)
			responses.InternalErrorResponse(w, "Failed to retrieve shell servers")
			return
		}

		totalPages := (total + pageSize - 1) / pageSize

		data := map[string]interface{}{
			"data":       servers,
			"page":       page,
			"pageSize":   pageSize,
			"total":      total,
			"totalPages": totalPages,
		}

		responses.SuccessResponse(w, data)
	}
}

// GetShellServerHandler handles GET /shell-servers/{id} request
func GetShellServerHandler(serverRepo repository.ShellServerRepository, log *logger.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(chi.URLParam(r, "id"))
		if err != nil {
			responses.BadRequestResponse(w, "Invalid server ID")
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
		defer cancel()

		server, err := serverRepo.GetByID(ctx, id)
		if err != nil {
			log.Error("Failed to get shell server", "id", id, "error", err)
			responses.NotFoundResponse(w, "Shell server not found")
			return
		}

		responses.SuccessResponse(w, server)
	}
}

// CreateShellServerHandler handles POST /shell-servers request
func CreateShellServerHandler(serverRepo repository.ShellServerRepository, log *logger.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Name       string `json:"name"`
			Host       string `json:"host"`
			Port       int    `json:"port"`
			Username   string `json:"username"`
			AuthType   string `json:"auth_type"`
			Password   string `json:"password"`
			PrivateKey string `json:"private_key"`
			Status     string `json:"status"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			responses.BadRequestResponse(w, "Invalid request body")
			return
		}

		// Validate required fields
		if req.Name == "" || req.Host == "" || req.Port == 0 || req.Username == "" {
			responses.BadRequestResponse(w, "Missing required fields: name, host, port, username")
			return
		}

		server := &models.ShellServer{
			Name:       req.Name,
			Host:       req.Host,
			Port:       req.Port,
			Username:   req.Username,
			AuthType:   req.AuthType,
			Password:   req.Password,
			PrivateKey: req.PrivateKey,
			Status:     "inactive",
		}

		if req.Status != "" {
			server.Status = req.Status
		}

		ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
		defer cancel()

		if err := serverRepo.Create(ctx, server); err != nil {
			log.Error("Failed to create shell server", "error", err)
			responses.InternalErrorResponse(w, "Failed to create shell server")
			return
		}

		responses.CreatedResponse(w, server)
	}
}

// UpdateShellServerHandler handles PUT /shell-servers/{id} request
func UpdateShellServerHandler(serverRepo repository.ShellServerRepository, log *logger.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(chi.URLParam(r, "id"))
		if err != nil {
			responses.BadRequestResponse(w, "Invalid server ID")
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
		defer cancel()

		server, err := serverRepo.GetByID(ctx, id)
		if err != nil {
			responses.NotFoundResponse(w, "Shell server not found")
			return
		}

		var req struct {
			Name       string `json:"name"`
			Host       string `json:"host"`
			Port       int    `json:"port"`
			Username   string `json:"username"`
			AuthType   string `json:"auth_type"`
			Password   string `json:"password"`
			PrivateKey string `json:"private_key"`
			Status     string `json:"status"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			responses.BadRequestResponse(w, "Invalid request body")
			return
		}

		// Update fields if provided
		if req.Name != "" {
			server.Name = req.Name
		}
		if req.Host != "" {
			server.Host = req.Host
		}
		if req.Port != 0 {
			server.Port = req.Port
		}
		if req.Username != "" {
			server.Username = req.Username
		}
		if req.AuthType != "" {
			server.AuthType = req.AuthType
		}
		if req.Password != "" {
			server.Password = req.Password
		}
		if req.PrivateKey != "" {
			server.PrivateKey = req.PrivateKey
		}
		if req.Status != "" {
			server.Status = req.Status
		}

		if err := serverRepo.Update(ctx, server); err != nil {
			log.Error("Failed to update shell server", "id", id, "error", err)
			responses.InternalErrorResponse(w, "Failed to update shell server")
			return
		}

		responses.SuccessResponse(w, server)
	}
}

// DeleteShellServerHandler handles DELETE /shell-servers/{id} request
func DeleteShellServerHandler(serverRepo repository.ShellServerRepository, log *logger.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(chi.URLParam(r, "id"))
		if err != nil {
			responses.BadRequestResponse(w, "Invalid server ID")
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
		defer cancel()

		if err := serverRepo.Delete(ctx, id); err != nil {
			log.Error("Failed to delete shell server", "id", id, "error", err)
			responses.InternalErrorResponse(w, "Failed to delete shell server")
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
