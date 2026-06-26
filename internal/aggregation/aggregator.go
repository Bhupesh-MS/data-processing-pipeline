package aggregation

import "data-processing-pipeline/internal/models"

func Aggregate(records []models.Record) models.Metrics {
	return models.Metrics{RecordsProcessed: len(records)}
}
