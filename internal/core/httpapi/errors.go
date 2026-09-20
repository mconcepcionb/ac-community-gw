// Package httpapi provides the HTTP server, middleware and JSON helpers for
// the gateway API. It uses net/http and http.ServeMux exclusively.
package httpapi

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/mconcepcionb/ac-community-gw/internal/core/auth"
	"github.com/mconcepcionb/ac-community-gw/internal/core/azerothcore"
	"github.com/mconcepcionb/ac-community-gw/internal/core/azerothdb"
	"github.com/mconcepcionb/ac-community-gw/internal/core/commands"
	"github.com/mconcepcionb/ac-community-gw/internal/core/services"
)

// APIError is an error carrying the HTTP status and stable error code returned
// to API consumers.
type APIError struct {
	Status  int
	Code    string
	Message string
	Details any
}

// Error implements error.
func (e *APIError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// NewAPIError builds an APIError.
func NewAPIError(status int, code, message string) *APIError {
	return &APIError{Status: status, Code: code, Message: message}
}

// WithDetails returns a copy of the error carrying extra details.
func (e *APIError) WithDetails(details any) *APIError {
	clone := *e
	clone.Details = details
	return &clone
}

var (
	// ErrUnauthorized is returned when authentication is required or invalid.
	ErrUnauthorized = NewAPIError(http.StatusUnauthorized, "unauthorized", "authentication required")
	// ErrForbidden is returned when the principal lacks a permission.
	ErrForbidden = NewAPIError(http.StatusForbidden, "forbidden", "insufficient permissions")
	// ErrNotFound is returned when a resource does not exist.
	ErrNotFound = NewAPIError(http.StatusNotFound, "not_found", "resource not found")
	// ErrBadRequest is returned for malformed requests.
	ErrBadRequest = NewAPIError(http.StatusBadRequest, "bad_request", "malformed request")
	// ErrConflict is returned when a request conflicts with current state.
	ErrConflict = NewAPIError(http.StatusConflict, "conflict", "resource conflict")
	// ErrUnprocessable is returned when a payload fails validation.
	ErrUnprocessable = NewAPIError(http.StatusUnprocessableEntity, "unprocessable_entity", "validation failed")
	// ErrInternal is returned for unexpected server errors.
	ErrInternal = NewAPIError(http.StatusInternalServerError, "internal_error", "internal server error")
	// ErrNotImplemented marks deliberately unimplemented endpoints.
	ErrNotImplemented = NewAPIError(http.StatusNotImplemented, "not_implemented", "not implemented yet")
	// ErrUnavailable is returned when a dependency required to serve the
	// request is not configured or temporarily unavailable.
	ErrUnavailable = NewAPIError(http.StatusServiceUnavailable, "unavailable", "service unavailable")
	// ErrTooManyRequests is returned when a client exceeds a rate limit.
	ErrTooManyRequests = NewAPIError(http.StatusTooManyRequests, "too_many_requests", "too many requests")
	// ErrRequestTooLarge is returned when a request body exceeds the limit.
	ErrRequestTooLarge = NewAPIError(http.StatusRequestEntityTooLarge, "request_too_large", "request body is too large")
	// ErrBadGateway is returned when an upstream integration fails.
	ErrBadGateway = NewAPIError(http.StatusBadGateway, "bad_gateway", "upstream integration failed")
)

// StatusForError maps any error to an APIError suitable for the response.
func StatusForError(err error) *APIError {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr
	}
	switch {
	case errors.Is(err, auth.ErrSessionNotFound), errors.Is(err, auth.ErrSessionExpired):
		return ErrUnauthorized
	case errors.Is(err, commands.ErrNotFound), errors.Is(err, services.ErrNotFound):
		return ErrNotFound
	case errors.Is(err, commands.ErrInvalidPayload), errors.Is(err, commands.ErrInvalidResult):
		return ErrUnprocessable
	case errors.Is(err, azerothdb.ErrUnavailable), errors.Is(err, azerothcore.ErrNotConfigured):
		return ErrUnavailable
	default:
		return ErrInternal
	}
}
