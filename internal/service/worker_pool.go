package service

import "context"

type Task struct {
	Name string
	Run  func(context.Context) (interface{}, error)
}

type TaskResult struct {
	Name  string
	Value interface{}
	Err   error
}

type WorkerPool struct {
	workers int
}

func NewWorkerPool(workers int) WorkerPool {
	if workers <= 0 {
		workers = 1
	}

	return WorkerPool{workers: workers}
}

func (p WorkerPool) Run(ctx context.Context, tasks []Task) map[string]TaskResult {
	jobs := make(chan Task)
	results := make(chan TaskResult, len(tasks))

	for i := 0; i < p.workers; i++ {
		go func() {
			for task := range jobs {
				value, err := task.Run(ctx)
				results <- TaskResult{Name: task.Name, Value: value, Err: err}
			}
		}()
	}

	go func() {
		defer close(jobs)
		for _, task := range tasks {
			select {
			case <-ctx.Done():
				return
			case jobs <- task:
			}
		}
	}()

	collected := make(map[string]TaskResult, len(tasks))
	for range tasks {
		select {
		case <-ctx.Done():
			return collected
		case result := <-results:
			collected[result.Name] = result
		}
	}

	return collected
}
