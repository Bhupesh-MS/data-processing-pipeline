package pipeline

import (
	"context"

	"data-processing-pipeline/internal/models"
)

type Pipeline struct {
	workers int
}

func New(workers int) *Pipeline {
	if workers <= 0 {
		workers = 1
	}
	return &Pipeline{workers: workers}
}

func (p *Pipeline) Run(ctx context.Context, records []models.Record) (*models.Metrics, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		return &models.Metrics{RecordsProcessed: len(records)}, nil
	}
}
