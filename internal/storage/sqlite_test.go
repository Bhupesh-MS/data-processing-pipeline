package storage

import (
	"context"
	"testing"
	"time"

	"data-processing-pipeline/internal/models"
)

func TestSQLiteStore_CRUD(t *testing.T) {
	// Use in-memory DB for tests
	store, err := NewSQLiteStore("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer store.Close()

	ctx := context.Background()

	// 1. Create Job
	job := &models.Job{
		ID:        "test-job-1",
		Status:    models.JobStatusPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Spec:      models.JobSpec{WorkerCount: 1},
	}

	err = store.CreateJob(ctx, job)
	if err != nil {
		t.Fatalf("CreateJob failed: %v", err)
	}

	// 2. Get Job
	fetched, err := store.GetJob(ctx, "test-job-1")
	if err != nil {
		t.Fatalf("GetJob failed: %v", err)
	}
	if fetched == nil || fetched.ID != job.ID {
		t.Errorf("GetJob mismatch, expected %v got %v", job.ID, fetched)
	}

	// 3. List Jobs
	jobs, err := store.ListJobs(ctx)
	if err != nil {
		t.Fatalf("ListJobs failed: %v", err)
	}
	if len(jobs) != 1 {
		t.Errorf("expected 1 job, got %v", len(jobs))
	}

	// 4. Update Job Status
	err = store.UpdateJobStatus(ctx, "test-job-1", models.JobStatusRunning, "")
	if err != nil {
		t.Fatalf("UpdateJobStatus failed: %v", err)
	}

	fetched, _ = store.GetJob(ctx, "test-job-1")
	if fetched.Status != models.JobStatusRunning {
		t.Errorf("expected status running, got %v", fetched.Status)
	}

	// 5. Progress
	prog, err := store.GetProgress(ctx, "test-job-1")
	if err != nil {
		t.Fatalf("GetProgress failed: %v", err)
	}
	if prog == nil {
		t.Fatalf("expected progress record to be created with job")
	}

	prog.PercentComplete = 50.0
	prog.Metrics.RecordsProcessed = 100
	err = store.UpdateProgress(ctx, prog)
	if err != nil {
		t.Fatalf("UpdateProgress failed: %v", err)
	}

	progFetched, _ := store.GetProgress(ctx, "test-job-1")
	if progFetched.PercentComplete != 50.0 {
		t.Errorf("expected 50 percent, got %v", progFetched.PercentComplete)
	}

	// 6. Errors
	errDetail := &models.ErrorDetails{
		JobID:      "test-job-1",
		Stage:      "validation",
		Message:    "invalid format",
		OccurredAt: time.Now().Format(time.RFC3339),
	}
	err = store.SaveError(ctx, errDetail)
	if err != nil {
		t.Fatalf("SaveError failed: %v", err)
	}

	errorsList, err := store.GetErrors(ctx, "test-job-1")
	if err != nil {
		t.Fatalf("GetErrors failed: %v", err)
	}
	if len(errorsList) != 1 {
		t.Errorf("expected 1 error, got %v", len(errorsList))
	}

	// 7. Delete Job
	err = store.DeleteJob(ctx, "test-job-1")
	if err != nil {
		t.Fatalf("DeleteJob failed: %v", err)
	}

	fetched, _ = store.GetJob(ctx, "test-job-1")
	if fetched != nil {
		t.Errorf("expected job to be deleted")
	}
}
