package controllers

import (
	"encoding/json"
	"log"
	"net/http"

	"data-processing-pipeline/internal/api/services"
	"data-processing-pipeline/internal/models"
)

type JobController struct {
	service services.JobService
}

func NewJobController(service services.JobService) *JobController {
	return &JobController{
		service: service,
	}
}

func (c *JobController) Health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (c *JobController) CreateJob(w http.ResponseWriter, r *http.Request) {
	var spec models.JobSpec
	if err := json.NewDecoder(r.Body).Decode(&spec); err != nil {
		log.Printf("CreateJob: failed to decode JSON payload: %v", err)
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json payload"})
		return
	}

	log.Printf("CreateJob: received request to create pipeline with %d workers", spec.WorkerCount)
	job, err := c.service.CreateJob(r.Context(), spec)
	if err != nil {
		log.Printf("CreateJob: service error: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to save job"})
		return
	}

	log.Printf("CreateJob: successfully created job %s", job.ID)
	writeJSON(w, http.StatusAccepted, job)
}

func (c *JobController) ListJobs(w http.ResponseWriter, r *http.Request) {
	jobs, err := c.service.ListJobs(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list jobs"})
		return
	}
	writeJSON(w, http.StatusOK, jobs)
}

func (c *JobController) GetJob(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	job, err := c.service.GetJob(r.Context(), id)
	if err != nil {
		if err.Error() == "job not found" {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "job not found"})
		} else {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		}
		return
	}
	writeJSON(w, http.StatusOK, job)
}

func (c *JobController) GetJobProgress(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	progress, err := c.service.GetJobProgress(r.Context(), id)
	if err != nil {
		if err.Error() == "progress not found" {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "progress not found"})
		} else {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		}
		return
	}
	writeJSON(w, http.StatusOK, progress)
}

func (c *JobController) GetJobErrors(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	errorsList, err := c.service.GetJobErrors(r.Context(), id)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	writeJSON(w, http.StatusOK, errorsList)
}

func (c *JobController) CancelJob(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	err := c.service.CancelJob(r.Context(), id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "job not running or not found"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "cancel signal sent"})
}

func (c *JobController) DeleteJob(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	err := c.service.DeleteJob(r.Context(), id)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to delete job"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func (c *JobController) GetJobResults(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"message": "Check sqlite or exported CSV/JSON files for results."})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
