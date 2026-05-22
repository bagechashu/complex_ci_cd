// Package releases provides HTTP handlers for release management and deployment orchestration.
//
// This package handles all API endpoints related to application releases:
//   - Create and execute new releases/deployments
//   - Retrieve release information and status
//   - List release history and events
//   - Perform rollbacks to previous releases
//
// All handlers coordinate with ReleaseService for business logic implementation.
package releases

import (
	"encoding/json"
	"net/http"
	"strconv"

	"built-and-deploy/internal/models"
	"built-and-deploy/internal/services"
	"built-and-deploy/pkg/handlers"
	"built-and-deploy/pkg/logger"
	"built-and-deploy/pkg/responses"
)

// Create handles POST /releases request to create and execute a new release.
//
// @Summary Create and Execute Release
// @Description Creates a release record and executes deployment to specified cluster
// @Tags Releases
// @Accept json
// @Produce json
// @Param request body handlers.CreateReleaseRequest true "Create Release Request"
// @Success 201 {object} models.ReleaseRecord "Release created and initiated successfully"
// @Failure 400 {object} responses.ErrorResponse "Invalid request body or parameters"
// @Failure 404 {object} responses.ErrorResponse "Application or cluster not found"
// @Failure 500 {object} responses.ErrorResponse "Internal server error or deployment failed"
// @Router /releases [post]
//
// Parameters:
//   - ApplicationID: Application to deploy (required)
//   - ClusterID: Target cluster for deployment (required)
//   - ImageName: Docker image name/tag to deploy (required)
//
// Example request:
//
//	{
//	  "application_id": 1,
//	  "cluster_id": 1,
//	  "image_name": "docker.io/myapp:v2.1.0"
//	}
//
// Example response (201):
//
//	{
//	  "id": 1,
//	  "application_id": 1,
//	  "cluster_id": 1,
//	  "image_name": "docker.io/myapp:v2.1.0",
//	  "status": "in_progress",
//	  "created_at": "2026-04-21T10:00:00Z",
//	  "updated_at": "2026-04-21T10:00:00Z"
//	}
func Create(releaseService *services.ReleaseService, log *logger.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req handlers.CreateReleaseRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			responses.BadRequestResponse(w, "invalid request body")
			return
		}

		release, err := releaseService.ReleaseWithRequest(r.Context(), &req)
		if err != nil {
			responses.InternalErrorResponse(w, err.Error())
			return
		}
		responses.AcceptedResponse(w, "release accepted", release)
	}
}

// Get handles GET /releases/{id} request to retrieve release details and status.
//
// @Summary Get Release Details
// @Description Retrieves detailed information about a specific release including current status
// @Tags Releases
// @Produce json
// @Param id path int true "Release ID"
// @Success 200 {object} models.ReleaseRecord "Release details retrieved successfully"
// @Failure 400 {object} responses.ErrorResponse "Invalid release ID format"
// @Failure 404 {object} responses.ErrorResponse "Release not found"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /releases/{id} [get]
//
// Example response (200):
//
//	{
//	  "id": 1,
//	  "application_id": 1,
//	  "cluster_id": 1,
//	  "image_name": "docker.io/myapp:v2.1.0",
//	  "status": "success",
//	  "created_at": "2026-04-21T10:00:00Z",
//	  "updated_at": "2026-04-21T10:05:00Z"
//	}
func Get(releaseService *services.ReleaseService, log *logger.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		if idStr == "" {
			responses.BadRequestResponse(w, "release id is required")
			return
		}

		id, err := strconv.Atoi(idStr)
		if err != nil {
			responses.BadRequestResponse(w, "invalid release id format")
			return
		}

		release, err := releaseService.GetReleaseStatus(r.Context(), id)
		if err != nil {
			responses.InternalErrorResponse(w, err.Error())
			return
		}

		if release == nil {
			responses.NotFoundResponse(w, "release not found")
			return
		}

		responses.SuccessResponse(w, release)
	}
}

// List handles GET /releases request to retrieve release history.
//
// @Summary List Releases
// @Description Retrieves a list of releases with pagination support
// @Tags Releases
// @Produce json
// @Param offset query int false "Offset for pagination (default: 0)"
// @Param limit query int false "Limit for pagination (default: 20)"
// @Success 200 {array} models.ReleaseRecord "List of releases retrieved successfully"
// @Failure 400 {object} responses.ErrorResponse "Invalid pagination parameters"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /releases [get]
//
// Example response (200):
//
//	[
//	  {
//	    "id": 2,
//	    "application_id": 1,
//	    "cluster_id": 1,
//	    "image_name": "docker.io/myapp:v2.1.0",
//	    "status": "success",
//	    "created_at": "2026-04-21T10:00:00Z"
//	  },
//	  {
//	    "id": 1,
//	    "application_id": 1,
//	    "cluster_id": 1,
//	    "image_name": "docker.io/myapp:v2.0.0",
//	    "status": "success",
//	    "created_at": "2026-04-20T14:30:00Z"
//	  }
//	]
func List(releaseService *services.ReleaseService, log *logger.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		offset := 0
		limit := 20

		if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
			if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
				offset = o
			}
		}

		if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
			if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
				limit = l
			}
		}

		// Optional app_id filter
		appIDStr := r.URL.Query().Get("app_id")

		var releases []*models.ReleaseRecord
		var err error

		if appIDStr != "" {
			// Filter by application ID
			appID, err := strconv.Atoi(appIDStr)
			if err != nil {
				responses.BadRequestResponse(w, "invalid app_id format")
				return
			}
			releases, err = releaseService.GetReleaseHistoryByApp(r.Context(), appID, offset, limit)
		} else {
			// Get all releases
			releases, err = releaseService.GetReleaseHistory(r.Context(), offset, limit)
		}

		if err != nil {
			responses.InternalErrorResponse(w, err.Error())
			return
		}

		if releases == nil {
			releases = make([]*models.ReleaseRecord, 0)
		}
		responses.SuccessResponse(w, releases)
	}
}

// ListEvents handles GET /releases/{id}/events request to retrieve release event audit trail.
//
// @Summary List Release Events
// @Description Retrieves all events (audit trail) for a specific release deployment
// @Tags Releases
// @Produce json
// @Param id path int true "Release ID"
// @Success 200 {array} models.ReleaseEvent "List of release events retrieved successfully"
// @Failure 400 {object} responses.ErrorResponse "Invalid release ID format"
// @Failure 404 {object} responses.ErrorResponse "Release not found"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /releases/{id}/events [get]
//
// Example response (200):
//
//	[
//	  {
//	    "id": 1,
//	    "release_id": 1,
//	    "type": "DEPLOYING",
//	    "message": "Starting deployment to production cluster",
//	    "details": "{\"cluster\":\"prod\"}",
//	    "created_at": "2026-04-21T10:00:00Z"
//	  },
//	  {
//	    "id": 2,
//	    "release_id": 1,
//	    "type": "COMPLETED",
//	    "message": "Deployment completed successfully",
//	    "details": "{}",
//	    "created_at": "2026-04-21T10:05:00Z"
//	  }
//	]
func ListEvents(releaseService *services.ReleaseService, log *logger.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		if idStr == "" {
			responses.BadRequestResponse(w, "release id is required")
			return
		}

		id, err := strconv.Atoi(idStr)
		if err != nil {
			responses.BadRequestResponse(w, "invalid release id format")
			return
		}

		events, err := releaseService.ListReleaseEvents(r.Context(), id)
		if err != nil {
			responses.InternalErrorResponse(w, err.Error())
			return
		}

		if events == nil {
			events = make([]interface{}, 0)
		}
		responses.SuccessResponse(w, events)
	}
}

// Rollback handles POST /releases/{id}/rollback request to perform release rollback.
//
// @Summary Rollback Release
// @Description Performs a rollback of a failed release to the previous version
// @Tags Releases
// @Accept json
// @Produce json
// @Param id path int true "Release ID to rollback"
// @Success 200 {object} models.ReleaseRecord "Rollback initiated successfully"
// @Failure 400 {object} responses.ErrorResponse "Invalid release ID format"
// @Failure 404 {object} responses.ErrorResponse "Release not found"
// @Failure 500 {object} responses.ErrorResponse "Internal server error or rollback failed"
// @Router /releases/{id}/rollback [post]
//
// Example response (200):
//
//	{
//	  "id": 1,
//	  "application_id": 1,
//	  "cluster_id": 1,
//	  "image_name": "docker.io/myapp:v2.0.0",
//	  "status": "rolling_back",
//	  "created_at": "2026-04-21T10:00:00Z",
//	  "updated_at": "2026-04-21T10:10:00Z"
//	}
func Rollback(releaseService *services.ReleaseService, log *logger.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		if idStr == "" {
			responses.BadRequestResponse(w, "release id is required")
			return
		}

		id, err := strconv.Atoi(idStr)
		if err != nil {
			responses.BadRequestResponse(w, "invalid release id format")
			return
		}

		release, err := releaseService.Rollback(r.Context(), id)
		if err != nil {
			responses.InternalErrorResponse(w, err.Error())
			return
		}

		if release == nil {
			responses.NotFoundResponse(w, "release not found")
			return
		}

		responses.SuccessResponse(w, release)
	}
}
