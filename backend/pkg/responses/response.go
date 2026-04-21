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

func JSON(w http.ResponseWriter, statusCode int, message string, data interface{}) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(statusCode)
    json.NewEncoder(w).Encode(Response{Code: statusCode, Message: message, Data: data})
}

func SuccessResponse(w http.ResponseWriter, data interface{}) { JSON(w, http.StatusOK, "success", data) }
func CreatedResponse(w http.ResponseWriter, data interface{}) { JSON(w, http.StatusCreated, "created", data) }
func BadRequestResponse(w http.ResponseWriter, msg string) { JSON(w, http.StatusBadRequest, msg, nil) }
func InternalErrorResponse(w http.ResponseWriter, msg string) { JSON(w, http.StatusInternalServerError, msg, nil) }
