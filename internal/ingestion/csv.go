package ingestion

import (
	"context"
	"encoding/csv"
	"io"

	"data-processing-pipeline/internal/models"
)

func ReadCSV(ctx context.Context, r io.Reader) ([]models.Record, error) {
	reader := csv.NewReader(r)
	rows, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	records := make([]models.Record, 0, len(rows))
	for _, row := range rows {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			records = append(records, models.Record{Fields: row})
		}
	}
	return records, nil
}
