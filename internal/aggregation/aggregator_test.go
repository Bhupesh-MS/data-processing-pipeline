package aggregation

import (
	"context"
	"testing"

	"data-processing-pipeline/internal/models"
)

func TestRunAggregation(t *testing.T) {
	inCh := make(chan models.Record, 5)
	outCh := make(chan models.Record, 5)

	rules := []models.AggregationRule{
		{Type: "sum", Field: "sales", GroupBy: []string{"region"}},
		{Type: "count", Field: "sales", GroupBy: []string{"region"}},
		{Type: "avg", Field: "sales", GroupBy: []string{"region"}},
	}

	inCh <- models.Record{Attributes: map[string]any{"region": "US", "sales": 100.0}}
	inCh <- models.Record{Attributes: map[string]any{"region": "US", "sales": 200.0}}
	inCh <- models.Record{Attributes: map[string]any{"region": "EU", "sales": 150.0}}
	close(inCh)

	RunAggregation(context.Background(), inCh, outCh, rules)
	close(outCh)

	results := make(map[string]map[string]any)
	for rec := range outCh {
		group := rec.Attributes["group"].(string)
		results[group] = rec.Attributes
	}

	if len(results) != 2 {
		t.Errorf("expected 2 groups, got %d", len(results))
	}

	if results["US"]["sum_sales"] != 300.0 {
		t.Errorf("expected US sum_sales 300, got %v", results["US"]["sum_sales"])
	}
	if results["US"]["count_sales"] != 2.0 {
		t.Errorf("expected US count_sales 2, got %v", results["US"]["count_sales"])
	}
	if results["US"]["avg_sales"] != 150.0 {
		t.Errorf("expected US avg_sales 150, got %v", results["US"]["avg_sales"])
	}
}
