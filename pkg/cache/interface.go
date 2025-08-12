package cache

type Cache[K comparable, V any] interface {
	Add(key K, value V) bool
	Get(key K) (V, bool)
}

type EmptyCache[K comparable, V any] struct {
}

func (EmptyCache[K, V]) Add(K, V) bool {
	return false
}

func (EmptyCache[K, V]) Get(key K) (a V, b bool) {
	b = false
	return
}
