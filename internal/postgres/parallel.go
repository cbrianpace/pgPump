package postgres

import (
	"context"
	"sync"

	log "github.com/sirupsen/logrus"
)

// task is a unit of work executed by runParallel.
type task func(ctx context.Context) error

// runParallel executes the given named tasks with at most `threads` running
// concurrently. Results are aggregated on the calling goroutine (single
// consumer of the results channel), so the success/failed tallies are free of
// data races. It returns the number of successful and failed tasks.
func runParallel(ctx context.Context, names []string, threads int, work func(name string) task) (success int, failed int) {
	if threads < 1 {
		threads = 1
	}

	type result struct {
		name string
		err  error
	}

	semaphore := make(chan struct{}, threads)
	results := make(chan result, len(names))
	var wg sync.WaitGroup

	for _, name := range names {
		wg.Add(1)
		semaphore <- struct{}{}

		go func(name string) {
			defer wg.Done()
			defer func() { <-semaphore }()
			results <- result{name: name, err: work(name)(ctx)}
		}(name)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	for r := range results {
		if r.err != nil {
			failed++
			log.Errorf("%s failed: %v", r.name, r.err)
		} else {
			success++
			log.Infof("%s succeeded", r.name)
		}
	}

	return success, failed
}
