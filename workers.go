package worker_pool

import (
	"fmt"
	"sync"
)

// exported

type IWorker interface {
	Start()
	Stop()
}

// unexported
// implement IWorker
// example
type simpleWorker struct {
	id         int
	stopSignal chan struct{}
	act_mu     *sync.RWMutex
	isActive   bool
	wg         *sync.WaitGroup
	inputCh    <-chan string
	outputCh   chan<- string
}

func (sw *simpleWorker) Start() {
	defer sw.wg.Done()

	for {
		select {
		case str := <-sw.inputCh:
			{
				buf := fmt.Sprintf("Message from Worker [id:%d]: \"%s\"", sw.id, str)
				sw.outputCh <- buf
				sw.act_mu.RLock()
				if !sw.isActive {
					sw.act_mu.RUnlock()
					return
				}
				sw.act_mu.RUnlock()
			}
		case <-sw.stopSignal:
			return

		}
	}

}

func (sw *simpleWorker) Stop() {
	close(sw.stopSignal)
	sw.act_mu.Lock()
	sw.isActive = false
	sw.act_mu.Unlock()
}

// exported
// factory for any worker
type WorkerCreatorFunc func(int, chan struct{}, *sync.RWMutex, *sync.WaitGroup, <-chan string, chan<- string) IWorker

// exported
// implement WorkerCreatorFunc
// factory for simpleWorker
func SimpleWorkerCreator(id int, stopSignal chan struct{}, act_mu *sync.RWMutex, wg *sync.WaitGroup, inputCh <-chan string, outputCh chan<- string) IWorker {
	return &simpleWorker{id, stopSignal, act_mu, true, wg, inputCh, outputCh}
}
