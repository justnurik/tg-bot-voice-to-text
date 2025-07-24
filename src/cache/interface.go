package cache

type Cache[K comparable, V any] interface {
	Add(key K, value V) bool
	Get(key K) (V, bool)
}
