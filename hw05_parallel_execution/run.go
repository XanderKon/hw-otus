package hw05parallelexecution

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

var ErrErrorsLimitExceeded = errors.New("errors limit exceeded")

type Task func() error

// Run starts tasks in n goroutines and stops its work when receiving m errors from tasks.
func Run(tasks []Task, n, m int) error {
	wg := sync.WaitGroup{}

	tCh := make(chan Task, len(tasks))

	var counter = atomic.Int32{}

	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for t := range tCh {
				if counter.Load() >= int32(m) {
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

	if counter.Load() >= int32(m) {
		return ErrErrorsLimitExceeded
	}

	return nil
}

func main() {
	tasksCount := 50
	tasks := make([]Task, 0, tasksCount)

	var runTasksCount int32
	var sumTime time.Duration

	for i := 0; i < tasksCount; i++ {
		taskSleep := time.Second / 2
		sumTime += taskSleep

		tasks = append(tasks, func() error {
			time.Sleep(taskSleep)
			atomic.AddInt32(&runTasksCount, 1)

			if runTasksCount == 10 || runTasksCount == 11 {
				return fmt.Errorf("error from task %d", i)
			}

			return nil
		})
	}

	workersCount := 5
	maxErrorsCount := 1

	start := time.Now()
	err := Run(tasks, workersCount, maxErrorsCount)
	elapsedTime := time.Since(start)

	fmt.Println(err)
	fmt.Println(elapsedTime)
	fmt.Println("runTasksCount:", runTasksCount)
}
