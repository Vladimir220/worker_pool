package worker_pool

// exported
type IWorker interface {
	Start()
	Stop()
}

// exported
type IWorkerPool interface {
	AddAndStartWorkers(count int)
	DropWorkers(count int) error
	Stop()
	GetNumOfWorkers() int
}
