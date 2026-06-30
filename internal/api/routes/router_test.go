package routes

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"data-processing-pipeline/internal/api/controllers"
	"data-processing-pipeline/internal/models"
)

type mockJobService struct{}

func (m *mockJobService) CreateJob(ctx context.Context, spec models.JobSpec) (*models.Job, error) {
	return nil, nil
}
func (m *mockJobService) ListJobs(ctx context.Context) ([]*models.Job, error) {
	return nil, nil
}
func (m *mockJobService) GetJob(ctx context.Context, id string) (*models.Job, error) {
	return nil, nil
}
func (m *mockJobService) GetJobProgress(ctx context.Context, id string) (*models.Progress, error) {
	return nil, nil
}
func (m *mockJobService) GetJobErrors(ctx context.Context, id string) ([]*models.ErrorDetails, error) {
	return nil, nil
}
func (m *mockJobService) CancelJob(ctx context.Context, id string) error {
	return nil
}
func (m *mockJobService) DeleteJob(ctx context.Context, id string) error {
	return nil
}

func TestRouter_Routes(t *testing.T) {
	svc := &mockJobService{}
	ctrl := controllers.NewJobController(svc)
	router := NewRouter(ctrl)

	// Test health route to ensure router is wired
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("expected %v, got %v", http.StatusOK, rr.Code)
	}
}

func TestLoggingMiddleware(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	middleware := LoggingMiddleware(handler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rr := httptest.NewRecorder()

	middleware.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}
