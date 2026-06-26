package validation

import (
	"data-processing-pipeline/internal/models"
	"data-processing-pipeline/internal/pipeline"
)

func Validate(record models.Record) error {
	if len(record.Fields) == 0 && len(record.Attributes) == 0 {
		return pipeline.ErrInvalidInput
	}
	return nil
}
