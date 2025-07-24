package scheduler

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type SimpleQueue[T any] struct {
	items chan T
}

func NewMockQueue[T any]() *SimpleQueue[T] {
	return &SimpleQueue[T]{
		items: make(chan T, 100),
	}
}

func (m *SimpleQueue[T]) Push(item T) {
	m.items <- item
}

func (m *SimpleQueue[T]) Pop() <-chan T {
	return m.items
}

func TestNewNamedWorkerSchedulerQueue(t *testing.T) {
	ctx := context.Background()
	queue := NewMockQueue[func(string)]()
	scheduler := NewNamedWorkerSchedulerQueue(ctx, queue)

	assert.NotNil(t, scheduler)
	assert.Equal(t, queue, scheduler.taskQueue)
	assert.Equal(t, ctx, scheduler.ctx)
	assert.Nil(t, scheduler.stop)
}

func TestStart(t *testing.T) {
	ctx := context.Background()
	queue := NewMockQueue[func(string)]()
	scheduler := NewNamedWorkerSchedulerQueue(ctx, queue)

	workersID := []string{"worker1", "worker2"}
	scheduler.Start(workersID)

	assert.NotNil(t, scheduler.stop)

	scheduler.Stop()
}

func TestSchedule(t *testing.T) {
	ctx := context.Background()
	queue := NewMockQueue[func(string)]()
	scheduler := NewNamedWorkerSchedulerQueue(ctx, queue)

	workersID := []string{"worker1"}
	scheduler.Start(workersID)

	done := scheduler.Schedule(func(string) {
		time.Sleep(time.Second)
	})

	time.Sleep(time.Millisecond * 10)

	select {
	case <-done:
		t.Fatalf("")
	case <-time.After(time.Second / 4):
	}

	select {
	case <-done:
	case <-time.After(time.Second * 2):
		t.Fatalf("")
	}

	scheduler.Stop()
}

func TestSchedul(t *testing.T) {
	ctx := context.Background()
	queue := NewMockQueue[func(string)]()
	scheduler := NewNamedWorkerSchedulerQueue(ctx, queue)

	workersID := []string{"worker1", "worker2", "worker3", "worker4"}
	scheduler.Start(workersID)

	var mu sync.Mutex
	executedWorkers := make(map[string]int)
	doneChans := make([]chan struct{}, 0, 10)

	for range 10000 {
		done := scheduler.Schedule(func(workerID string) {
			mu.Lock()
			executedWorkers[workerID]++
			mu.Unlock()
		})
		doneChans = append(doneChans, done)
	}

	for _, done := range doneChans {
		select {
		case <-done:
		case <-time.After(time.Second):
			t.Fatal("Task was not executed within 1 second")
		}
	}

	mu.Lock()
	assert.Equal(t, 10000, executedWorkers["worker1"]+executedWorkers["worker2"]+executedWorkers["worker3"]+executedWorkers["worker4"], "All tasks should be executed")
	mu.Unlock()

	scheduler.Stop()
}

func TestScheduleMultipleWorkers(t *testing.T) {
	ctx := context.Background()
	queue := NewMockQueue[func(string)]()
	scheduler := NewNamedWorkerSchedulerQueue(ctx, queue)

	workersID := []string{"worker1", "worker2"}
	scheduler.Start(workersID)

	var mu sync.Mutex
	executedWorkers := make(map[string]int)
	doneChans := make([]chan struct{}, 0, 10)

	for range 2 {
		done := scheduler.Schedule(func(workerID string) {
			mu.Lock()
			executedWorkers[workerID]++
			mu.Unlock()
			time.Sleep(time.Second / 4)
		})
		doneChans = append(doneChans, done)
	}

	for _, done := range doneChans {
		select {
		case <-done:
		case <-time.After(time.Second):
			t.Fatal("Task was not executed within 1 second")
		}
	}

	mu.Lock()
	assert.Equal(t, 1, executedWorkers["worker1"])
	assert.Equal(t, 1, executedWorkers["worker2"])
	assert.Equal(t, 2, executedWorkers["worker1"]+executedWorkers["worker2"], "All tasks should be executed")
	mu.Unlock()

	scheduler.Stop()
}

func TestStop(t *testing.T) {
	ctx := context.Background()
	queue := NewMockQueue[func(string)]()
	scheduler := NewNamedWorkerSchedulerQueue(ctx, queue)

	workersID := []string{"worker1"}
	scheduler.Start(workersID)
	scheduler.Stop()

	done := scheduler.Schedule(func(workerID string) {
		t.Fatal("Task should not be executed after Stop")
	})

	select {
	case _, open := <-scheduler.stop:
		assert.False(t, open, "stop channel should be closed")
	default:
		t.Fatal("stop channel should be closed after Stop")
	}

	doneCh := make(chan struct{})
	go func() {
		scheduler.wg.Wait()
		close(doneCh)
	}()
	select {
	case <-doneCh:
	case <-time.After(time.Second):
		t.Fatal("WaitGroup did not complete within 1 second")
	}

	select {
	case <-done:
		t.Fatal("Task should not be executed after Stop")
	default:
	}
}

func TestContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	queue := NewMockQueue[func(string)]()
	scheduler := NewNamedWorkerSchedulerQueue(ctx, queue)

	workersID := []string{"worker1"}
	scheduler.Start(workersID)

	cancel()

	done := scheduler.Schedule(func(workerID string) {
		t.Fatal("Task should not be executed after context cancellation")
	})

	doneCh := make(chan struct{})
	go func() {
		scheduler.wg.Wait()
		close(doneCh)
	}()
	select {
	case <-doneCh:
	case <-time.After(time.Second):
		t.Fatal("WaitGroup did not complete within 1 second")
	}

	select {
	case <-done:
		t.Fatal("Task should not be executed after context cancellation")
	default:
	}
}

func TestGenericType(t *testing.T) {
	ctx := context.Background()
	queue := NewMockQueue[func(int)]()
	scheduler := NewNamedWorkerSchedulerQueue(ctx, queue)

	workersID := []int{1, 2}
	scheduler.Start(workersID)

	var receivedWorkerID int
	done := scheduler.Schedule(func(workerID int) {
		receivedWorkerID = workerID
	})

	select {
	case <-done:
		assert.Contains(t, workersID, receivedWorkerID)
	case <-time.After(time.Second):
		t.Fatal("Task was not executed within 1 second")
	}

	scheduler.Stop()
}
