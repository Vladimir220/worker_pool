package worker_pool

import (
	"fmt"
	"sync"
)

// unexported
// implement IWorker
// example
type simpleWorker struct {
	id         int
	stopSignal chan struct{}
	isActive   bool
	once       sync.Once
	muAct      sync.RWMutex
	wg         *sync.WaitGroup
	inputCh    <-chan string
	outputCh   chan<- string
}

func (sw *simpleWorker) Start() {
	sw.once.Do(sw.startOnce)
}

func (sw *simpleWorker) startOnce() {

	sw.muAct.RLock()
	if !sw.isActive {
		sw.muAct.RUnlock()
		return
	}
	sw.muAct.RUnlock()

	sw.wg.Add(1)
	defer sw.wg.Done()

	for {
		select {
		case str := <-sw.inputCh:
			select {
			case <-sw.stopSignal:
				return
			default:
			}

			buf := fmt.Sprintf("Message from Worker [id:%d]: \"%s\"", sw.id, str)

			select {
			case sw.outputCh <- buf:
			case <-sw.stopSignal:
				return
			}
		case <-sw.stopSignal:
			return
		}
	}

}

func (sw *simpleWorker) Stop() {
	sw.muAct.Lock()
	if !sw.isActive {
		sw.muAct.Unlock()
		return
	}
	sw.isActive = false
	sw.muAct.Unlock()
	close(sw.stopSignal)
}

// exported
// factory for any worker
type WorkerCreatorFunc func(int, *sync.WaitGroup, <-chan string, chan<- string) IWorker

// exported
// implement WorkerCreatorFunc
// factory for simpleWorker
func SimpleWorkerCreator(id int, wg *sync.WaitGroup, inputCh <-chan string, outputCh chan<- string) IWorker {
	return &simpleWorker{id, make(chan struct{}), true, sync.Once{}, sync.RWMutex{}, wg, inputCh, outputCh}
}
