package models

import "time"

type Metrics struct {
	RecordsProcessed int     `json:"records_processed"`
	RecordsFailed    int     `json:"records_failed"`
	RecordsTotal     int     `json:"records_total,omitempty"`
	ProcessingRate   float64 `json:"processing_rate"` // records per second
}

type Progress struct {
	JobID           string     `json:"job_id"`
	Status          JobStatus  `json:"status"`
	PercentComplete float64    `json:"percent_complete"`
	Metrics         Metrics    `json:"metrics"`
	StartTime       time.Time  `json:"start_time"`
	EndTime         *time.Time `json:"end_time,omitempty"`
}

type ErrorDetails struct {
	JobID      string `json:"job_id"`
	Stage      string `json:"stage"`
	Message    string `json:"message"`
	RecordData string `json:"record_data,omitempty"`
	OccurredAt string `json:"occurred_at"`
}
