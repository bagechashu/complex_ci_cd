package responses

import (
	"encoding/json"
	"net/http"
)

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// JSONWithCode writes response with explicit HTTP status code and business code
func JSONWithCode(w http.ResponseWriter, httpStatus int, businessCode int, message string, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatus)
	json.NewEncoder(w).Encode(Response{Code: businessCode, Message: message, Data: data})
}

// JSON writes response with business code = 0 (success)
func JSON(w http.ResponseWriter, httpStatus int, message string, data interface{}) {
	JSONWithCode(w, httpStatus, 0, message, data)
}

// SuccessResponse: HTTP 200, code=0
func SuccessResponse(w http.ResponseWriter, data interface{}) {
	JSONWithCode(w, http.StatusOK, 0, "success", data)
}

// CreatedResponse: HTTP 201, code=0 (for non-async resource creation)
func CreatedResponse(w http.ResponseWriter, data interface{}) {
	JSONWithCode(w, http.StatusCreated, 0, "created", data)
}

// AcceptedResponse: HTTP 202, code=0 (for async operations)
func AcceptedResponse(w http.ResponseWriter, msg string, data interface{}) {
	JSONWithCode(w, http.StatusAccepted, 0, msg, data)
}

// BadRequestResponse: HTTP 200, code=3001 (validation error)
func BadRequestResponse(w http.ResponseWriter, msg string) {
	JSONWithCode(w, http.StatusOK, 3001, msg, nil)
}

// BadRequestResponseWithCode: HTTP 200, custom code (3xxx range)
func BadRequestResponseWithCode(w http.ResponseWriter, code int, msg string) {
	JSONWithCode(w, http.StatusOK, code, msg, nil)
}

// NotFoundResponse: HTTP 200, code=1001 (resource not found)
func NotFoundResponse(w http.ResponseWriter, msg string) {
	JSONWithCode(w, http.StatusOK, 1001, msg, nil)
}

// NotFoundResponseWithCode: HTTP 200, custom code (1xxx range)
func NotFoundResponseWithCode(w http.ResponseWriter, code int, msg string) {
	JSONWithCode(w, http.StatusOK, code, msg, nil)
}

// InternalErrorResponse: HTTP 200, code=9999 (internal server error)
func InternalErrorResponse(w http.ResponseWriter, msg string) {
	JSONWithCode(w, http.StatusOK, 9999, msg, nil)
}

// InternalErrorResponseWithCode: HTTP 200, custom code (9xxx range)
func InternalErrorResponseWithCode(w http.ResponseWriter, code int, msg string) {
	JSONWithCode(w, http.StatusOK, code, msg, nil)
}
