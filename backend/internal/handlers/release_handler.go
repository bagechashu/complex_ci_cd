package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/op/release-control/internal/models"
	"github.com/op/release-control/internal/services"
	"github.com/op/release-control/pkg/logger"
	"github.com/op/release-control/pkg/middleware"
)

// releaseHandler handles POST /api/v1/releases
func releaseHandler(service *services.ReleaseService, log *logger.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req models.ReleaseRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			middleware.ErrorResponse(w, 400, "BAD_REQUEST", "Invalid request body")
			return
		}

		release, err := service.Release(r.Context(), &req)
		if err != nil {
			log.Error("Failed to create release: %v", err)
			middleware.ErrorResponse(w, 500, "INTERNAL_ERROR", err.Error())
			return
		}

		middleware.JSONResponse(w, 202, release)
	}
}

// getReleaseHandler handles GET /api/v1/releases/{id}
func getReleaseHandler(service *services.ReleaseService, log *logger.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := chi.URLParam(r, "id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			middleware.ErrorResponse(w, 400, "BAD_REQUEST", "Invalid release ID")
			return
		}

		release, err := service.GetReleaseStatus(id)
		if err != nil {
			log.Error("Failed to get release: %v", err)
			middleware.ErrorResponse(w, 404, "NOT_FOUND", "Release not found")
			return
		}

		middleware.JSONResponse(w, 200, release)
	}
}

// getReleaseEventsHandler handles GET /api/v1/releases/{id}/events
func getReleaseEventsHandler(service *services.ReleaseService, log *logger.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := chi.URLParam(r, "id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			middleware.ErrorResponse(w, 400, "BAD_REQUEST", "Invalid release ID")
			return
		}

		events, err := service.GetReleaseEvents(id)
		if err != nil {
			log.Error("Failed to get release events: %v", err)
			middleware.ErrorResponse(w, 500, "INTERNAL_ERROR", "Failed to retrieve events")
			return
		}

		middleware.JSONResponse(w, 200, events)
	}
}

// listReleasesHandler handles GET /api/v1/releases
func listReleasesHandler(service *services.ReleaseService, log *logger.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		limit := 20
		offset := 0

		if l := r.URL.Query().Get("limit"); l != "" {
			if parsed, err := strconv.Atoi(l); err == nil {
				limit = parsed
			}
		}

		if o := r.URL.Query().Get("offset"); o != "" {
			if parsed, err := strconv.Atoi(o); err == nil {
				offset = parsed
			}
		}

		releases, err := service.ListReleases(limit, offset)
		if err != nil {
			log.Error("Failed to list releases: %v", err)
			middleware.ErrorResponse(w, 500, "INTERNAL_ERROR", "Failed to list releases")
			return
		}

		middleware.JSONResponse(w, 200, releases)
	}
}

// rollbackReleaseHandler handles POST /api/v1/releases/{id}/rollback
func rollbackReleaseHandler(service *services.ReleaseService, log *logger.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := chi.URLParam(r, "id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			middleware.ErrorResponse(w, 400, "BAD_REQUEST", "Invalid release ID")
			return
		}

		err = service.Rollback(r.Context(), id)
		if err != nil {
			log.Error("Failed to rollback release: %v", err)
			middleware.ErrorResponse(w, 500, "INTERNAL_ERROR", err.Error())
			return
		}

		middleware.JSONResponse(w, 200, map[string]string{"message": "Rollback initiated"})
	}
}
