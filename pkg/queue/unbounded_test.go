package queue

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestQueue_Sequential(t *testing.T) {
	q := NewUnboundedChanQueue[int]()

	q.Push(1)
	q.Push(2)
	q.Push(3)

	assert.Equal(t, 1, <-q.Pop())
	assert.Equal(t, 2, <-q.Pop())
	assert.Equal(t, 3, <-q.Pop())
}

func TestQueue_ConcurrentPush(t *testing.T) {
	q := NewUnboundedChanQueue[int]()
	const count = 1000

	var wg sync.WaitGroup
	wg.Add(count)

	for i := range count {
		go func(i int) {
			defer wg.Done()
			q.Push(i)
		}(i)
	}

	wg.Wait()

	received := make(map[int]bool, count)
	for range count {
		v := <-q.Pop()
		require.False(t, received[v], "duplicate value %d", v)
		received[v] = true
	}
}

func TestQueue_ConcurrentPop(t *testing.T) {
	q := NewUnboundedChanQueue[string]()
	const count = 1000

	for i := range count {
		q.Push("item-" + string(rune(i)))
	}

	var wg sync.WaitGroup
	wg.Add(count)

	received := make(chan string, count)
	for range count {
		go func() {
			defer wg.Done()
			received <- <-q.Pop()
		}()
	}

	wg.Wait()
	close(received)

	items := make(map[string]bool, count)
	for item := range received {
		require.False(t, items[item], "duplicate item %s", item)
		items[item] = true
	}
	assert.Equal(t, count, len(items))
}

func TestQueue_MixedOperations(t *testing.T) {
	q := NewUnboundedChanQueue[int]()
	const count = 10_000

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		for i := range count {
			q.Push(i)
			time.Sleep(time.Microsecond)
		}
	}()

	received := make(chan int, count)
	go func() {
		defer wg.Done()
		for range count {
			received <- <-q.Pop()
		}
	}()

	wg.Wait()
	close(received)

	expected := 0
	for v := range received {
		assert.Equal(t, expected, v)
		expected++
	}
}
