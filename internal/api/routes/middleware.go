package routes

import (
	"log"
	"net/http"
	"time"
)

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("-> Received %s request for %s", r.Method, r.URL.Path)
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("<- %s %s completed in %s", r.Method, r.URL.Path, time.Since(start))
	})
}
