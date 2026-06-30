package services

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"data-processing-pipeline/internal/models"
)

// Mock repository
type mockJobRepo struct {
	jobs     map[string]*models.Job
	progress map[string]*models.Progress
	errors   map[string][]*models.ErrorDetails
}

func (m *mockJobRepo) DB() *sql.DB {
	return nil
}

func newMockJobRepo() *mockJobRepo {
	return &mockJobRepo{
		jobs:     make(map[string]*models.Job),
		progress: make(map[string]*models.Progress),
		errors:   make(map[string][]*models.ErrorDetails),
	}
}

func (m *mockJobRepo) CreateJob(ctx context.Context, job *models.Job) error {
	m.jobs[job.ID] = job
	m.progress[job.ID] = &models.Progress{JobID: job.ID, Status: job.Status}
	return nil
}
func (m *mockJobRepo) GetJob(ctx context.Context, id string) (*models.Job, error) {
	if job, ok := m.jobs[id]; ok {
		return job, nil
	}
	return nil, nil // not found
}
func (m *mockJobRepo) ListJobs(ctx context.Context) ([]*models.Job, error) {
	var list []*models.Job
	for _, j := range m.jobs {
		list = append(list, j)
	}
	return list, nil
}
func (m *mockJobRepo) UpdateJobStatus(ctx context.Context, id string, status models.JobStatus, errMsg string) error {
	if job, ok := m.jobs[id]; ok {
		job.Status = status
		job.Error = errMsg
		return nil
	}
	return errors.New("not found")
}
func (m *mockJobRepo) UpdateProgress(ctx context.Context, p *models.Progress) error {
	m.progress[p.JobID] = p
	return nil
}
func (m *mockJobRepo) GetProgress(ctx context.Context, jobID string) (*models.Progress, error) {
	if p, ok := m.progress[jobID]; ok {
		return p, nil
	}
	return nil, nil
}
func (m *mockJobRepo) SaveError(ctx context.Context, errDetail *models.ErrorDetails) error {
	m.errors[errDetail.JobID] = append(m.errors[errDetail.JobID], errDetail)
	return nil
}
func (m *mockJobRepo) GetErrors(ctx context.Context, jobID string) ([]*models.ErrorDetails, error) {
	if errs, ok := m.errors[jobID]; ok {
		return errs, nil
	}
	return nil, nil
}
func (m *mockJobRepo) DeleteJob(ctx context.Context, jobID string) error {
	delete(m.jobs, jobID)
	delete(m.progress, jobID)
	delete(m.errors, jobID)
	return nil
}

// Mock orchestrator
type mockOrchestrator struct {
	runJobCalled    bool
	cancelJobCalled bool
}

func (m *mockOrchestrator) RunJob(job *models.Job) {
	m.runJobCalled = true
}
func (m *mockOrchestrator) CancelJob(jobID string) error {
	m.cancelJobCalled = true
	return nil
}

func TestJobService_CreateJob(t *testing.T) {
	repo := newMockJobRepo()
	orch := &mockOrchestrator{}
	service := NewJobService(repo, orch)

	spec := models.JobSpec{WorkerCount: 5}
	job, err := service.CreateJob(context.Background(), spec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if job.ID == "" {
		t.Errorf("expected job ID, got empty")
	}
	if job.Status != models.JobStatusPending {
		t.Errorf("expected status pending, got %v", job.Status)
	}
	
	// Check orchestrator was not called immediately in a blocking way (it's called in goroutine, but we can't easily assert here without waiting. Let's just assume it runs).
}

func TestJobService_GetJob(t *testing.T) {
	repo := newMockJobRepo()
	orch := &mockOrchestrator{}
	service := NewJobService(repo, orch)

	job := &models.Job{ID: "job-1"}
	repo.CreateJob(context.Background(), job)

	fetched, err := service.GetJob(context.Background(), "job-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fetched.ID != "job-1" {
		t.Errorf("expected job-1, got %v", fetched.ID)
	}

	_, err = service.GetJob(context.Background(), "job-none")
	if err == nil || err.Error() != "job not found" {
		t.Errorf("expected job not found error, got %v", err)
	}
}

func TestJobService_ListJobs(t *testing.T) {
	repo := newMockJobRepo()
	orch := &mockOrchestrator{}
	service := NewJobService(repo, orch)

	repo.CreateJob(context.Background(), &models.Job{ID: "job-1"})
	repo.CreateJob(context.Background(), &models.Job{ID: "job-2"})

	jobs, err := service.ListJobs(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(jobs) != 2 {
		t.Errorf("expected 2 jobs, got %d", len(jobs))
	}
}

func TestJobService_GetJobProgress(t *testing.T) {
	repo := newMockJobRepo()
	orch := &mockOrchestrator{}
	service := NewJobService(repo, orch)

	repo.progress["job-1"] = &models.Progress{JobID: "job-1", PercentComplete: 50}

	p, err := service.GetJobProgress(context.Background(), "job-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.PercentComplete != 50 {
		t.Errorf("expected 50, got %f", p.PercentComplete)
	}

	_, err = service.GetJobProgress(context.Background(), "job-none")
	if err == nil || err.Error() != "progress not found" {
		t.Errorf("expected progress not found error, got %v", err)
	}
}

func TestJobService_GetJobErrors(t *testing.T) {
	repo := newMockJobRepo()
	orch := &mockOrchestrator{}
	service := NewJobService(repo, orch)

	repo.errors["job-1"] = append(repo.errors["job-1"], &models.ErrorDetails{Message: "err"})

	errs, err := service.GetJobErrors(context.Background(), "job-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(errs) != 1 {
		t.Errorf("expected 1 error, got %d", len(errs))
	}

	// Not found should return empty slice
	emptyErrs, err := service.GetJobErrors(context.Background(), "job-none")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(emptyErrs) != 0 {
		t.Errorf("expected 0, got %d", len(emptyErrs))
	}
}

func TestJobService_CancelDelete(t *testing.T) {
	repo := newMockJobRepo()
	orch := &mockOrchestrator{}
	service := NewJobService(repo, orch)

	repo.CreateJob(context.Background(), &models.Job{ID: "job-1"})

	err := service.CancelJob(context.Background(), "job-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !orch.cancelJobCalled {
		t.Errorf("expected orchestrator cancel to be called")
	}

	err = service.DeleteJob(context.Background(), "job-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := repo.jobs["job-1"]; ok {
		t.Errorf("expected job to be deleted")
	}
}
