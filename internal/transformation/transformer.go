package transformation

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"data-processing-pipeline/internal/models"
)

func TransformRecord(record *models.Record, rules []models.TransformRule) error {
	for _, rule := range rules {
		val, exists := record.Attributes[rule.Field]
		if !exists || val == nil {
			continue // skip if missing
		}

		strVal := fmt.Sprintf("%v", val)
		var newVal any = val

		switch rule.Action {
		case "lowercase":
			newVal = strings.ToLower(strVal)
		case "uppercase":
			newVal = strings.ToUpper(strVal)
		case "trim":
			newVal = strings.TrimSpace(strVal)
		case "to_int":
			if v, err := strconv.Atoi(strVal); err == nil {
				newVal = v
			} else {
				return fmt.Errorf("failed to cast %s to int", rule.Field)
			}
		case "to_float":
			if v, err := strconv.ParseFloat(strVal, 64); err == nil {
				newVal = v
			} else {
				return fmt.Errorf("failed to cast %s to float", rule.Field)
			}
		}

		target := rule.Target
		if target == "" {
			target = rule.Field
		}
		record.Attributes[target] = newVal
	}
	return nil
}

func RunTransformation(ctx context.Context, inCh <-chan models.Record, outCh chan<- models.Record, errCh chan<- *models.ErrorDetails, rules []models.TransformRule, jobID string) {
	for {
		select {
		case <-ctx.Done():
			return
		case rec, ok := <-inCh:
			if !ok {
				return
			}
			if err := TransformRecord(&rec, rules); err != nil {
				errCh <- &models.ErrorDetails{
					JobID:      jobID,
					Stage:      "transformation",
					Message:    err.Error(),
					RecordData: fmt.Sprintf("%v", rec.Attributes),
				}
				continue
			}
			outCh <- rec
		}
	}
}
