package queue

import "sync/atomic"

type UnboundedChanQueue[T any] struct {
	mu    chan struct{}
	queue []T

	isWake atomic.Bool
	wake   chan struct{}
}

func NewUnboundedChanQueue[T any]() *UnboundedChanQueue[T] {
	return &UnboundedChanQueue[T]{
		mu:    make(chan struct{}, 1),
		queue: make([]T, 0),

		wake: make(chan struct{}),
	}
}

func (q *UnboundedChanQueue[T]) Push(item T) {
	q.mu <- struct{}{}
	defer func() { <-q.mu }()

	q.queue = append(q.queue, item)

	if q.isWake.CompareAndSwap(false, true) {
		close(q.wake) // wake
	}
}

func (q *UnboundedChanQueue[T]) Pop() <-chan T {
	return chanWait(func() T {
		q.mu <- struct{}{}
		for len(q.queue) == 0 {
			<-q.mu // unlock mutex

			<-q.wake // sleep

			q.mu <- struct{}{} // lock mutex

			if q.isWake.CompareAndSwap(true, false) {
				q.wake = make(chan struct{})
			}
		}
		defer func() { <-q.mu }()

		begin := q.queue[0]
		q.queue = q.queue[1:]

		return begin
	})
}

func chanWait[T any](fn func() T) <-chan T {
	result := make(chan T, 1)

	go func() {
		result <- fn()
	}()

	return result
}
