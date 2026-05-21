// Package applications provides HTTP handlers for application management.
//
// This package handles all API endpoints related to applications:
//   - Create new applications
//   - List existing applications
//   - Retrieve application details
//   - Update application information
//   - Delete applications
//
// All handlers coordinate with ApplicationService for business logic implementation.
package applications

import (
	"encoding/json"
	"net/http"
	"strconv"

	"built-and-deploy/internal/services"
	"built-and-deploy/pkg/handlers"
	"built-and-deploy/pkg/logger"
	"built-and-deploy/pkg/responses"
)

// Create handles POST /applications request to create a new application.
//
// @Summary Create Application
// @Description Creates a new application with the provided information
// @Tags Applications
// @Accept json
// @Produce json
// @Param request body handlers.CreateApplicationRequest true "Create Application Request"
// @Success 201 {object} models.Application "Application created successfully"
// @Failure 400 {object} responses.ErrorResponse "Invalid request body"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /applications [post]
//
// Parameters:
//   - Name: Application name (required, must be unique)
//   - Repository: Docker image repository (required)
//
// Example request:
//
//	{
//	  "name": "api-service",
//	  "repository": "docker.io/myapp:v1.0.0"
//	}
//
// Example response (201):
//
//	{
//	  "id": 1,
//	  "name": "api-service",
//	  "image_name": "docker.io/myapp:v1.0.0",
//	  "created_at": "2026-04-21T10:00:00Z",
//	  "updated_at": "2026-04-21T10:00:00Z"
//	}
func Create(appService *services.ApplicationService, log *logger.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req handlers.CreateApplicationRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			responses.BadRequestResponse(w, "invalid request body")
			return
		}
		
		app, err := appService.Create(r.Context(), &req)
		if err != nil {
			responses.InternalErrorResponse(w, err.Error())
			return
		}
		responses.CreatedResponse(w, app)
	}
}

// List handles GET /applications request to retrieve all applications.
//
// @Summary List Applications
// @Description Retrieves a list of all applications in the system
// @Tags Applications
// @Produce json
// @Success 200 {array} models.Application "List of applications"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /applications [get]
//
// Returns:
//   - Array of Application objects sorted by creation time
//   - Empty array if no applications exist
//
// Example response (200):
//
//	[
//	  {
//	    "id": 1,
//	    "name": "api-service",
//	    "image_name": "docker.io/myapp:v1.0.0",
//	    "created_at": "2026-04-21T10:00:00Z"
//	  },
//	  {
//	    "id": 2,
//	    "name": "web-ui",
//	    "image_name": "docker.io/web:v1.0.0",
//	    "created_at": "2026-04-21T10:05:00Z"
//	  }
//	]
func List(appService *services.ApplicationService, log *logger.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get pagination parameters
		page := 1
		pageSize := 10
		
		if p := r.URL.Query().Get("page"); p != "" {
			if parsedP, err := strconv.Atoi(p); err == nil && parsedP > 0 {
				page = parsedP
			}
		}
		if ps := r.URL.Query().Get("pageSize"); ps != "" {
			if parsedPS, err := strconv.Atoi(ps); err == nil && parsedPS > 0 {
				pageSize = parsedPS
			}
		}
		
		offset := (page - 1) * pageSize
		apps, total, err := appService.ListApplicationsWithPagination(r.Context(), offset, pageSize)
		if err != nil {
			log.Error("Failed to list applications", "error", err)
			responses.InternalErrorResponse(w, "Failed to retrieve applications")
			return
		}
		
		totalPages := (total + pageSize - 1) / pageSize
		
		data := map[string]interface{}{
			"data":       apps,
			"page":       page,
			"pageSize":   pageSize,
			"total":      total,
			"totalPages": totalPages,
		}
		
		responses.SuccessResponse(w, data)
	}
}

// Update handles PUT /applications/{id} request to update an application.
//
// @Summary Update Application
// @Description Updates an existing application with the provided information
// @Tags Applications
// @Accept json
// @Produce json
// @Param id path int true "Application ID"
// @Param request body handlers.UpdateApplicationRequest true "Update Application Request"
// @Success 200 {object} models.Application "Application updated successfully"
// @Failure 400 {object} responses.ErrorResponse "Invalid request body"
// @Failure 404 {object} responses.ErrorResponse "Application not found"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /applications/{id} [put]
func Update(appService *services.ApplicationService, log *logger.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			responses.BadRequestResponse(w, "invalid application id")
			return
		}

		var req handlers.UpdateApplicationRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			responses.BadRequestResponse(w, "invalid request body")
			return
		}

		app, err := appService.UpdateApplication(r.Context(), id, &req)
		if err != nil {
			responses.InternalErrorResponse(w, err.Error())
			return
		}
		responses.SuccessResponse(w, app)
	}
}

// Delete handles DELETE /applications/{id} request to delete an application.
//
// @Summary Delete Application
// @Description Deletes an existing application
// @Tags Applications
// @Param id path int true "Application ID"
// @Success 200 {object} map[string]string "Application deleted successfully"
// @Failure 400 {object} responses.ErrorResponse "Invalid request"
// @Failure 404 {object} responses.ErrorResponse "Application not found"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /applications/{id} [delete]
func Delete(appService *services.ApplicationService, log *logger.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			responses.BadRequestResponse(w, "invalid application id")
			return
		}

		err = appService.DeleteApplication(r.Context(), id)
		if err != nil {
			responses.InternalErrorResponse(w, err.Error())
			return
		}
		responses.SuccessResponse(w, map[string]string{"message": "Application deleted successfully"})
	}
}
