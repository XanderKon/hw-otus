package hw05parallelexecution

import (
	"errors"
	"sync"
	"sync/atomic"
)

var ErrErrorsLimitExceeded = errors.New("errors limit exceeded")

type Task func() error

// Run starts tasks in n goroutines and stops its work when receiving m errors from tasks.
func Run(tasks []Task, n, m int) error {
	wg := sync.WaitGroup{}
	tCh := make(chan Task, len(tasks))

	var counter atomic.Int32

	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for t := range tCh {
				if counter.Load() >= int32(m) && m != 0 {
					return
				}
				err := t()
				if err != nil {
					counter.Add(1)
				}
			}
		}()
	}

	for _, t := range tasks {
		tCh <- t
	}

	close(tCh)
	wg.Wait()

	if counter.Load() >= int32(m) && m != 0 {
		return ErrErrorsLimitExceeded
	}

	return nil
}
