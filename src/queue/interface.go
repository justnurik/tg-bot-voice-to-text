package queue

type Queue[T any] interface {
	Push(T)
	Pop() <-chan T
}
