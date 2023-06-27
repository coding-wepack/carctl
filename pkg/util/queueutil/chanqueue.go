package queueutil

import (
	"sync"
	"sync/atomic"
)

func Producer[T any](files []T, out chan<- T) {
	defer close(out)
	for _, file := range files {
		out <- file
	}
}

func Consumer[T any](in <-chan T, ec chan<- error, wg *sync.WaitGroup, count *int32, fn func(file T) error) {
	defer wg.Done()
	for {
		select {
		case file, ok := <-in:
			if !ok {
				return
			}
			atomic.AddInt32(count, 1)
			err := fn(file)
			if err != nil {
				ec <- err
			}
		}
	}
}
