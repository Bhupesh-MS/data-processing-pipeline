package transformation

import (
	"context"
	"testing"

	"data-processing-pipeline/internal/models"
)

func TestTransformRecord(t *testing.T) {
	tests := []struct {
		name    string
		record  *models.Record
		rules   []models.TransformRule
		wantErr bool
		verify  func(*models.Record) bool
	}{
		{
			name:    "lowercase",
			record:  &models.Record{Attributes: map[string]any{"name": "JOHN"}},
			rules:   []models.TransformRule{{Field: "name", Action: "lowercase"}},
			wantErr: false,
			verify:  func(r *models.Record) bool { return r.Attributes["name"] == "john" },
		},
		{
			name:    "to_int",
			record:  &models.Record{Attributes: map[string]any{"age": "25"}},
			rules:   []models.TransformRule{{Field: "age", Action: "to_int"}},
			wantErr: false,
			verify:  func(r *models.Record) bool { return r.Attributes["age"] == 25 },
		},
		{
			name:    "target_rename",
			record:  &models.Record{Attributes: map[string]any{"age": "25"}},
			rules:   []models.TransformRule{{Field: "age", Action: "to_int", Target: "age_int"}},
			wantErr: false,
			verify:  func(r *models.Record) bool { return r.Attributes["age_int"] == 25 },
		},
		{
			name:    "to_int error",
			record:  &models.Record{Attributes: map[string]any{"age": "twenty"}},
			rules:   []models.TransformRule{{Field: "age", Action: "to_int"}},
			wantErr: true,
			verify:  func(r *models.Record) bool { return true },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := TransformRecord(tt.record, tt.rules)
			if (err != nil) != tt.wantErr {
				t.Errorf("TransformRecord() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && !tt.verify(tt.record) {
				t.Errorf("TransformRecord() verify failed for %v", tt.record)
			}
		})
	}
}

func TestRunTransformation(t *testing.T) {
	inCh := make(chan models.Record, 1)
	outCh := make(chan models.Record, 1)
	errCh := make(chan *models.ErrorDetails, 1)

	rules := []models.TransformRule{{Field: "name", Action: "lowercase"}}

	inCh <- models.Record{Attributes: map[string]any{"name": "JOHN"}}
	close(inCh)

	RunTransformation(context.Background(), inCh, outCh, errCh, rules, "job-1")

	close(outCh)
	close(errCh)

	rec := <-outCh
	if rec.Attributes["name"] != "john" {
		t.Errorf("expected john, got %v", rec.Attributes["name"])
	}
}
