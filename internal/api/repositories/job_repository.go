package repositories

import (
	"context"
	"database/sql"

	"data-processing-pipeline/internal/models"
)

type JobRepository interface {
	DB() *sql.DB
	CreateJob(ctx context.Context, job *models.Job) error
	GetJob(ctx context.Context, id string) (*models.Job, error)
	ListJobs(ctx context.Context) ([]*models.Job, error)
	UpdateJobStatus(ctx context.Context, id string, status models.JobStatus, errMsg string) error
	UpdateProgress(ctx context.Context, p *models.Progress) error
	GetProgress(ctx context.Context, jobID string) (*models.Progress, error)
	SaveError(ctx context.Context, errDetail *models.ErrorDetails) error
	GetErrors(ctx context.Context, jobID string) ([]*models.ErrorDetails, error)
	DeleteJob(ctx context.Context, jobID string) error
}
