package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"data-processing-pipeline/internal/api/controllers"
	"data-processing-pipeline/internal/api/routes"
	"data-processing-pipeline/internal/api/services"
	"data-processing-pipeline/internal/models"
	"data-processing-pipeline/internal/pipeline"
	"data-processing-pipeline/internal/storage"
)

func TestAPIIntegration(t *testing.T) {
	// Setup real SQLite in-memory database
	store, err := storage.NewSQLiteStore("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer store.Close()

	orchestrator := pipeline.NewOrchestrator(store)
	jobService := services.NewJobService(store, orchestrator)
	jobController := controllers.NewJobController(jobService)
	router := routes.NewRouter(jobController)

	// Create a test server
	ts := httptest.NewServer(router)
	defer ts.Close()

	// 1. Create a Job
	spec := models.JobSpec{WorkerCount: 1}
	body, _ := json.Marshal(spec)
	resp, err := http.Post(ts.URL+"/api/v1/pipelines", "application/json", bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("Post failed: %v", err)
	}
	if resp.StatusCode != http.StatusAccepted {
		t.Errorf("expected 202, got %v", resp.StatusCode)
	}

	var job models.Job
	if err := json.NewDecoder(resp.Body).Decode(&job); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	resp.Body.Close()

	if job.ID == "" {
		t.Fatalf("expected job ID, got empty")
	}

	// 2. Get the Job
	resp, err = http.Get(ts.URL + "/api/v1/pipelines/" + job.ID)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %v", resp.StatusCode)
	}
	resp.Body.Close()

	// 3. List Jobs
	resp, err = http.Get(ts.URL + "/api/v1/pipelines")
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %v", resp.StatusCode)
	}
	var jobs []models.Job
	json.NewDecoder(resp.Body).Decode(&jobs)
	resp.Body.Close()
	if len(jobs) != 1 {
		t.Errorf("expected 1 job in list, got %d", len(jobs))
	}

	// 4. Get Progress
	resp, err = http.Get(ts.URL + "/api/v1/pipelines/" + job.ID + "/progress")
	if err != nil {
		t.Fatalf("Get Progress failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 for progress, got %v", resp.StatusCode)
	}
	resp.Body.Close()
}
