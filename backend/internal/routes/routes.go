package routes

import (
	"net/http"
	"time"

	"built-and-deploy/internal/handlers/applications"
	"built-and-deploy/internal/handlers/clusters"
	"built-and-deploy/internal/handlers/environments"
	"built-and-deploy/internal/handlers/releases"
	"built-and-deploy/internal/handlers/shell_command_executions"
	"built-and-deploy/internal/handlers/shell_commands"
	"built-and-deploy/internal/handlers/shell_servers"
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
        
        // Releases endpoints
        r.Get("/releases", releases.List(container.Release(), log))
        r.Post("/releases", releases.Create(container.Release(), log))
        r.Get("/releases/{id}", releases.Get(container.Release(), log))
        r.Get("/releases/{id}/events", releases.ListEvents(container.Release(), log))
        r.Post("/releases/{id}/rollback", releases.Rollback(container.Release(), log))
        
        r.Get("/clusters", clusters.ListClustersHandler(container.Cluster(), log))
        r.Post("/clusters", clusters.CreateClusterHandler(container.Cluster(), log))
        r.Get("/clusters/{id}", clusters.GetClusterHandler(container.Cluster(), log))
        r.Put("/clusters/{id}", clusters.UpdateClusterHandler(container.Cluster(), log))
        r.Delete("/clusters/{id}", clusters.DeleteClusterHandler(container.Cluster(), log))
        r.Post("/clusters/{id}/test-connection", clusters.TestClusterConnectionHandler(container.Cluster(), log))
        r.Get("/environments", environments.ListEnvironmentsHandler(container.EnvironmentRepo(), log))
        r.Post("/environments", environments.CreateEnvironmentHandler(container.EnvironmentRepo(), log))
        r.Get("/workload-targets", workloads.ListWorkloadTargetsHandler(container.Workload(), log))
        r.Post("/workload-targets", workloads.CreateWorkloadTargetHandler(container.Workload(), log))
        r.Get("/workload-targets/by-app/{id}", workloads.GetAppClusterConfigsByAppHandler(container.Workload(), log))

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
        r.Post("/shell-commands/{id}/execute", shell_commands.Execute(container.Shell(), log))
        r.Delete("/shell-commands/{id}", shell_commands.DeleteShellCommandHandler(container.ShellCommandRepo(), log))

        // Shell Command Execution endpoints
        r.Get("/shell-command-executions", shell_command_executions.ListExecutions(container.Shell(), log))
        r.Get("/shell-command-executions/{id}", shell_command_executions.GetExecution(container.Shell(), log))
    })

    log.Info("NewRouter complete - routes registered")
    return r
}
