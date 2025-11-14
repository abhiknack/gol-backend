package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/supabase-community/supabase-go"
)

// Pagination holds pagination parameters
type Pagination struct {
	Limit  int
	Offset int
}

// SupabaseRepository defines the interface for Supabase data access
type SupabaseRepository interface {
	Query(ctx context.Context, table string, filters map[string]interface{}, pagination Pagination) ([]map[string]interface{}, error)
	GetByID(ctx context.Context, table string, id string) (map[string]interface{}, error)
}

// supabaseRepository implements SupabaseRepository
type supabaseRepository struct {
	client *supabase.Client
}

// NewSupabaseRepository creates a new Supabase repository instance
func NewSupabaseRepository(url, apiKey string) (SupabaseRepository, error) {
	if url == "" || apiKey == "" {
		return nil, NewConnectionError(errors.New("Supabase URL and API key are required"))
	}

	client, err := supabase.NewClient(url, apiKey, nil)
	if err != nil {
		return nil, NewConnectionError(err)
	}

	return &supabaseRepository{
		client: client,
	}, nil
}

// Query retrieves records from a Supabase table with filtering and pagination
func (r *supabaseRepository) Query(ctx context.Context, table string, filters map[string]interface{}, pagination Pagination) ([]map[string]interface{}, error) {
	// Check for context cancellation or timeout
	if err := ctx.Err(); err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return nil, NewTimeoutError(err)
		}
		return nil, NewQueryError(err)
	}

	// Execute query with timeout handling
	resultChan := make(chan queryResult, 1)
	go func() {
		results, err := r.executeQuery(table, filters, pagination)
		resultChan <- queryResult{data: results, err: err}
	}()

	select {
	case <-ctx.Done():
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return nil, NewTimeoutError(ctx.Err())
		}
		return nil, NewQueryError(ctx.Err())
	case result := <-resultChan:
		if result.err != nil {
			return nil, r.handleError(result.err, table)
		}
		return result.data, nil
	}
}

// executeQuery performs the actual query execution
func (r *supabaseRepository) executeQuery(table string, filters map[string]interface{}, pagination Pagination) ([]map[string]interface{}, error) {
	// Start building the query
	query := r.client.From(table).Select("*", "exact", false)

	// Apply filters
	for key, value := range filters {
		query = query.Eq(key, fmt.Sprintf("%v", value))
	}

	// Apply pagination
	if pagination.Limit > 0 {
		query = query.Limit(pagination.Limit, "")
	}
	if pagination.Offset > 0 {
		query = query.Range(pagination.Offset, pagination.Offset+pagination.Limit-1, "")
	}

	// Execute query
	var results []map[string]interface{}
	_, err := query.ExecuteTo(&results)
	if err != nil {
		return nil, err
	}

	return results, nil
}

type queryResult struct {
	data []map[string]interface{}
	err  error
}

// GetByID retrieves a single record by ID from a Supabase table
func (r *supabaseRepository) GetByID(ctx context.Context, table string, id string) (map[string]interface{}, error) {
	// Check for context cancellation or timeout
	if err := ctx.Err(); err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return nil, NewTimeoutError(err)
		}
		return nil, NewQueryError(err)
	}

	// Execute query with timeout handling
	resultChan := make(chan getByIDResult, 1)
	go func() {
		result, err := r.executeGetByID(table, id)
		resultChan <- getByIDResult{data: result, err: err}
	}()

	select {
	case <-ctx.Done():
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return nil, NewTimeoutError(ctx.Err())
		}
		return nil, NewQueryError(ctx.Err())
	case result := <-resultChan:
		if result.err != nil {
			// Check if it's a not found error
			if r.isNotFoundError(result.err) {
				return nil, NewNotFoundError(table, id)
			}
			return nil, r.handleError(result.err, table)
		}
		return result.data, nil
	}
}

// executeGetByID performs the actual get by ID execution
func (r *supabaseRepository) executeGetByID(table string, id string) (map[string]interface{}, error) {
	query := r.client.From(table).Select("*", "exact", false).Eq("id", id).Single()

	var result map[string]interface{}
	_, err := query.ExecuteTo(&result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

type getByIDResult struct {
	data map[string]interface{}
	err  error
}

// handleError converts Supabase errors to appropriate RepositoryErrors
func (r *supabaseRepository) handleError(err error, table string) error {
	errMsg := err.Error()
	errMsgLower := strings.ToLower(errMsg)

	// Check for connection errors
	if strings.Contains(errMsgLower, "connection") || 
	   strings.Contains(errMsgLower, "network") ||
	   strings.Contains(errMsgLower, "dial") {
		return NewConnectionError(err)
	}

	// Check for timeout errors
	if strings.Contains(errMsgLower, "timeout") || 
	   strings.Contains(errMsgLower, "deadline") {
		return NewTimeoutError(err)
	}

	// Default to query error
	return NewQueryError(err)
}

// isNotFoundError checks if the error indicates a record was not found
func (r *supabaseRepository) isNotFoundError(err error) bool {
	errMsg := strings.ToLower(err.Error())
	return strings.Contains(errMsg, "not found") || 
	       strings.Contains(errMsg, "no rows") ||
	       strings.Contains(errMsg, "pgrst116")
}
