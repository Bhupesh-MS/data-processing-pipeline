package routes

import (
	"net/http"

	"data-processing-pipeline/internal/api/controllers"
)

func NewRouter(controller *controllers.JobController) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", controller.Health)
	
	mux.HandleFunc("POST /api/v1/pipelines", controller.CreateJob)
	mux.HandleFunc("GET /api/v1/pipelines", controller.ListJobs)
	mux.HandleFunc("GET /api/v1/pipelines/{id}", controller.GetJob)
	mux.HandleFunc("GET /api/v1/pipelines/{id}/progress", controller.GetJobProgress)
	mux.HandleFunc("GET /api/v1/pipelines/{id}/results", controller.GetJobResults)
	mux.HandleFunc("GET /api/v1/pipelines/{id}/errors", controller.GetJobErrors)
	mux.HandleFunc("PATCH /api/v1/pipelines/{id}/cancel", controller.CancelJob)
	mux.HandleFunc("DELETE /api/v1/pipelines/{id}", controller.DeleteJob)

	return LoggingMiddleware(mux)
}
