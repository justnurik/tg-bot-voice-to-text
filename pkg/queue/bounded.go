package queue

type BoundedQueue[T any] struct {
	Queue chan T
}

func (b BoundedQueue[T]) Push(val T) {
	b.Queue <- val
}

func (b BoundedQueue[T]) Pop() <-chan T {
	return b.Queue
}
