package models

import "time"

type JobStatus string

const (
	JobStatusPending   JobStatus = "pending"
	JobStatusRunning   JobStatus = "running"
	JobStatusCompleted JobStatus = "completed"
	JobStatusFailed    JobStatus = "failed"
	JobStatusCancelled JobStatus = "cancelled"
)

type SourceConfig struct {
	Type   string `json:"type"`             // "csv", "json", "api"
	URL    string `json:"url,omitempty"`    // for http downloads/api
	Path   string `json:"path,omitempty"`   // for local files
	Method string `json:"method,omitempty"` // for api
}

type ValidationRule struct {
	Field string `json:"field"`
	Type  string `json:"type"` // "string", "int", "float", "boolean", "date"
	Required bool `json:"required"`
}

type TransformRule struct {
	Field    string `json:"field"`
	Action   string `json:"action"` // "lowercase", "uppercase", "trim", "to_int", "to_float"
	Target   string `json:"target,omitempty"` // target field if renaming/copying
}

type AggregationRule struct {
	Type   string `json:"type"` // "count", "sum", "avg", "min", "max"
	Field  string `json:"field"`
	GroupBy []string `json:"group_by,omitempty"`
}

type ExportConfig struct {
	Type   string `json:"type"` // "sqlite", "csv", "json"
	Path   string `json:"path,omitempty"`
	Table  string `json:"table,omitempty"`
}

type JobSpec struct {
	Sources         []SourceConfig    `json:"sources"`
	Validations     []ValidationRule  `json:"validations,omitempty"`
	Transformations []TransformRule   `json:"transformations,omitempty"`
	Aggregations    []AggregationRule `json:"aggregations,omitempty"`
	Exports         []ExportConfig    `json:"exports"`
	WorkerCount     int               `json:"worker_count,omitempty"` // Default 5
}

type Job struct {
	ID        string    `json:"id"`
	Status    JobStatus `json:"status"`
	Spec      JobSpec   `json:"spec"`
	Error     string    `json:"error,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
