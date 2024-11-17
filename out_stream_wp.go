package worker_pool

import (
	"io"
	"strings"
	"sync"
)

// ------ division of proxy extension for WorkerPool (example) ------

// unexported
// implement IWorkerPool
// example
type outStreamWorkerPool struct {
	wp            workerPool
	wg            *sync.WaitGroup
	mu            *sync.RWMutex
	inputCh       <-chan string
	outputCh      chan string
	workerCreator WorkerCreatorFunc
	writer        io.Writer
	isListening   bool
	stopSignal    chan struct{}
}

func (p *outStreamWorkerPool) GetNumOfWorkers() int {
	return p.wp.GetNumOfWorkers()
}

func (p *outStreamWorkerPool) AddWorkersAndStart(count int) {
	p.wp.AddWorkersAndStart(count)
}

func (p *outStreamWorkerPool) DropWorkers(count int) error {
	return p.wp.DropWorkers(count)
}

func (p *outStreamWorkerPool) Stop() {
	p.wp.Stop()
	p.stopListening()
}

func (p *outStreamWorkerPool) listeningCh() {

	go func() {
		for {
			select {
			case str := <-p.outputCh:
				{
					io.Copy(p.writer, strings.NewReader(str+"\n"))
					p.mu.RLock()
					if !p.isListening {
						p.mu.RUnlock()
						return
					}
					p.mu.RUnlock()
				}
			case <-p.stopSignal:
				return
			}
		}
	}()
}

func (p *outStreamWorkerPool) stopListening() {
	p.mu.Lock()
	p.isListening = false
	p.mu.Unlock()
	close(p.stopSignal)
}

// exported
// factory for OutStreamWorkerPoolCraetor
func OutStreamWorkerPoolCraetor(wg *sync.WaitGroup, inputCh <-chan string, workerCreator WorkerCreatorFunc, writer io.Writer) *outStreamWorkerPool {
	outputCh := make(chan string)
	wp := WorkerPoolCraetor(wg, inputCh, outputCh, workerCreator)
	oswp := outStreamWorkerPool{wp, &sync.WaitGroup{}, new(sync.RWMutex), inputCh, outputCh, workerCreator, writer, true, make(chan struct{})}
	oswp.listeningCh()
	return &oswp
}