package services

import (
	"context"
	"errors"
	"time"

	"data-processing-pipeline/internal/api/repositories"
	"data-processing-pipeline/internal/models"
)

type JobOrchestrator interface {
	RunJob(job *models.Job)
	CancelJob(jobID string) error
}

type JobService interface {
	CreateJob(ctx context.Context, spec models.JobSpec) (*models.Job, error)
	ListJobs(ctx context.Context) ([]*models.Job, error)
	GetJob(ctx context.Context, id string) (*models.Job, error)
	GetJobProgress(ctx context.Context, id string) (*models.Progress, error)
	GetJobErrors(ctx context.Context, id string) ([]*models.ErrorDetails, error)
	CancelJob(ctx context.Context, id string) error
	DeleteJob(ctx context.Context, id string) error
}

type jobService struct {
	repo         repositories.JobRepository
	orchestrator JobOrchestrator
}

func NewJobService(repo repositories.JobRepository, orchestrator JobOrchestrator) JobService {
	return &jobService{
		repo:         repo,
		orchestrator: orchestrator,
	}
}

func (s *jobService) CreateJob(ctx context.Context, spec models.JobSpec) (*models.Job, error) {
	jobID := time.Now().Format("20060102150405") + "-" + randomString(6)
	job := &models.Job{
		ID:        jobID,
		Status:    models.JobStatusPending,
		Spec:      spec,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.repo.CreateJob(ctx, job); err != nil {
		return nil, err
	}

	// Start pipeline in background
	go s.orchestrator.RunJob(job)

	return job, nil
}

func (s *jobService) ListJobs(ctx context.Context) ([]*models.Job, error) {
	jobs, err := s.repo.ListJobs(ctx)
	if err != nil {
		return nil, err
	}
	if jobs == nil {
		jobs = []*models.Job{}
	}
	return jobs, nil
}

func (s *jobService) GetJob(ctx context.Context, id string) (*models.Job, error) {
	job, err := s.repo.GetJob(ctx, id)
	if err != nil {
		return nil, err
	}
	if job == nil {
		return nil, errors.New("job not found")
	}
	return job, nil
}

func (s *jobService) GetJobProgress(ctx context.Context, id string) (*models.Progress, error) {
	progress, err := s.repo.GetProgress(ctx, id)
	if err != nil {
		return nil, err
	}
	if progress == nil {
		return nil, errors.New("progress not found")
	}
	return progress, nil
}

func (s *jobService) GetJobErrors(ctx context.Context, id string) ([]*models.ErrorDetails, error) {
	errorsList, err := s.repo.GetErrors(ctx, id)
	if err != nil {
		return nil, err
	}
	if errorsList == nil {
		errorsList = []*models.ErrorDetails{}
	}
	return errorsList, nil
}

func (s *jobService) CancelJob(ctx context.Context, id string) error {
	return s.orchestrator.CancelJob(id)
}

func (s *jobService) DeleteJob(ctx context.Context, id string) error {
	return s.repo.DeleteJob(ctx, id)
}

func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[time.Now().UnixNano()%int64(len(letters))]
	}
	return string(b)
}
