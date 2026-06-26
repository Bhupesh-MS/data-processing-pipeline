package export

import (
	"context"
	"encoding/json"
	"io"

	"data-processing-pipeline/internal/models"
)

func WriteJSON(ctx context.Context, w io.Writer, records []models.Record) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return json.NewEncoder(w).Encode(records)
	}
}
