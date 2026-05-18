package routes

import (
	"net/http"
	"time"

	"built-and-deploy/internal/handlers/applications"
	"built-and-deploy/internal/handlers/clusters"
	"built-and-deploy/internal/handlers/environments"
	"built-and-deploy/internal/handlers/shell_commands"
	"built-and-deploy/internal/handlers/shell_servers"
	"built-and-deploy/internal/handlers/shell_tasks"
	"built-and-deploy/internal/handlers/workloads"
	"built-and-deploy/internal/services"
	"built-and-deploy/pkg/logger"
	"built-and-deploy/pkg/middleware"

	"github.com/go-chi/chi/v5"
)

func NewRouter(container *services.ServiceContainer, log *logger.Logger) http.Handler {
    log.Info("NewRouter called - starting to register routes")
    r := chi.NewRouter()

    // Apply global middleware
    r.Use(middleware.Timeout(30 * time.Second))
    r.Use(middleware.RequestID)
    r.Use(middleware.Logging(log))

    log.Info("Registering API v1 routes")
    r.Route("/api/v1", func(r chi.Router) {
        r.Get("/applications", applications.List(container.Application(), log))
        r.Post("/applications", applications.Create(container.Application(), log))
        r.Get("/clusters", clusters.ListClustersHandler(container.Cluster(), log))
        r.Post("/clusters", clusters.CreateClusterHandler(container.Cluster(), log))
        r.Get("/clusters/{id}", clusters.GetClusterHandler(container.Cluster(), log))
        r.Put("/clusters/{id}", clusters.UpdateClusterHandler(container.Cluster(), log))
        r.Delete("/clusters/{id}", clusters.DeleteClusterHandler(container.Cluster(), log))
        r.Get("/environments", environments.ListEnvironmentsHandler(container.EnvironmentRepo(), log))
        r.Post("/environments", environments.CreateEnvironmentHandler(container.EnvironmentRepo(), log))
        r.Get("/workload-targets", workloads.ListWorkloadTargetsHandler(container.Workload(), log))
        r.Post("/workload-targets", workloads.CreateWorkloadTargetHandler(container.Workload(), log))
        r.Get("/app-cluster-configs/by-app/{id}", workloads.GetAppClusterConfigsByAppHandler(container.Workload(), log))
        r.Get("/app-cluster-configs", workloads.GetAppClusterConfigsHandler(container.Workload(), log))
        r.Post("/app-cluster-configs", workloads.CreateAppClusterConfigHandler(container.WorkloadRepo(), log))

        // Shell Servers endpoints
        r.Get("/shell-servers", shell_servers.ListShellServersHandler(container.ShellServerRepo(), log))
        r.Get("/shell-servers/{id}", shell_servers.GetShellServerHandler(container.ShellServerRepo(), log))
        r.Post("/shell-servers", shell_servers.CreateShellServerHandler(container.ShellServerRepo(), log))
        r.Put("/shell-servers/{id}", shell_servers.UpdateShellServerHandler(container.ShellServerRepo(), log))
        r.Delete("/shell-servers/{id}", shell_servers.DeleteShellServerHandler(container.ShellServerRepo(), log))

        // Shell Commands endpoints
        r.Get("/shell-commands", shell_commands.ListShellCommandsHandler(container.ShellCommandRepo(), log))
        r.Get("/shell-commands/{id}", shell_commands.GetShellCommandHandler(container.ShellCommandRepo(), log))
        r.Post("/shell-commands", shell_commands.CreateShellCommandHandler(container.ShellCommandRepo(), log))
        r.Put("/shell-commands/{id}", shell_commands.UpdateShellCommandHandler(container.ShellCommandRepo(), log))
        r.Post("/shell-commands/{id}/publish", shell_commands.PublishShellCommandHandler(container.ShellCommandRepo(), log))
        r.Post("/shell-commands/{id}/unpublish", shell_commands.UnpublishShellCommandHandler(container.ShellCommandRepo(), log))
        r.Delete("/shell-commands/{id}", shell_commands.DeleteShellCommandHandler(container.ShellCommandRepo(), log))

        // Shell Tasks endpoints
        r.Get("/shell-tasks", shell_tasks.List(container.Shell(), log))
        r.Post("/shell-tasks", shell_tasks.Create(container.Shell(), log))
        r.Get("/shell-tasks/{id}", shell_tasks.Get(container.Shell(), log))
        r.Put("/shell-tasks/{id}", shell_tasks.Update(container.Shell(), log))
        r.Delete("/shell-tasks/{id}", shell_tasks.Delete(container.Shell(), log))
        r.Post("/shell-commands/execute", shell_tasks.Execute(container.Shell(), log))
    })

    log.Info("NewRouter complete - routes registered")
    return r
}
