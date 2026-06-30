package ingestion

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"

	"data-processing-pipeline/internal/models"
)

func RunIngestion(ctx context.Context, sources []models.SourceConfig, outCh chan<- models.Record, errCh chan<- *models.ErrorDetails, jobID string) {
	var wg sync.WaitGroup

	for _, src := range sources {
		wg.Add(1)
		go func(source models.SourceConfig) {
			defer wg.Done()

			switch source.Type {
			case "csv":
				processCSV(ctx, source, outCh, errCh, jobID)
			case "json":
				processJSON(ctx, source, outCh, errCh, jobID)
			case "api":
				processAPI(ctx, source, outCh, errCh, jobID)
			default:
				sendErr(errCh, jobID, "ingestion", fmt.Sprintf("unknown source type: %s", source.Type), "")
			}
		}(src)
	}

	wg.Wait()
	close(outCh) // Ingestion done, close channel
}

func processCSV(ctx context.Context, source models.SourceConfig, outCh chan<- models.Record, errCh chan<- *models.ErrorDetails, jobID string) {
	var reader io.ReadCloser
	var err error

	if source.URL != "" {
		req, _ := http.NewRequestWithContext(ctx, "GET", source.URL, nil)
		resp, reqErr := http.DefaultClient.Do(req)
		if reqErr != nil {
			sendErr(errCh, jobID, "ingestion", fmt.Sprintf("failed to fetch CSV url %s: %v", source.URL, reqErr), "")
			return
		}
		reader = resp.Body
	} else if source.Path != "" {
		reader, err = os.Open(source.Path)
		if err != nil {
			sendErr(errCh, jobID, "ingestion", fmt.Sprintf("failed to open CSV path %s: %v", source.Path, err), "")
			return
		}
	} else {
		sendErr(errCh, jobID, "ingestion", "csv source missing URL or Path", "")
		return
	}
	defer reader.Close()

	csvReader := csv.NewReader(reader)
	// Read header
	headers, err := csvReader.Read()
	if err != nil {
		sendErr(errCh, jobID, "ingestion", fmt.Sprintf("failed to read CSV headers: %v", err), "")
		return
	}

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		row, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			sendErr(errCh, jobID, "ingestion", fmt.Sprintf("csv read error: %v", err), "")
			continue
		}

		attr := make(map[string]any)
		for i, h := range headers {
			if i < len(row) {
				attr[h] = row[i]
			}
		}

		outCh <- models.Record{Attributes: attr}
	}
}

func processJSON(ctx context.Context, source models.SourceConfig, outCh chan<- models.Record, errCh chan<- *models.ErrorDetails, jobID string) {
	var reader io.ReadCloser
	var err error

	if source.URL != "" {
		req, _ := http.NewRequestWithContext(ctx, "GET", source.URL, nil)
		resp, reqErr := http.DefaultClient.Do(req)
		if reqErr != nil {
			sendErr(errCh, jobID, "ingestion", fmt.Sprintf("failed to fetch JSON url %s: %v", source.URL, reqErr), "")
			return
		}
		reader = resp.Body
	} else if source.Path != "" {
		reader, err = os.Open(source.Path)
		if err != nil {
			sendErr(errCh, jobID, "ingestion", fmt.Sprintf("failed to open JSON path %s: %v", source.Path, err), "")
			return
		}
	} else {
		sendErr(errCh, jobID, "ingestion", "json source missing URL or Path", "")
		return
	}
	defer reader.Close()

	decoder := json.NewDecoder(reader)

	// Expect array of objects
	t, err := decoder.Token()
	if err != nil || t != json.Delim('[') {
		sendErr(errCh, jobID, "ingestion", "expected JSON array", "")
		return
	}

	for decoder.More() {
		select {
		case <-ctx.Done():
			return
		default:
		}

		var obj map[string]any
		if err := decoder.Decode(&obj); err != nil {
			sendErr(errCh, jobID, "ingestion", fmt.Sprintf("json decode error: %v", err), "")
			continue
		}
		outCh <- models.Record{Attributes: obj}
	}
}

func processAPI(ctx context.Context, source models.SourceConfig, outCh chan<- models.Record, errCh chan<- *models.ErrorDetails, jobID string) {
	// For API, it's very similar to JSON fetching but could be paged.
	// We'll treat it like JSON URL for this assignment to keep it simple.
	processJSON(ctx, source, outCh, errCh, jobID)
}

func sendErr(errCh chan<- *models.ErrorDetails, jobID, stage, msg, data string) {
	errCh <- &models.ErrorDetails{
		JobID:      jobID,
		Stage:      stage,
		Message:    msg,
		RecordData: data,
		OccurredAt: fmt.Sprintf("%v", context.Background()), // replaced later
	}
}
