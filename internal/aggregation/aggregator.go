package aggregation

import (
	"context"
	"fmt"
	"strings"

	"data-processing-pipeline/internal/models"
)

func RunAggregation(ctx context.Context, inCh <-chan models.Record, outCh chan<- models.Record, rules []models.AggregationRule) {
	// We will hold state for all aggregations
	// map of string (group by key) to map of metric to value
	state := make(map[string]map[string]float64)

	// Since we need counts separately from sum for avg
	counts := make(map[string]map[string]int)

	for {
		select {
		case <-ctx.Done():
			return
		case rec, ok := <-inCh:
			if !ok {
				// Channel closed, process state and send to outCh
				sendAggregations(outCh, state, counts, rules)
				return
			}
			processRecord(&rec, rules, state, counts)
		}
	}
}

func processRecord(rec *models.Record, rules []models.AggregationRule, state map[string]map[string]float64, counts map[string]map[string]int) {
	for _, rule := range rules {
		// build group by key
		var groupKeys []string
		for _, g := range rule.GroupBy {
			if val, ok := rec.Attributes[g]; ok {
				groupKeys = append(groupKeys, fmt.Sprintf("%v", val))
			} else {
				groupKeys = append(groupKeys, "unknown")
			}
		}
		groupKey := strings.Join(groupKeys, "|")
		if groupKey == "" {
			groupKey = "total"
		}

		if _, ok := state[groupKey]; !ok {
			state[groupKey] = make(map[string]float64)
			counts[groupKey] = make(map[string]int)
		}

		// get the field value if needed
		var val float64
		if rule.Type != "count" {
			if v, ok := rec.Attributes[rule.Field]; ok {
				switch vt := v.(type) {
				case float64:
					val = vt
				case int:
					val = float64(vt)
				case string:
					// already parsed float ideally, skip if not float
					fmt.Sscanf(vt, "%f", &val)
				}
			}
		}

		key := rule.Type + "_" + rule.Field
		switch rule.Type {
		case "count":
			state[groupKey][key]++
		case "sum":
			state[groupKey][key] += val
		case "avg":
			state[groupKey][key] += val
			counts[groupKey][key]++
		case "min":
			if counts[groupKey][key] == 0 || val < state[groupKey][key] {
				state[groupKey][key] = val
			}
			counts[groupKey][key]++
		case "max":
			if counts[groupKey][key] == 0 || val > state[groupKey][key] {
				state[groupKey][key] = val
			}
			counts[groupKey][key]++
		}
	}
}

func sendAggregations(outCh chan<- models.Record, state map[string]map[string]float64, counts map[string]map[string]int, rules []models.AggregationRule) {
	for groupKey, metrics := range state {
		attr := make(map[string]any)
		attr["group"] = groupKey

		for key, val := range metrics {
			if strings.HasPrefix(key, "avg_") {
				c := counts[groupKey][key]
				if c > 0 {
					attr[key] = val / float64(c)
				} else {
					attr[key] = 0
				}
			} else {
				attr[key] = val
			}
		}
		outCh <- models.Record{Attributes: attr}
	}
}
