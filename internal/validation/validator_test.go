package validation

import (
	"context"
	"testing"

	"data-processing-pipeline/internal/models"
)

func TestValidateRecord(t *testing.T) {
	tests := []struct {
		name    string
		record  *models.Record
		rules   []models.ValidationRule
		wantErr bool
	}{
		{
			name:   "valid record",
			record: &models.Record{Attributes: map[string]any{"age": 25, "name": "John"}},
			rules: []models.ValidationRule{
				{Field: "age", Type: "int", Required: true},
				{Field: "name", Type: "string", Required: true},
			},
			wantErr: false,
		},
		{
			name:   "missing required field",
			record: &models.Record{Attributes: map[string]any{"name": "John"}},
			rules: []models.ValidationRule{
				{Field: "age", Type: "int", Required: true},
			},
			wantErr: true,
		},
		{
			name:   "invalid type",
			record: &models.Record{Attributes: map[string]any{"age": "twenty"}},
			rules: []models.ValidationRule{
				{Field: "age", Type: "int", Required: true},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ValidateRecord(tt.record, tt.rules); (err != nil) != tt.wantErr {
				t.Errorf("ValidateRecord() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRunValidation(t *testing.T) {
	inCh := make(chan models.Record, 1)
	outCh := make(chan models.Record, 1)
	errCh := make(chan *models.ErrorDetails, 1)

	rules := []models.ValidationRule{{Field: "age", Type: "int", Required: true}}

	inCh <- models.Record{Attributes: map[string]any{"age": 30}}
	close(inCh)

	RunValidation(context.Background(), inCh, outCh, errCh, rules, "job-1")

	close(outCh)
	close(errCh)

	validCount := 0
	for range outCh {
		validCount++
	}

	if validCount != 1 {
		t.Errorf("expected 1 valid record, got %d", validCount)
	}
}
