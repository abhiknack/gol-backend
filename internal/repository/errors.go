package repository

import (
	"errors"
	"fmt"
	"net/http"
)

// RepositoryError represents a repository-level error with HTTP status code
type RepositoryError struct {
	StatusCode int
	Message    string
	Err        error
}

func (e *RepositoryError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func (e *RepositoryError) Unwrap() error {
	return e.Err
}

// Error constructors
func NewConnectionError(err error) *RepositoryError {
	return &RepositoryError{
		StatusCode: http.StatusServiceUnavailable,
		Message:    "Failed to connect to Supabase",
		Err:        err,
	}
}

func NewQueryError(err error) *RepositoryError {
	return &RepositoryError{
		StatusCode: http.StatusInternalServerError,
		Message:    "Failed to execute query",
		Err:        err,
	}
}

func NewTimeoutError(err error) *RepositoryError {
	return &RepositoryError{
		StatusCode: http.StatusGatewayTimeout,
		Message:    "Request timeout",
		Err:        err,
	}
}

func NewNotFoundError(table, id string) *RepositoryError {
	return &RepositoryError{
		StatusCode: http.StatusNotFound,
		Message:    fmt.Sprintf("Record not found in table %s with id %s", table, id),
		Err:        nil,
	}
}

// IsRepositoryError checks if an error is a RepositoryError
func IsRepositoryError(err error) bool {
	var repoErr *RepositoryError
	return errors.As(err, &repoErr)
}

// GetStatusCode extracts the HTTP status code from a RepositoryError
func GetStatusCode(err error) int {
	var repoErr *RepositoryError
	if errors.As(err, &repoErr) {
		return repoErr.StatusCode
	}
	return http.StatusInternalServerError
}
