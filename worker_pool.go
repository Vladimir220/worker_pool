package worker_pool

import (
	"fmt"
	"sync"
	"time"
)

// unexported
// implement IWorkerPool
type workerPool struct {
	wg            *sync.WaitGroup
	mu            *sync.RWMutex
	inputCh       <-chan string
	outputCh      chan<- string
	workerCreator WorkerCreatorFunc
	workers       []IWorker
}

func (wp *workerPool) GetNumOfWorkers() int {
	return len(wp.workers)
}

func (wp *workerPool) AddWorkersAndStart(count int) {
	sizeW := len(wp.workers)
	for i := sizeW; i < sizeW+count; i++ {
		worker := wp.workerCreator(i, make(chan struct{}), wp.mu, wp.wg, wp.inputCh, wp.outputCh)
		wp.workers = append(wp.workers, worker)
		go worker.Start()
	}
}

func (wp *workerPool) DropWorkers(count int) error {
	if count > len(wp.workers) {
		return fmt.Errorf("you are trying to drop (%d) more workers than there are in the pool (%d)", count, len(wp.workers))
	}
	time.Sleep(time.Second * 1)
	for i := 1; i <= count; i++ {
		wp.workers[len(wp.workers)-i].Stop()
	}
	wp.workers = wp.workers[:len(wp.workers)-count]
	return nil
}

func (wp *workerPool) Stop() {
	time.Sleep(time.Second * 1)
	for _, w := range wp.workers {
		w.Stop()
	}
	wp.workers = make([]IWorker, 0)
}

// exported
// factory for WorkerPool
func WorkerPoolCraetor(wg *sync.WaitGroup, inputCh <-chan string, outputCh chan<- string, workerCreator WorkerCreatorFunc) workerPool {
	return workerPool{wg, new(sync.RWMutex), inputCh, outputCh, workerCreator, make([]IWorker, 0)}
}
