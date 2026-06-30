package validation

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"data-processing-pipeline/internal/models"
)

func ValidateRecord(record *models.Record, rules []models.ValidationRule) error {
	for _, rule := range rules {
		val, exists := record.Attributes[rule.Field]
		if !exists {
			if rule.Required {
				return fmt.Errorf("missing required field: %s", rule.Field)
			}
			continue
		}

		if val == nil || val == "" {
			if rule.Required {
				return fmt.Errorf("empty required field: %s", rule.Field)
			}
			continue
		}

		// Basic type checks
		strVal := fmt.Sprintf("%v", val)
		switch rule.Type {
		case "int":
			if _, err := strconv.Atoi(strVal); err != nil {
				return fmt.Errorf("field %s must be int", rule.Field)
			}
		case "float":
			if _, err := strconv.ParseFloat(strVal, 64); err != nil {
				return fmt.Errorf("field %s must be float", rule.Field)
			}
		case "boolean":
			if _, err := strconv.ParseBool(strVal); err != nil {
				return fmt.Errorf("field %s must be boolean", rule.Field)
			}
		case "date": // basic check for RFC3339 or similar date, simplified
			if _, err := time.Parse(time.RFC3339, strVal); err != nil {
				if _, err2 := time.Parse("2006-01-02", strVal); err2 != nil {
					return fmt.Errorf("field %s must be date format", rule.Field)
				}
			}
		}
	}
	return nil
}

func RunValidation(ctx context.Context, inCh <-chan models.Record, outCh chan<- models.Record, errCh chan<- *models.ErrorDetails, rules []models.ValidationRule, jobID string) {
	for {
		select {
		case <-ctx.Done():
			return
		case rec, ok := <-inCh:
			if !ok {
				return // channel closed
			}
			if err := ValidateRecord(&rec, rules); err != nil {
				errCh <- &models.ErrorDetails{
					JobID:      jobID,
					Stage:      "validation",
					Message:    err.Error(),
					RecordData: fmt.Sprintf("%v", rec.Attributes),
				}
				continue // skip invalid
			}
			outCh <- rec
		}
	}
}
