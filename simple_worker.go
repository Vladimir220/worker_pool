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
	once       sync.Once
	mu         sync.RWMutex
	ready      sync.WaitGroup
	wg         *sync.WaitGroup
	inputCh    <-chan string
	outputCh   chan<- string
}

func (sw *simpleWorker) Start() {
	sw.once.Do(sw.startOnce)
}

func (sw *simpleWorker) startOnce() {
	select {
	case <-sw.stopSignal:
		return
	default:
	}

	defer sw.wg.Done()
	sw.ready.Done()

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
	select {
	case <-sw.stopSignal:
		return
	default:
	}

	sw.ready.Wait()
	close(sw.stopSignal)
}

// exported
// factory for any worker
type WorkerCreatorFunc func(int, *sync.WaitGroup, <-chan string, chan<- string) IWorker

// exported
// implement WorkerCreatorFunc
// factory for simpleWorker
func SimpleWorkerCreator(id int, wg *sync.WaitGroup, inputCh <-chan string, outputCh chan<- string) IWorker {
	sw := &simpleWorker{id, make(chan struct{}), sync.Once{}, sync.RWMutex{}, sync.WaitGroup{}, wg, inputCh, outputCh}
	sw.ready.Add(1)
	return sw
}
