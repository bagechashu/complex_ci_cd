package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"built-and-deploy/pkg/errors"
	"built-and-deploy/pkg/logger"
)

// === 统一响应格式 ===

// Response 通用API响应结构
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   *ErrorInfo  `json:"error,omitempty"`
}

// ErrorInfo 错误信息详情
type ErrorInfo struct {
	Type      string `json:"type"`
	Details   string `json:"details,omitempty"`
	Timestamp string `json:"timestamp"`
	RequestID string `json:"request_id"`
}

// === Application DTOs ===

type CreateApplicationRequest struct {
	Name        string `json:"name" validate:"required,min=1,max=100"`
	ImageName   string `json:"image_name" validate:"required"`
	GitRepo     string `json:"git_repo"`
	BuildType   string `json:"build_type" validate:"required,oneof=docker maven npm go other"`
	Description string `json:"description" validate:"max=500"`
}

type UpdateApplicationRequest struct {
	Name        string `json:"name" validate:"max=100"`
	ImageName   string `json:"image_name"`
	GitRepo     string `json:"git_repo"`
	BuildType   string `json:"build_type" validate:"oneof=docker maven npm go other"`
	Description string `json:"description" validate:"max=500"`
}

type ApplicationResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Repository  string    `json:"repository"`
	BuildType   string    `json:"build_type"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// === Cluster DTOs ===

type CreateClusterRequest struct {
	Name              string `json:"name" validate:"required,min=1,max=100"`
	Type              string `json:"type" validate:"required,oneof=kubernetes"`
	Environment       string `json:"environment" validate:"max=100"`
	RegistryPrefix    string `json:"registry_prefix" validate:"max=200"`
	Labels            string `json:"labels" validate:"max=500"`
	Kubeconfig        string `json:"kubeconfig" validate:"required"`
	KubernetesVersion string `json:"kubernetes_version"` // 可选，K8s 版本号，如 "1.23.6"
	Description       string `json:"description" validate:"max=500"`
}

type UpdateClusterRequest struct {
	Name                string `json:"name" validate:"max=100"`
	Kubeconfig          string `json:"kubeconfig"` // 可选，如果为空字符串则不更新 kubeconfig
	KubernetesVersion   string `json:"kubernetes_version"` // 可选，K8s 版本号，如 "1.23.6"
	Description         string `json:"description" validate:"max=500"`
}

type ClusterResponse struct {
	ID                  string    `json:"id"`
	Name                string    `json:"name"`
	Type                string    `json:"type"`
	K8sConnectionStatus string    `json:"k8s_connection_status"`
	Description         string    `json:"description"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

// === Workload Target DTOs ===

type CreateWorkloadTargetRequest struct {
	AppID          string `json:"app_id" validate:"required"`
	EnvID          string `json:"env_id" validate:"required"`
	ClusterID      string `json:"cluster_id" validate:"required"`
	Namespace      string `json:"namespace" validate:"required,min=1,max=50"`
	WorkloadName   string `json:"workload_name" validate:"required,min=1,max=100"`
	WorkloadType   string `json:"workload_type" validate:"required,oneof=Deployment StatefulSet DaemonSet"`
	ContainerName  string `json:"container_name" validate:"required,min=1,max=100"`
	RegistryDomain string `json:"registry_domain" validate:"required"`
	ImageRepo      string `json:"image_repo" validate:"required"`
}

type UpdateWorkloadTargetRequest struct {
	Namespace      string `json:"namespace" validate:"max=50"`
	WorkloadName   string `json:"workload_name" validate:"max=100"`
	WorkloadType   string `json:"workload_type" validate:"oneof=Deployment StatefulSet DaemonSet"`
	ContainerName  string `json:"container_name" validate:"max=100"`
	RegistryDomain string `json:"registry_domain"`
	ImageRepo      string `json:"image_repo"`
}

type WorkloadTargetResponse struct {
	ID             string    `json:"id"`
	AppID          string    `json:"app_id"`
	EnvID          string    `json:"env_id"`
	ClusterID      string    `json:"cluster_id"`
	Namespace      string    `json:"namespace"`
	WorkloadName   string    `json:"workload_name"`
	WorkloadType   string    `json:"workload_type"`
	ContainerName  string    `json:"container_name"`
	RegistryDomain string    `json:"registry_domain"`
	ImageRepo      string    `json:"image_repo"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// === Release DTOs ===

type CreateReleaseRequest struct {
	AppID     int    `json:"app_id" validate:"required,gt=0"`
	EnvID     int    `json:"env_id" validate:"required,gt=0"`
	ClusterID int    `json:"cluster_id" validate:"required,gt=0"`
	Image     string `json:"image" validate:"required"`
	User      string `json:"user" validate:"omitempty"`
}

type ReleaseStatusResponse struct {
	ID            string     `json:"id"`
	AppID         string     `json:"app_id"`
	EnvID         string     `json:"env_id"`
	ClusterID     string     `json:"cluster_id"`
	Image         string     `json:"image"`
	Status        string     `json:"status"`
	PreviousImage string     `json:"previous_image,omitempty"`
	TriggeredBy   string     `json:"triggered_by"`
	StartedAt     time.Time  `json:"started_at"`
	CompletedAt   *time.Time `json:"completed_at,omitempty"`
	ErrorMessage  string     `json:"error_message,omitempty"`
	RequestID     string     `json:"request_id,omitempty"`
}

type ReleaseEventResponse struct {
	ID           string    `json:"id"`
	ReleaseID    string    `json:"release_id"`
	EventType    string    `json:"event_type"`
	EventMessage string    `json:"event_message"`
	CreatedAt    time.Time `json:"created_at"`
}

// === Shell Command DTOs ===

type CreateShellCommandRequest struct {
	ServerID    string `json:"server_id" validate:"required"`
	Command     string `json:"command" validate:"required"`
	Description string `json:"description" validate:"max=500"`
}

type PublishShellCommandRequest struct {
	Publish bool `json:"publish"`
}

type ShellCommandResponse struct {
	ID          string    `json:"id"`
	ServerID    string    `json:"server_id"`
	Command     string    `json:"command"`
	Description string    `json:"description"`
	IsPublished bool      `json:"is_published"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// === Shell Server DTOs ===

type CreateShellServerRequest struct {
	Name       string `json:"name" validate:"required,min=1,max=100"`
	Host       string `json:"host" validate:"required"`
	Port       int    `json:"port" validate:"required,min=1,max=65535"`
	Username   string `json:"username" validate:"required"`
	AuthType   string `json:"auth_type" validate:"required,oneof=password key"`
	Password   string `json:"password"`
	PrivateKey string `json:"private_key"`
}

type UpdateShellServerRequest struct {
	Name       string `json:"name" validate:"max=100"`
	Host       string `json:"host"`
	Port       int    `json:"port" validate:"omitempty,min=1,max=65535"`
	Username   string `json:"username"`
	Password   string `json:"password"`
	PrivateKey string `json:"private_key"`
}

type ShellServerResponse struct {
	ID              string     `json:"id"`
	Name            string     `json:"name"`
	Host            string     `json:"host"`
	Port            int        `json:"port"`
	Username        string     `json:"username"`
	AuthType        string     `json:"auth_type"`
	Status          string     `json:"status"`
	LastConnectedAt *time.Time `json:"last_connected_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// === Shell Execution DTOs ===

type ExecuteShellCommandRequest struct {
	CommandID string `json:"command_id" validate:"required"`
	ServerID  string `json:"server_id" validate:"required"`
}

type ExecuteShellResponse struct {
	ExecutionID string     `json:"execution_id"`
	Status      string     `json:"status"`
	Output      string     `json:"output,omitempty"`
	ExitCode    int        `json:"exit_code,omitempty"`
	Duration    float64    `json:"duration,omitempty"`
	Error       string     `json:"error,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

// === Pagination ===

type PaginationRequest struct {
	Page     int `query:"page"`
	PageSize int `query:"page_size"`
}

type PaginatedResponse struct {
	Items      interface{} `json:"items"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
	Total      int         `json:"total"`
	TotalPages int         `json:"total_pages"`
}

// === Response Functions ===

// RespondJSON 返回成功JSON响应
func RespondJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := Response{
		Code:    0,
		Message: getMessageByStatus(statusCode),
		Data:    data,
	}

	json.NewEncoder(w).Encode(response)
}

// RespondError 返回错误JSON响应
func RespondError(w http.ResponseWriter, statusCode int, message string, err error, requestID string, log *logger.Logger) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	// 记录错误日志
	if log != nil && err != nil {
		log.Error("API error",
			"status", statusCode,
			"message", message,
			"error", err,
			"request_id", requestID)
	}

	response := Response{
		Code:    statusCode,
		Message: message,
		Error: &ErrorInfo{
			Type:      getErrorType(statusCode),
			Details:   fmt.Sprintf("%v", err),
			Timestamp: time.Now().Format(time.RFC3339),
			RequestID: requestID,
		},
	}

	json.NewEncoder(w).Encode(response)
}

// RespondServiceError 根据ServiceError返回响应
func RespondServiceError(w http.ResponseWriter, svcErr *errors.ServiceError, requestID string, log *logger.Logger) {
	RespondError(w, svcErr.Status, svcErr.Message, svcErr.Err, requestID, log)
}

// HandleServiceError 处理Service层错误
func HandleServiceError(w http.ResponseWriter, err error, requestID string, log *logger.Logger) {
	if svcErr, ok := errors.IsServiceError(err); ok {
		RespondServiceError(w, svcErr, requestID, log)
	} else {
		RespondError(w, http.StatusInternalServerError, "Internal server error", err, requestID, log)
	}
}

// ValidateJSONRequest 验证JSON请求
func ValidateJSONRequest(w http.ResponseWriter, r *http.Request, target interface{}, requestID string, log *logger.Logger) error {
	if r.Body == nil {
		msg := "request body cannot be empty"
		RespondError(w, http.StatusBadRequest, msg, fmt.Errorf(msg), requestID, log)
		return fmt.Errorf(msg)
	}

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(target); err != nil {
		msg := "invalid JSON in request body"
		RespondError(w, http.StatusBadRequest, msg, err, requestID, log)
		return err
	}

	return nil
}

// === 辅助函数 ===

// getMessageByStatus 根据HTTP状态码获取消息
func getMessageByStatus(statusCode int) string {
	messages := map[int]string{
		http.StatusOK:                  "success",
		http.StatusCreated:             "created",
		http.StatusAccepted:            "accepted",
		http.StatusNoContent:           "deleted",
		http.StatusBadRequest:          "bad request",
		http.StatusUnauthorized:        "unauthorized",
		http.StatusForbidden:           "forbidden",
		http.StatusNotFound:            "not found",
		http.StatusConflict:            "conflict",
		http.StatusInternalServerError: "internal server error",
		http.StatusServiceUnavailable:  "service unavailable",
	}

	if msg, ok := messages[statusCode]; ok {
		return msg
	}
	return "unknown"
}

// getErrorType 根据HTTP状态码获取错误类型
func getErrorType(statusCode int) string {
	types := map[int]string{
		http.StatusBadRequest:          "ValidationError",
		http.StatusUnauthorized:        "AuthenticationError",
		http.StatusForbidden:           "AuthorizationError",
		http.StatusNotFound:            "NotFoundError",
		http.StatusConflict:            "ConflictError",
		http.StatusInternalServerError: "ServerError",
	}

	if t, ok := types[statusCode]; ok {
		return t
	}
	return "Error"
}
