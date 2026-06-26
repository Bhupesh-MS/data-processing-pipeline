package ingestion

import (
	"context"
	"encoding/json"
	"io"

	"data-processing-pipeline/internal/models"
)

func ReadJSON(ctx context.Context, r io.Reader) ([]models.Record, error) {
	var records []models.Record
	if err := json.NewDecoder(r).Decode(&records); err != nil {
		return nil, err
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		return records, nil
	}
}
