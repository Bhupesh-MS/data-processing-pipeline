package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"data-processing-pipeline/internal/models"

	_ "github.com/mattn/go-sqlite3"
)

type SQLiteStore struct {
	db *sql.DB
}

func (s *SQLiteStore) DB() *sql.DB {
	return s.db
}

func NewSQLiteStore(dbPath string) (*SQLiteStore, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	store := &SQLiteStore{db: db}
	if err := store.migrate(); err != nil {
		return nil, err
	}

	return store, nil
}

func (s *SQLiteStore) migrate() error {
	query := `
	CREATE TABLE IF NOT EXISTS jobs (
		id TEXT PRIMARY KEY,
		status TEXT,
		spec TEXT,
		error TEXT,
		created_at DATETIME,
		updated_at DATETIME
	);
	CREATE TABLE IF NOT EXISTS progress (
		job_id TEXT PRIMARY KEY,
		status TEXT,
		percent_complete REAL,
		records_processed INTEGER,
		records_failed INTEGER,
		records_total INTEGER,
		processing_rate REAL,
		start_time DATETIME,
		end_time DATETIME
	);
	CREATE TABLE IF NOT EXISTS errors (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		job_id TEXT,
		stage TEXT,
		message TEXT,
		record_data TEXT,
		occurred_at DATETIME
	);
	`
	_, err := s.db.Exec(query)
	return err
}

func (s *SQLiteStore) Close() error {
	return s.db.Close()
}

func (s *SQLiteStore) CreateJob(ctx context.Context, job *models.Job) error {
	specJSON, _ := json.Marshal(job.Spec)
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO jobs (id, status, spec, error, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)`,
		job.ID, job.Status, string(specJSON), job.Error, job.CreatedAt, job.UpdatedAt,
	)
	if err != nil {
		return err
	}

	// Init progress
	_, err = s.db.ExecContext(ctx,
		`INSERT INTO progress (job_id, status, percent_complete, records_processed, records_failed, records_total, processing_rate, start_time) 
		 VALUES (?, ?, 0, 0, 0, 0, 0, ?)`,
		job.ID, job.Status, job.CreatedAt,
	)
	return err
}

func (s *SQLiteStore) GetJob(ctx context.Context, id string) (*models.Job, error) {
	row := s.db.QueryRowContext(ctx, `SELECT id, status, spec, error, created_at, updated_at FROM jobs WHERE id = ?`, id)

	var job models.Job
	var specStr string
	err := row.Scan(&job.ID, &job.Status, &specStr, &job.Error, &job.CreatedAt, &job.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Not found
		}
		return nil, err
	}

	json.Unmarshal([]byte(specStr), &job.Spec)
	return &job, nil
}

func (s *SQLiteStore) ListJobs(ctx context.Context) ([]*models.Job, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, status, spec, error, created_at, updated_at FROM jobs ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jobs []*models.Job
	for rows.Next() {
		var job models.Job
		var specStr string
		if err := rows.Scan(&job.ID, &job.Status, &specStr, &job.Error, &job.CreatedAt, &job.UpdatedAt); err != nil {
			return nil, err
		}
		json.Unmarshal([]byte(specStr), &job.Spec)
		jobs = append(jobs, &job)
	}
	return jobs, nil
}

func (s *SQLiteStore) UpdateJobStatus(ctx context.Context, id string, status models.JobStatus, errMsg string) error {
	_, err := s.db.ExecContext(ctx, `UPDATE jobs SET status = ?, error = ?, updated_at = ? WHERE id = ?`, status, errMsg, time.Now(), id)
	return err
}

func (s *SQLiteStore) UpdateProgress(ctx context.Context, p *models.Progress) error {
	var endTime *time.Time
	if p.EndTime != nil {
		endTime = p.EndTime
	}
	_, err := s.db.ExecContext(ctx,
		`UPDATE progress SET status = ?, percent_complete = ?, records_processed = ?, records_failed = ?, records_total = ?, processing_rate = ?, end_time = ? WHERE job_id = ?`,
		p.Status, p.PercentComplete, p.Metrics.RecordsProcessed, p.Metrics.RecordsFailed, p.Metrics.RecordsTotal, p.Metrics.ProcessingRate, endTime, p.JobID,
	)
	return err
}

func (s *SQLiteStore) GetProgress(ctx context.Context, jobID string) (*models.Progress, error) {
	row := s.db.QueryRowContext(ctx, `SELECT job_id, status, percent_complete, records_processed, records_failed, records_total, processing_rate, start_time, end_time FROM progress WHERE job_id = ?`, jobID)

	var p models.Progress
	var endTime sql.NullTime
	err := row.Scan(&p.JobID, &p.Status, &p.PercentComplete, &p.Metrics.RecordsProcessed, &p.Metrics.RecordsFailed, &p.Metrics.RecordsTotal, &p.Metrics.ProcessingRate, &p.StartTime, &endTime)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	if endTime.Valid {
		p.EndTime = &endTime.Time
	}
	return &p, nil
}

func (s *SQLiteStore) SaveError(ctx context.Context, errDetail *models.ErrorDetails) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO errors (job_id, stage, message, record_data, occurred_at) VALUES (?, ?, ?, ?, ?)`,
		errDetail.JobID, errDetail.Stage, errDetail.Message, errDetail.RecordData, errDetail.OccurredAt,
	)
	return err
}

func (s *SQLiteStore) GetErrors(ctx context.Context, jobID string) ([]*models.ErrorDetails, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT job_id, stage, message, record_data, occurred_at FROM errors WHERE job_id = ? ORDER BY occurred_at DESC`, jobID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var errorsList []*models.ErrorDetails
	for rows.Next() {
		var e models.ErrorDetails
		if err := rows.Scan(&e.JobID, &e.Stage, &e.Message, &e.RecordData, &e.OccurredAt); err != nil {
			return nil, err
		}
		errorsList = append(errorsList, &e)
	}
	return errorsList, nil
}

func (s *SQLiteStore) DeleteJob(ctx context.Context, jobID string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM jobs WHERE id = ?`, jobID)
	if err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx, `DELETE FROM progress WHERE job_id = ?`, jobID)
	if err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx, `DELETE FROM errors WHERE job_id = ?`, jobID)
	return err
}
