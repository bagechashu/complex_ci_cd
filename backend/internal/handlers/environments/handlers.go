package environments

import (
	"encoding/json"
	"net/http"

	"built-and-deploy/internal/repository"
	"built-and-deploy/pkg/logger"
	"built-and-deploy/pkg/responses"
)

// ListEnvironmentsHandler handles GET /environments request to retrieve all environments.
func ListEnvironmentsHandler(envRepo *repository.EnvironmentRepository, log *logger.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		environments, err := envRepo.List(1000, 0)
		if err != nil {
			responses.InternalErrorResponse(w, err.Error())
			return
		}
		responses.SuccessResponse(w, environments)
	}
}

// CreateEnvironmentHandler handles POST /environments request to create a new environment.
func CreateEnvironmentHandler(envRepo *repository.EnvironmentRepository, log *logger.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			responses.BadRequestResponse(w, "invalid request body")
			return
		}

		// For now, just return success (full implementation would validate and save)
		responses.SuccessResponse(w, map[string]string{"message": "created"})
	}
}
