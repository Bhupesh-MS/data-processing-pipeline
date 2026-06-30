package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"data-processing-pipeline/internal/api/controllers"
	"data-processing-pipeline/internal/api/routes"
	"data-processing-pipeline/internal/api/services"
	"data-processing-pipeline/internal/models"
	"data-processing-pipeline/internal/pipeline"
	"data-processing-pipeline/internal/storage"
)

func TestFullPipeline(t *testing.T) {
	// 1. Setup SQLite Store
	dbPath := "test_pipeline.db"
	os.Remove(dbPath)
	defer os.Remove(dbPath)

	store, err := storage.NewSQLiteStore(dbPath)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer store.Close()

	// 2. Setup Orchestrator & Router
	orchestrator := pipeline.NewOrchestrator(store)
	jobService := services.NewJobService(store, orchestrator)
	jobController := controllers.NewJobController(jobService)
	router := routes.NewRouter(jobController)

	// 3. Mock API Server for JSON data
	mockData := `[{"id": 1, "name": "alice", "score": "80"}, {"id": 2, "name": "bob", "score": "95"}]`
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(mockData))
	}))
	defer mockServer.Close()

	// 4. Create Job Spec
	jobSpec := models.JobSpec{
		Sources: []models.SourceConfig{
			{Type: "json", URL: mockServer.URL},
		},
		Validations: []models.ValidationRule{
			{Field: "id", Type: "int", Required: true},
			{Field: "name", Type: "string", Required: true},
		},
		Transformations: []models.TransformRule{
			{Field: "name", Action: "uppercase"},
			{Field: "score", Action: "to_int"},
		},
		Aggregations: []models.AggregationRule{
			{Type: "sum", Field: "score", GroupBy: []string{}},
			{Type: "count", Field: "score", GroupBy: []string{}},
		},
		Exports: []models.ExportConfig{
			{Type: "sqlite", Table: "results_table"},
		},
		WorkerCount: 2,
	}

	// 5. Post Job
	jobBytes, _ := json.Marshal(jobSpec)
	req := httptest.NewRequest("POST", "/api/v1/pipelines", strings.NewReader(string(jobBytes)))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Result().StatusCode != http.StatusAccepted {
		t.Fatalf("expected 202, got %d", w.Result().StatusCode)
	}

	var createdJob models.Job
	json.NewDecoder(w.Body).Decode(&createdJob)
	
	// 6. Wait for job to finish
	time.Sleep(3 * time.Second) // wait a bit for background goroutines

	// 7. Check Progress
	reqProgress := httptest.NewRequest("GET", "/api/v1/pipelines/"+createdJob.ID+"/progress", nil)
	wProgress := httptest.NewRecorder()
	router.ServeHTTP(wProgress, reqProgress)

	var prog models.Progress
	json.NewDecoder(wProgress.Body).Decode(&prog)

	if prog.Status != models.JobStatusCompleted {
		t.Errorf("expected job to be completed, got %v", prog.Status)
	}

	// 8. Check SQLite for exported data
	var count int
	var sumScore float64
	var countScore float64
	row := store.DB().QueryRow("SELECT COUNT(*), sum_score, count_score FROM results_table")
	if err := row.Scan(&count, &sumScore, &countScore); err != nil {
		t.Logf("failed to query results: %v", err)
	}

	rows, _ := store.DB().Query("SELECT stage, message FROM errors")
	for rows != nil && rows.Next() {
		var stage, msg string
		rows.Scan(&stage, &msg)
		t.Logf("ERROR from DB: %s %s", stage, msg)
	}

	if count != 1 {
		t.Errorf("expected 1 result row, got %d", count)
	}
	if sumScore != 175 {
		t.Errorf("expected sum_score 175, got %f", sumScore)
	}
	if countScore != 2 {
		t.Errorf("expected count_score 2, got %f", countScore)
	}
}

