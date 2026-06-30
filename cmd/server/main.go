package main

import (
	"log"
	"net/http"
	"os"

	"data-processing-pipeline/internal/api/controllers"
	"data-processing-pipeline/internal/api/routes"
	"data-processing-pipeline/internal/api/services"
	"data-processing-pipeline/internal/pipeline"
	"data-processing-pipeline/internal/storage"
)

func main() {
	addr := os.Getenv("SERVER_ADDR")
	if addr == "" {
		addr = ":8080"
	}

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "pipeline.db"
	}

	store, err := storage.NewSQLiteStore(dbPath)
	if err != nil {
		log.Fatalf("failed to initialize db: %v", err)
	}
	defer store.Close()

	orchestrator := pipeline.NewOrchestrator(store)
	jobService := services.NewJobService(store, orchestrator)
	jobController := controllers.NewJobController(jobService)
	router := routes.NewRouter(jobController)

	log.Printf("starting server on %s", addr)
	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
