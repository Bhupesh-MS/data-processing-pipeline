package controllers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"data-processing-pipeline/internal/models"
)

type mockJobService struct {
	createJobErr error
	listJobsErr  error
	getJobErr    error
	progressErr  error
	errorsErr    error
	cancelErr    error
	deleteErr    error
}

func (m *mockJobService) CreateJob(ctx context.Context, spec models.JobSpec) (*models.Job, error) {
	if m.createJobErr != nil {
		return nil, m.createJobErr
	}
	return &models.Job{ID: "job-1", Status: models.JobStatusPending}, nil
}
func (m *mockJobService) ListJobs(ctx context.Context) ([]*models.Job, error) {
	if m.listJobsErr != nil {
		return nil, m.listJobsErr
	}
	return []*models.Job{{ID: "job-1"}}, nil
}
func (m *mockJobService) GetJob(ctx context.Context, id string) (*models.Job, error) {
	if m.getJobErr != nil {
		return nil, m.getJobErr
	}
	return &models.Job{ID: id}, nil
}
func (m *mockJobService) GetJobProgress(ctx context.Context, id string) (*models.Progress, error) {
	if m.progressErr != nil {
		return nil, m.progressErr
	}
	return &models.Progress{JobID: id, PercentComplete: 100}, nil
}
func (m *mockJobService) GetJobErrors(ctx context.Context, id string) ([]*models.ErrorDetails, error) {
	if m.errorsErr != nil {
		return nil, m.errorsErr
	}
	return []*models.ErrorDetails{}, nil
}
func (m *mockJobService) CancelJob(ctx context.Context, id string) error {
	return m.cancelErr
}
func (m *mockJobService) DeleteJob(ctx context.Context, id string) error {
	return m.deleteErr
}

func TestJobController_Health(t *testing.T) {
	svc := &mockJobService{}
	ctrl := NewJobController(svc)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()

	ctrl.Health(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("expected status %v, got %v", http.StatusOK, status)
	}
}

func TestJobController_CreateJob(t *testing.T) {
	svc := &mockJobService{}
	ctrl := NewJobController(svc)

	// Valid request
	spec := models.JobSpec{WorkerCount: 2}
	body, _ := json.Marshal(spec)
	req := httptest.NewRequest(http.MethodPost, "/jobs", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()
	ctrl.CreateJob(rr, req)

	if status := rr.Code; status != http.StatusAccepted {
		t.Errorf("expected %v, got %v", http.StatusAccepted, status)
	}

	// Invalid JSON
	req = httptest.NewRequest(http.MethodPost, "/jobs", bytes.NewBuffer([]byte("{invalid")))
	rr = httptest.NewRecorder()
	ctrl.CreateJob(rr, req)
	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("expected %v, got %v", http.StatusBadRequest, status)
	}

	// Service error
	svc.createJobErr = errors.New("db error")
	req = httptest.NewRequest(http.MethodPost, "/jobs", bytes.NewBuffer(body))
	rr = httptest.NewRecorder()
	ctrl.CreateJob(rr, req)
	if status := rr.Code; status != http.StatusInternalServerError {
		t.Errorf("expected %v, got %v", http.StatusInternalServerError, status)
	}
}

func TestJobController_GetJob(t *testing.T) {
	svc := &mockJobService{}
	ctrl := NewJobController(svc)

	req := httptest.NewRequest(http.MethodGet, "/jobs/job-1", nil)
	req.SetPathValue("id", "job-1")
	rr := httptest.NewRecorder()
	ctrl.GetJob(rr, req)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("expected %v, got %v", http.StatusOK, status)
	}

	svc.getJobErr = errors.New("job not found")
	rr = httptest.NewRecorder()
	ctrl.GetJob(rr, req)
	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("expected %v, got %v", http.StatusNotFound, status)
	}

	svc.getJobErr = errors.New("db err")
	rr = httptest.NewRecorder()
	ctrl.GetJob(rr, req)
	if status := rr.Code; status != http.StatusInternalServerError {
		t.Errorf("expected %v, got %v", http.StatusInternalServerError, status)
	}
}

func TestJobController_ListJobs(t *testing.T) {
	svc := &mockJobService{}
	ctrl := NewJobController(svc)

	req := httptest.NewRequest(http.MethodGet, "/jobs", nil)
	rr := httptest.NewRecorder()
	ctrl.ListJobs(rr, req)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("expected %v, got %v", http.StatusOK, status)
	}
	
	svc.listJobsErr = errors.New("err")
	rr = httptest.NewRecorder()
	ctrl.ListJobs(rr, req)
	if status := rr.Code; status != http.StatusInternalServerError {
		t.Errorf("expected %v, got %v", http.StatusInternalServerError, status)
	}
}

func TestJobController_GetJobProgress(t *testing.T) {
	svc := &mockJobService{}
	ctrl := NewJobController(svc)

	req := httptest.NewRequest(http.MethodGet, "/jobs/1/progress", nil)
	req.SetPathValue("id", "1")
	rr := httptest.NewRecorder()
	ctrl.GetJobProgress(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("expected %v, got %v", http.StatusOK, rr.Code)
	}

	svc.progressErr = errors.New("progress not found")
	rr = httptest.NewRecorder()
	ctrl.GetJobProgress(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Errorf("expected %v, got %v", http.StatusNotFound, rr.Code)
	}
}

func TestJobController_GetJobErrors(t *testing.T) {
	svc := &mockJobService{}
	ctrl := NewJobController(svc)

	req := httptest.NewRequest(http.MethodGet, "/jobs/1/errors", nil)
	rr := httptest.NewRecorder()
	ctrl.GetJobErrors(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("expected %v, got %v", http.StatusOK, rr.Code)
	}

	svc.errorsErr = errors.New("db err")
	rr = httptest.NewRecorder()
	ctrl.GetJobErrors(rr, req)
	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected %v, got %v", http.StatusInternalServerError, rr.Code)
	}
}

func TestJobController_CancelJob(t *testing.T) {
	svc := &mockJobService{}
	ctrl := NewJobController(svc)

	req := httptest.NewRequest(http.MethodPatch, "/jobs/1/cancel", nil)
	rr := httptest.NewRecorder()
	ctrl.CancelJob(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("expected %v, got %v", http.StatusOK, rr.Code)
	}

	svc.cancelErr = errors.New("not found")
	rr = httptest.NewRecorder()
	ctrl.CancelJob(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Errorf("expected %v, got %v", http.StatusNotFound, rr.Code)
	}
}

func TestJobController_DeleteJob(t *testing.T) {
	svc := &mockJobService{}
	ctrl := NewJobController(svc)

	req := httptest.NewRequest(http.MethodDelete, "/jobs/1", nil)
	rr := httptest.NewRecorder()
	ctrl.DeleteJob(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("expected %v, got %v", http.StatusOK, rr.Code)
	}

	svc.deleteErr = errors.New("err")
	rr = httptest.NewRecorder()
	ctrl.DeleteJob(rr, req)
	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected %v, got %v", http.StatusInternalServerError, rr.Code)
	}
}

func TestJobController_GetJobResults(t *testing.T) {
	svc := &mockJobService{}
	ctrl := NewJobController(svc)

	req := httptest.NewRequest(http.MethodGet, "/jobs/1/results", nil)
	rr := httptest.NewRecorder()
	ctrl.GetJobResults(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("expected %v, got %v", http.StatusOK, rr.Code)
	}
}
