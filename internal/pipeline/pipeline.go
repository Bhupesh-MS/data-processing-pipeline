package pipeline

import (
	"context"
	"sync"
	"time"

	"data-processing-pipeline/internal/aggregation"
	"data-processing-pipeline/internal/api/repositories"
	"data-processing-pipeline/internal/export"
	"data-processing-pipeline/internal/ingestion"
	"data-processing-pipeline/internal/models"
	"data-processing-pipeline/internal/transformation"
	"data-processing-pipeline/internal/validation"
)

type Orchestrator struct {
	repo       repositories.JobRepository
	activeJobs map[string]context.CancelFunc
	mu         sync.Mutex
}

func NewOrchestrator(repo repositories.JobRepository) *Orchestrator {
	return &Orchestrator{
		repo:       repo,
		activeJobs: make(map[string]context.CancelFunc),
	}
}

func (o *Orchestrator) RunJob(job *models.Job) {
	ctx, cancel := context.WithCancel(context.Background())

	o.mu.Lock()
	o.activeJobs[job.ID] = cancel
	o.mu.Unlock()

	defer func() {
		o.mu.Lock()
		delete(o.activeJobs, job.ID)
		o.mu.Unlock()
	}()

	err := o.repo.UpdateJobStatus(context.Background(), job.ID, models.JobStatusRunning, "")
	if err != nil {
		return
	}

	// This is where we run the pipeline
	p := NewPipeline(job, o.repo)
	err = p.Run(ctx)

	status := models.JobStatusCompleted
	errMsg := ""
	if err != nil {
		if ctx.Err() == context.Canceled {
			status = models.JobStatusCancelled
		} else {
			status = models.JobStatusFailed
			errMsg = err.Error()
		}
	}

	o.repo.UpdateJobStatus(context.Background(), job.ID, status, errMsg)
}

func (o *Orchestrator) CancelJob(jobID string) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	if cancel, ok := o.activeJobs[jobID]; ok {
		cancel()
		return nil
	}
	return context.DeadlineExceeded // or some "not found" error
}

type Pipeline struct {
	job  *models.Job
	repo repositories.JobRepository
}

func NewPipeline(job *models.Job, repo repositories.JobRepository) *Pipeline {
	return &Pipeline{
		job:  job,
		repo: repo,
	}
}

func (p *Pipeline) Run(ctx context.Context) error {
	// Channels
	recordsCh := make(chan models.Record, 100)
	validatedCh := make(chan models.Record, 100)
	transformedCh := make(chan models.Record, 100)
	resultCh := make(chan models.Record, 100)
	errCh := make(chan *models.ErrorDetails, 100)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var wg sync.WaitGroup

	// Error collector
	wg.Add(1)
	go func() {
		defer wg.Done()
		for errDetail := range errCh {
			if errDetail.OccurredAt == "" {
				errDetail.OccurredAt = time.Now().Format(time.RFC3339)
			}
			p.repo.SaveError(context.Background(), errDetail)
			// update progress failed count
			prog, _ := p.repo.GetProgress(context.Background(), p.job.ID)
			if prog != nil {
				prog.Metrics.RecordsFailed++
				p.repo.UpdateProgress(context.Background(), prog)
			}
		}
	}()

	// Progress tracker
	wg.Add(1)
	go func() {
		defer wg.Done()
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				prog, _ := p.repo.GetProgress(context.Background(), p.job.ID)
				if prog != nil {
					elapsed := time.Since(prog.StartTime).Seconds()
					if elapsed > 0 {
						prog.Metrics.ProcessingRate = float64(prog.Metrics.RecordsProcessed) / elapsed
					}
					p.repo.UpdateProgress(context.Background(), prog)
				}
			}
		}
	}()

	// Stages
	go func() {
		ingestion.RunIngestion(ctx, p.job.Spec.Sources, recordsCh, errCh, p.job.ID)
	}()

	// Validation workers
	workerCount := p.job.Spec.WorkerCount
	if workerCount <= 0 {
		workerCount = 5
	}

	var valWg sync.WaitGroup
	for i := 0; i < workerCount; i++ {
		valWg.Add(1)
		go func() {
			defer valWg.Done()
			validation.RunValidation(ctx, recordsCh, validatedCh, errCh, p.job.Spec.Validations, p.job.ID)
		}()
	}
	go func() {
		valWg.Wait()
		close(validatedCh)
	}()

	// Transformation workers
	var transWg sync.WaitGroup
	for i := 0; i < workerCount; i++ {
		transWg.Add(1)
		go func() {
			defer transWg.Done()
			transformation.RunTransformation(ctx, validatedCh, transformedCh, errCh, p.job.Spec.Transformations, p.job.ID)
		}()
	}
	go func() {
		transWg.Wait()
		close(transformedCh)
	}()

	// Aggregation
	go func() {
		if len(p.job.Spec.Aggregations) > 0 {
			aggregation.RunAggregation(ctx, transformedCh, resultCh, p.job.Spec.Aggregations)
		} else {
			// Passthrough if no aggregations
			for rec := range transformedCh {
				resultCh <- rec
			}
		}
		close(resultCh)
	}()

	// Export (consumer of resultCh)
	// Track processed count inline or in export
	exportDone := make(chan struct{})
	go func() {
		// we intercept resultCh to increment progress
		exportCh := make(chan models.Record, 100)
		go func() {
			for rec := range resultCh {
				prog, _ := p.repo.GetProgress(context.Background(), p.job.ID)
				if prog != nil {
					prog.Metrics.RecordsProcessed++
					p.repo.UpdateProgress(context.Background(), prog)
				}
				exportCh <- rec
			}
			close(exportCh)
		}()

		// Export
		export.RunExport(ctx, exportCh, errCh, p.job.Spec.Exports, p.job.ID, p.repo.DB())
		close(exportDone)
	}()

	<-exportDone

	close(errCh)
	cancel() // cancel context so progress tracker stops
	wg.Wait()

	// final update
	prog, _ := p.repo.GetProgress(context.Background(), p.job.ID)
	if prog != nil {
		now := time.Now()
		prog.EndTime = &now
		prog.Status = models.JobStatusCompleted
		p.repo.UpdateProgress(context.Background(), prog)
	}

	return nil
}
