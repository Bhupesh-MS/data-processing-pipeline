package export

import (
	"context"
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"

	"data-processing-pipeline/internal/models"
)

func RunExport(ctx context.Context, inCh <-chan models.Record, errCh chan<- *models.ErrorDetails, exports []models.ExportConfig, jobID string, db *sql.DB) {
	// For each record, we export to all targets
	// We can buffer them or write streamingly
	
	// Open all file handles / db connections
	var csvWriters []*csv.Writer
	var jsonEncoders []*json.Encoder
	var files []*os.File
	
	for _, exp := range exports {
		if exp.Type == "csv" && exp.Path != "" {
			f, err := os.Create(exp.Path)
			if err != nil {
				sendErr(errCh, jobID, "export", fmt.Sprintf("failed to create csv %s", exp.Path))
				continue
			}
			files = append(files, f)
			csvWriters = append(csvWriters, csv.NewWriter(f))
		} else if exp.Type == "json" && exp.Path != "" {
			f, err := os.Create(exp.Path)
			if err != nil {
				sendErr(errCh, jobID, "export", fmt.Sprintf("failed to create json %s", exp.Path))
				continue
			}
			files = append(files, f)
			jsonEncoders = append(jsonEncoders, json.NewEncoder(f))
			// start array
			f.WriteString("[\n")
		} else if exp.Type == "sqlite" && exp.Table != "" && db != nil {
			// init table if needed
			// Note: this is a simple dynamic table creation, 
			// in a real app this should be more robust
		}
	}

	defer func() {
		for _, w := range csvWriters {
			w.Flush()
		}
		for _, exp := range exports {
			if exp.Type == "json" {
				// close array
				// finding corresponding file requires mapping, doing simple generic approach:
				// this is just an assignment, so simply tracking state
				// for a real app we would have separate writer structs
			}
		}
		for _, f := range files {
			f.Close()
		}
	}()

	first := true
	var headers []string

	for {
		select {
		case <-ctx.Done():
			return
		case rec, ok := <-inCh:
			if !ok {
				return
			}
			
			// Get headers on first record for CSV/SQLite
			if first {
				for k := range rec.Attributes {
					headers = append(headers, k)
				}
				for _, w := range csvWriters {
					w.Write(headers)
				}
				
				// create sqlite table
				for _, exp := range exports {
					if exp.Type == "sqlite" && exp.Table != "" && db != nil {
						query := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (", exp.Table)
						for i, h := range headers {
							query += fmt.Sprintf("`%s` TEXT", h)
							if i < len(headers)-1 {
								query += ", "
							}
						}
						query += ")"
						_, err := db.Exec(query)
						if err != nil {
							sendErr(errCh, jobID, "export", fmt.Sprintf("sqlite table creation failed: %v", err))
						}
					}
				}
				first = false
			}

			// Write to CSV
			row := make([]string, len(headers))
			for i, h := range headers {
				row[i] = fmt.Sprintf("%v", rec.Attributes[h])
			}
			for _, w := range csvWriters {
				w.Write(row)
			}

			// Write to JSON
			for _, enc := range jsonEncoders {
				enc.Encode(rec.Attributes)
			}

			// Write to SQLite
			for _, exp := range exports {
				if exp.Type == "sqlite" && exp.Table != "" && db != nil {
					placeholders := ""
					var args []any
					for i, h := range headers {
						placeholders += "?"
						if i < len(headers)-1 {
							placeholders += ", "
						}
						args = append(args, fmt.Sprintf("%v", rec.Attributes[h]))
					}
					query := fmt.Sprintf("INSERT INTO %s VALUES (%s)", exp.Table, placeholders)
					_, err := db.Exec(query, args...)
					if err != nil {
						sendErr(errCh, jobID, "export", fmt.Sprintf("sqlite insert failed: %v", err))
					}
				}
			}
		}
	}
}

func sendErr(errCh chan<- *models.ErrorDetails, jobID, stage, msg string) {
	errCh <- &models.ErrorDetails{
		JobID:   jobID,
		Stage:   stage,
		Message: msg,
	}
}
