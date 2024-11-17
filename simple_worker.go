package worker_pool

import (
	"fmt"
	"sync"
	"time"
)

// unexported
// implement IWorker
// example
type simpleWorker struct {
	id         int
	stopSignal chan struct{}
	isActive   bool
	wg         *sync.WaitGroup
	inputCh    <-chan string
	outputCh   chan<- string
}

func (sw *simpleWorker) Start() {
	sw.wg.Add(1)
	defer sw.wg.Done()

	select {
	case _, ok := <-sw.stopSignal:
		if !ok {
			sw.stopSignal = make(chan struct{})
		}
	default:
	}

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
	time.Sleep(700 * time.Millisecond)
	close(sw.stopSignal)
	time.Sleep(700 * time.Millisecond)
}

// exported
// factory for any worker
type WorkerCreatorFunc func(int, *sync.WaitGroup, <-chan string, chan<- string) IWorker

// exported
// implement WorkerCreatorFunc
// factory for simpleWorker
func SimpleWorkerCreator(id int, wg *sync.WaitGroup, inputCh <-chan string, outputCh chan<- string) IWorker {
	return &simpleWorker{id, make(chan struct{}), true, wg, inputCh, outputCh}
}
