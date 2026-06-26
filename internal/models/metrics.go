package models

type Metrics struct {
	RecordsProcessed int `json:"records_processed"`
	RecordsFailed    int `json:"records_failed"`
}
