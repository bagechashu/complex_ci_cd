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
		apps, err := appService.ListApplications(r.Context(), 0, 100)
		if err != nil {
			responses.InternalErrorResponse(w, err.Error())
			return
		}
		responses.SuccessResponse(w, apps)
	}
}
