package pipeline

import "sync"

type WorkerPool struct {
	workers int
}

func NewWorkerPool(workers int) *WorkerPool {
	if workers <= 0 {
		workers = 1
	}
	return &WorkerPool{workers: workers}
}

func (p *WorkerPool) Run(tasks []func()) {
	jobs := make(chan func())
	var wg sync.WaitGroup

	for range p.workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for task := range jobs {
				task()
			}
		}()
	}

	for _, task := range tasks {
		jobs <- task
	}
	close(jobs)
	wg.Wait()
}
