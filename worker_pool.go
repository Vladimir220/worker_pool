package worker_pool

import (
	"fmt"
	"io"
	"strings"
	"sync"
	"time"
)

// ------ division of workers ------

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
				if !sw.isActive {
					return
				}
			}
		case <-sw.stopSignal:
			return

		}
	}

}

func (sw *simpleWorker) Stop() {
	close(sw.stopSignal)
	sw.isActive = false
}

// exported
// factory for any worker
type WorkerCreatorFunc func(int, chan struct{}, *sync.WaitGroup, <-chan string, chan<- string) IWorker

// exported
// implement WorkerCreatorFunc
// factory for simpleWorker
func SimpleWorkerCreator(id int, stopSignal chan struct{}, wg *sync.WaitGroup, inputCh <-chan string, outputCh chan<- string) IWorker {
	return &simpleWorker{id, stopSignal, true, wg, inputCh, outputCh}
}

// ------ division of worker pools ------

// exported
type IWorkerPool interface {
	AddAndStartWorkers(count int)
	DropWorkers(count int) error
	Stop()
	GetNumOfWorkers() int
}

// unexported
// implement IWorkerPool
type workerPool struct {
	wg            *sync.WaitGroup
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
		wp.wg.Add(1)
		worker := wp.workerCreator(i, make(chan struct{}), wp.wg, wp.inputCh, wp.outputCh)
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
	return workerPool{wg, inputCh, outputCh, workerCreator, make([]IWorker, 0)}
}

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
	isListening   bool
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
	p.StopListening()
}

func (p *outStreamWorkerPool) ListeningCh() {

	go func() {
		for str := range p.outputCh {
			if !p.isListening {
				return
			}
			io.Copy(p.writer, strings.NewReader(str+"\n"))
		}
	}()
}

func (p *outStreamWorkerPool) StopListening() {
	p.isListening = false
}

// exported
// factory for OutStreamWorkerPoolCraetor
func OutStreamWorkerPoolCraetor(wg *sync.WaitGroup, inputCh <-chan string, workerCreator WorkerCreatorFunc, writer io.Writer) *outStreamWorkerPool {
	outputCh := make(chan string)
	wp := WorkerPoolCraetor(wg, inputCh, outputCh, workerCreator)
	oswp := outStreamWorkerPool{wp, &sync.WaitGroup{}, inputCh, outputCh, workerCreator, writer, true}
	oswp.ListeningCh()
	return &oswp
}
