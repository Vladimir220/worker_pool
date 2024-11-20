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
	inputCh       <-chan string
	outputCh      chan string
	workerCreator WorkerCreatorFunc
	writer        io.Writer
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
	select {
	case <-p.stopSignal:
		return
	default:
	}
	p.wp.Stop()
	close(p.stopSignal)
}

func (p *outStreamWorkerPool) listeningCh() {

	go func() {
		for {
			select {
			case str := <-p.outputCh:
				{
					select {
					case <-p.stopSignal:
						return
					default:
					}

					io.Copy(p.writer, strings.NewReader(str+"\n")) // вот тут нужен мутекс
				}
			case <-p.stopSignal:
				return
			}
		}
	}()
}

// exported
// factory for OutStreamWorkerPoolCraetor
func OutStreamWorkerPoolCraetor(wg *sync.WaitGroup, inputCh <-chan string, workerCreator WorkerCreatorFunc, writer io.Writer) *outStreamWorkerPool {
	outputCh := make(chan string)
	wp := WorkerPoolCraetor(wg, inputCh, outputCh, workerCreator)
	oswp := outStreamWorkerPool{wp, &sync.WaitGroup{}, inputCh, outputCh, workerCreator, writer, make(chan struct{})}
	oswp.listeningCh()
	return &oswp
}
