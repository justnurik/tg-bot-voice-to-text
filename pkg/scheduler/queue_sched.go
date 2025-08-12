package scheduler

import (
	"context"
	"sync"

	"tg-bot-voice-to-text/pkg/queue"
)

type NamedWorkerSchedulerQueue[K any] struct {
	taskQueue queue.Queue[func(K)]

	wg   sync.WaitGroup
	ctx  context.Context
	stop chan struct{}
}

func NewNamedWorkerSchedulerQueue[K any](ctx context.Context, queue queue.Queue[func(K)]) *NamedWorkerSchedulerQueue[K] {
	return &NamedWorkerSchedulerQueue[K]{
		taskQueue: queue,
		ctx:       ctx,
	}
}

func (n *NamedWorkerSchedulerQueue[K]) Start(workersID []K) {
	n.stop = make(chan struct{})

	n.wg.Add(len(workersID))
	for _, workerID := range workersID {
		go n.worker(workerID)
	}
}

func (n *NamedWorkerSchedulerQueue[K]) Schedule(task func(workerID K)) chan struct{} {
	done := make(chan struct{})

	n.taskQueue.Push(func(workerID K) {
		defer close(done)
		task(workerID)
	})

	return done
}

func (n *NamedWorkerSchedulerQueue[K]) Stop() {
	close(n.stop)
	n.wg.Wait()
}

func (n *NamedWorkerSchedulerQueue[K]) worker(workerID K) {
	defer n.wg.Done()

	for {
		select {
		case <-n.ctx.Done():
			return
		case <-n.stop:
			return
		case task := <-n.taskQueue.Pop():
			task(workerID)
		}
	}
}
