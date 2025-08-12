package scheduler

type NamedWorkerScheduler[K any] interface {
	Start(workersID []K)
	Schedule(func(workerID K)) chan struct{}
	Stop()
}
