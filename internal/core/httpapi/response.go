package httpapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
)

const maxRequestBody = 1 << 20

// ErrorResponse is the consistent JSON error envelope.
type ErrorResponse struct {
	Error     ErrorBody `json:"error"`
	RequestID string    `json:"request_id,omitempty"`
} // @name ErrorResponse

// ErrorBody is the error detail object.
type ErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
} // @name ErrorBody

// CommandResult is the JSON body returned by command-style endpoints that
// forward the raw AzerothCore CLI output.
type CommandResult struct {
	Result string `json:"result"`
} // @name CommandResult

// WriteJSON writes v as a JSON response with the given status.
func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if v == nil {
		return
	}
	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Default().Warn("http: encode response", slog.String("error", err.Error()))
	}
}

// WriteError maps err to an APIError and writes the JSON envelope.
func WriteError(w http.ResponseWriter, r *http.Request, err error) {
	apiErr := StatusForError(err)
	WriteJSON(w, apiErr.Status, ErrorResponse{
		Error: ErrorBody{
			Code:    apiErr.Code,
			Message: apiErr.Message,
			Details: apiErr.Details,
		},
		RequestID: RequestIDFromContext(r.Context()),
	})
}

// DecodeJSON decodes a single JSON object from the request body, enforcing the
// body size limit. An oversized body returns ErrRequestTooLarge (413).
func DecodeJSON(r *http.Request, dst any) error {
	r.Body = http.MaxBytesReader(nil, r.Body, maxRequestBody)
	defer func() { _, _ = io.CopyN(io.Discard, r.Body, 4096) }()

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(dst); err != nil {
		return decodeJSONError(err, "malformed JSON body")
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return decodeJSONError(err, "request body must contain a single JSON object")
	}
	return nil
}

func decodeJSONError(err error, message string) error {
	var maxErr *http.MaxBytesError
	if errors.As(err, &maxErr) {
		return ErrRequestTooLarge
	}
	return fmt.Errorf("%w: %s: %v", ErrBadRequest, message, err)
}
