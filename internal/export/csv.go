package export

import (
	"context"
	"encoding/csv"
	"io"

	"data-processing-pipeline/internal/models"
)

func WriteCSV(ctx context.Context, w io.Writer, records []models.Record) error {
	writer := csv.NewWriter(w)
	defer writer.Flush()

	for _, record := range records {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if err := writer.Write(record.Fields); err != nil {
				return err
			}
		}
	}
	return writer.Error()
}
