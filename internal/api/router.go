package api

import "net/http"

func NewRouter() http.Handler {
	handler := NewHandler()
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", handler.Health)
	mux.HandleFunc("POST /jobs", handler.CreateJob)

	return LoggingMiddleware(mux)
}
