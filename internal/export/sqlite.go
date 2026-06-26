package export

import (
	"context"

	"data-processing-pipeline/internal/models"
)

type SQLiteExporter struct {
	path string
}

func NewSQLiteExporter(path string) *SQLiteExporter {
	return &SQLiteExporter{path: path}
}

func (e *SQLiteExporter) Export(ctx context.Context, records []models.Record) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return nil
	}
}
