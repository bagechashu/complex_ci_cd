package routes

import (
	"net/http"
	"time"

	"built-and-deploy/internal/handlers/applications"
	"built-and-deploy/internal/handlers/clusters"
	"built-and-deploy/internal/handlers/shell_tasks"
	"built-and-deploy/internal/services"
	"built-and-deploy/pkg/logger"
	"built-and-deploy/pkg/middleware"

	"github.com/go-chi/chi/v5"
)

func NewRouter(container *services.ServiceContainer, log *logger.Logger) http.Handler {
    r := chi.NewRouter()

    // Apply global middleware
    r.Use(middleware.Timeout(30 * time.Second)) // 30 second request timeout
    r.Use(middleware.RequestID)
    r.Use(middleware.Logging(log))

    r.Route("/api/v1", func(r chi.Router) {
        r.Get("/applications", applications.List(container.Application(), log))
        r.Post("/applications", applications.Create(container.Application(), log))
        r.Get("/clusters", clusters.ListClustersHandler(container.Cluster(), log))
        r.Post("/clusters", clusters.CreateClusterHandler(container.Cluster(), log))
        r.Get("/clusters/:id", clusters.GetClusterHandler(container.Cluster(), log))
        r.Put("/clusters/:id", clusters.UpdateClusterHandler(container.Cluster(), log))
        r.Delete("/clusters/:id", clusters.DeleteClusterHandler(container.Cluster(), log))
        r.Get("/shell-tasks", shell_tasks.List(container.Shell(), log))
        r.Post("/shell-tasks", shell_tasks.Create(container.Shell(), log))
    })

    return r
}
