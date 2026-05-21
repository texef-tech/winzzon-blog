package handler

import (
	"encoding/json"
	"net/http"
)

// SuccessResponse defines the standard payload structure for generic successful requests.
type SuccessResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// ErrorResponse defines the standard payload structure for failed requests.
type ErrorResponse struct {
	Success bool        `json:"success"`
	Error   ErrorDetail `json:"error"`
}

type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// ValidationErrorResponse defines the structure for field-level validation errors.
type ValidationErrorResponse struct {
	Success bool                  `json:"success"`
	Error   ValidationErrorDetail `json:"error"`
}

type ValidationErrorDetail struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Errors  map[string]string `json:"errors"`
}

// JSON sends a raw JSON response with a specific HTTP status code.
func JSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// SuccessJSON sends a wrapped standard success response.
func SuccessJSON(w http.ResponseWriter, status int, message string, data interface{}) {
	JSON(w, status, SuccessResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// ErrorJSON sends a standard error response containing code and message description.
func ErrorJSON(w http.ResponseWriter, status int, code, message string) {
	JSON(w, status, ErrorResponse{
		Success: false,
		Error: ErrorDetail{
			Code:    code,
			Message: message,
		},
	})
}

// ValidationErrorJSON sends a HTTP 422 standard validation error response with field details.
func ValidationErrorJSON(w http.ResponseWriter, errors map[string]string) {
	JSON(w, http.StatusUnprocessableEntity, ValidationErrorResponse{
		Success: false,
		Error: ValidationErrorDetail{
			Code:    "VALIDATION_FAILED",
			Message: "Validation failed",
			Errors:  errors,
		},
	})
}

// Helper wrappers for common HTTP status responses
func BadRequest(w http.ResponseWriter, code, message string) {
	ErrorJSON(w, http.StatusBadRequest, code, message)
}

func Unauthorized(w http.ResponseWriter, code, message string) {
	ErrorJSON(w, http.StatusUnauthorized, code, message)
}

func Forbidden(w http.ResponseWriter, code, message string) {
	ErrorJSON(w, http.StatusForbidden, code, message)
}

func NotFound(w http.ResponseWriter, code, message string) {
	ErrorJSON(w, http.StatusNotFound, code, message)
}

func Unprocessable(w http.ResponseWriter, code, message string) {
	ErrorJSON(w, http.StatusUnprocessableEntity, code, message)
}

func ServerError(w http.ResponseWriter, code, message string) {
	ErrorJSON(w, http.StatusInternalServerError, code, message)
}

