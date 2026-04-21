package errors

import (
	"fmt"
)

// ServiceError 统一的服务层错误类型
type ServiceError struct {
	Code    string // 业务错误码
	Message string // 用户友好的消息
	Err     error  // 原始错误（用于日志）
	Status  int    // HTTP状态码
}

func (e *ServiceError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// NewServiceError 创建错误
func NewServiceError(code, message string) *ServiceError {
	return &ServiceError{
		Code:    code,
		Message: message,
		Status:  mapCodeToStatus(code),
	}
}

// NewServiceErrorWithCause 创建带原始错误的错误
func NewServiceErrorWithCause(code, message string, err error) *ServiceError {
	return &ServiceError{
		Code:    code,
		Message: message,
		Err:     err,
		Status:  mapCodeToStatus(code),
	}
}

// mapCodeToStatus 映射错误码到HTTP状态码
func mapCodeToStatus(code string) int {
	mapping := map[string]int{
		"NOT_FOUND":           404,
		"PERMISSION_DENIED":   403,
		"INVALID_INPUT":       400,
		"INVALID_JSON":        400,
		"CONFLICT":            409,
		"CONNECTION_FAILED":   503,
		"DATABASE":            500,
		"ENCRYPTION":          500,
		"DECRYPTION":          500,
		"INVALID_STATE":       400,
		"INTERNAL_ERROR":      500,
	}

	if status, ok := mapping[code]; ok {
		return status
	}
	return 500
}

// IsServiceError 检查是否为ServiceError
func IsServiceError(err error) (*ServiceError, bool) {
	svcErr, ok := err.(*ServiceError)
	return svcErr, ok
}
