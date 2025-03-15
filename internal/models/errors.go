package models

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// AppError represents an application-specific error
type AppError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
	Code    int    `json:"-"`
}

func (e *AppError) Error() string {
	return e.Message
}

// NewValidationError creates a new validation error
func NewValidationError(message string) *AppError {
	return &AppError{
		Type:    "validation_error",
		Message: message,
		Code:    http.StatusBadRequest,
	}
}

// NewNotFoundError creates a new not found error
func NewNotFoundError(message string) *AppError {
	return &AppError{
		Type:    "not_found",
		Message: message,
		Code:    http.StatusNotFound,
	}
}

// NewInternalError creates a new internal server error
func NewInternalError(message string) *AppError {
	return &AppError{
		Type:    "internal_error",
		Message: message,
		Code:    http.StatusInternalServerError,
	}
}

// ReadJSON reads JSON from request body into target
func ReadJSON(r *http.Request, target interface{}) error {
	if err := json.NewDecoder(r.Body).Decode(target); err != nil {
		return fmt.Errorf("failed to decode JSON: %w", err)
	}
	return nil
}

// WriteJSON writes JSON to response writer
func WriteJSON(w http.ResponseWriter, data interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(data)
}
