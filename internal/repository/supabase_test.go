package repository

import (
	"context"
	"errors"
	"testing"
	"time"
)

// MockSupabaseClient is a mock implementation for testing
type MockSupabaseClient struct {
	queryFunc   func(table string, filters map[string]interface{}, pagination Pagination) ([]map[string]interface{}, error)
	getByIDFunc func(table string, id string) (map[string]interface{}, error)
}

// mockSupabaseRepository is a test implementation that uses mock functions
type mockSupabaseRepository struct {
	mock *MockSupabaseClient
}

func newMockRepository(mock *MockSupabaseClient) SupabaseRepository {
	return &mockSupabaseRepository{mock: mock}
}

func (m *mockSupabaseRepository) Query(ctx context.Context, table string, filters map[string]interface{}, pagination Pagination) ([]map[string]interface{}, error) {
	if err := ctx.Err(); err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return nil, NewTimeoutError(err)
		}
		return nil, NewQueryError(err)
	}

	if m.mock.queryFunc != nil {
		return m.mock.queryFunc(table, filters, pagination)
	}
	return nil, errors.New("queryFunc not implemented")
}

func (m *mockSupabaseRepository) GetByID(ctx context.Context, table string, id string) (map[string]interface{}, error) {
	if err := ctx.Err(); err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return nil, NewTimeoutError(err)
		}
		return nil, NewQueryError(err)
	}

	if m.mock.getByIDFunc != nil {
		return m.mock.getByIDFunc(table, id)
	}
	return nil, errors.New("getByIDFunc not implemented")
}

func TestNewSupabaseRepository(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		apiKey  string
		wantErr bool
	}{
		{
			name:    "empty url",
			url:     "",
			apiKey:  "test-key",
			wantErr: true,
		},
		{
			name:    "empty api key",
			url:     "https://test.supabase.co",
			apiKey:  "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewSupabaseRepository(tt.url, tt.apiKey)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewSupabaseRepository() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestQuery(t *testing.T) {
	tests := []struct {
		name       string
		table      string
		filters    map[string]interface{}
		pagination Pagination
		mockFunc   func(table string, filters map[string]interface{}, pagination Pagination) ([]map[string]interface{}, error)
		wantErr    bool
		wantLen    int
	}{
		{
			name:  "successful query",
			table: "products",
			filters: map[string]interface{}{
				"category": "dairy",
			},
			pagination: Pagination{Limit: 10, Offset: 0},
			mockFunc: func(table string, filters map[string]interface{}, pagination Pagination) ([]map[string]interface{}, error) {
				return []map[string]interface{}{
					{"id": "1", "name": "Milk"},
					{"id": "2", "name": "Cheese"},
				}, nil
			},
			wantErr: false,
			wantLen: 2,
		},
		{
			name:  "query with pagination",
			table: "products",
			filters: map[string]interface{}{
				"category": "dairy",
			},
			pagination: Pagination{Limit: 5, Offset: 10},
			mockFunc: func(table string, filters map[string]interface{}, pagination Pagination) ([]map[string]interface{}, error) {
				if pagination.Limit != 5 || pagination.Offset != 10 {
					t.Errorf("Expected pagination Limit=5, Offset=10, got Limit=%d, Offset=%d", pagination.Limit, pagination.Offset)
				}
				return []map[string]interface{}{
					{"id": "11", "name": "Product 11"},
				}, nil
			},
			wantErr: false,
			wantLen: 1,
		},
		{
			name:    "connection error",
			table:   "products",
			filters: map[string]interface{}{},
			pagination: Pagination{Limit: 10, Offset: 0},
			mockFunc: func(table string, filters map[string]interface{}, pagination Pagination) ([]map[string]interface{}, error) {
				return nil, errors.New("connection refused")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockSupabaseClient{
				queryFunc: tt.mockFunc,
			}
			repo := newMockRepository(mock)

			ctx := context.Background()
			results, err := repo.Query(ctx, tt.table, tt.filters, tt.pagination)

			if (err != nil) != tt.wantErr {
				t.Errorf("Query() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && len(results) != tt.wantLen {
				t.Errorf("Query() returned %d results, want %d", len(results), tt.wantLen)
			}
		})
	}
}

func TestQueryTimeout(t *testing.T) {
	mock := &MockSupabaseClient{
		queryFunc: func(table string, filters map[string]interface{}, pagination Pagination) ([]map[string]interface{}, error) {
			return []map[string]interface{}{{"id": "1"}}, nil
		},
	}
	repo := newMockRepository(mock)

	// Create an already-expired context
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()
	time.Sleep(10 * time.Millisecond) // Ensure context is expired

	_, err := repo.Query(ctx, "products", map[string]interface{}{}, Pagination{})
	if err == nil {
		t.Error("Expected timeout error, got nil")
	}

	var repoErr *RepositoryError
	if !errors.As(err, &repoErr) {
		t.Errorf("Expected RepositoryError, got %T", err)
	}
}

func TestGetByID(t *testing.T) {
	tests := []struct {
		name     string
		table    string
		id       string
		mockFunc func(table string, id string) (map[string]interface{}, error)
		wantErr  bool
		wantID   string
	}{
		{
			name:  "successful get",
			table: "products",
			id:    "123",
			mockFunc: func(table string, id string) (map[string]interface{}, error) {
				return map[string]interface{}{
					"id":   "123",
					"name": "Test Product",
				}, nil
			},
			wantErr: false,
			wantID:  "123",
		},
		{
			name:  "not found error",
			table: "products",
			id:    "999",
			mockFunc: func(table string, id string) (map[string]interface{}, error) {
				return nil, errors.New("not found")
			},
			wantErr: true,
		},
		{
			name:  "connection error",
			table: "products",
			id:    "123",
			mockFunc: func(table string, id string) (map[string]interface{}, error) {
				return nil, errors.New("connection refused")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockSupabaseClient{
				getByIDFunc: tt.mockFunc,
			}
			repo := newMockRepository(mock)

			ctx := context.Background()
			result, err := repo.GetByID(ctx, tt.table, tt.id)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetByID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if id, ok := result["id"].(string); !ok || id != tt.wantID {
					t.Errorf("GetByID() returned id = %v, want %v", result["id"], tt.wantID)
				}
			}
		})
	}
}

func TestGetByIDTimeout(t *testing.T) {
	mock := &MockSupabaseClient{
		getByIDFunc: func(table string, id string) (map[string]interface{}, error) {
			return map[string]interface{}{"id": id}, nil
		},
	}
	repo := newMockRepository(mock)

	// Create an already-expired context
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()
	time.Sleep(10 * time.Millisecond) // Ensure context is expired

	_, err := repo.GetByID(ctx, "products", "123")
	if err == nil {
		t.Error("Expected timeout error, got nil")
	}

	var repoErr *RepositoryError
	if !errors.As(err, &repoErr) {
		t.Errorf("Expected RepositoryError, got %T", err)
	}
}

func TestRepositoryError(t *testing.T) {
	tests := []struct {
		name           string
		err            *RepositoryError
		wantStatusCode int
		wantMessage    string
	}{
		{
			name:           "connection error",
			err:            NewConnectionError(errors.New("connection failed")),
			wantStatusCode: 503,
			wantMessage:    "Failed to connect to Supabase",
		},
		{
			name:           "query error",
			err:            NewQueryError(errors.New("query failed")),
			wantStatusCode: 500,
			wantMessage:    "Failed to execute query",
		},
		{
			name:           "timeout error",
			err:            NewTimeoutError(errors.New("timeout")),
			wantStatusCode: 504,
			wantMessage:    "Request timeout",
		},
		{
			name:           "not found error",
			err:            NewNotFoundError("products", "123"),
			wantStatusCode: 404,
			wantMessage:    "Record not found in table products with id 123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.StatusCode != tt.wantStatusCode {
				t.Errorf("StatusCode = %d, want %d", tt.err.StatusCode, tt.wantStatusCode)
			}
			if tt.err.Message != tt.wantMessage {
				t.Errorf("Message = %s, want %s", tt.err.Message, tt.wantMessage)
			}
		})
	}
}

func TestGetStatusCode(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want int
	}{
		{
			name: "repository error",
			err:  NewConnectionError(errors.New("test")),
			want: 503,
		},
		{
			name: "regular error",
			err:  errors.New("regular error"),
			want: 500,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetStatusCode(tt.err); got != tt.want {
				t.Errorf("GetStatusCode() = %d, want %d", got, tt.want)
			}
		})
	}
}
