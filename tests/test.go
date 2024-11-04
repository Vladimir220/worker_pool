package main

import (
	"fmt"
	"sync"

	wp ".."
)

func main() {

	inputCh := make(chan string)
	outputCh := make(chan string)
	wg := &sync.WaitGroup{}
	wp := wp.WorkerPoolCraetor(wg, inputCh, outputCh, wp.SimpleWorkerCreator)

	defer wp.Stop()

	// прослушиваем результаты
	isStop := false
	go func() {

		for d := range outputCh {
			if isStop {
				return
			}
			fmt.Println(d)
		}
	}()

	fmt.Println("Тест базового Worker Pool")
	fmt.Println("Добавляем и запускаем 3 исполнителя")
	wp.AddWorkersAndStart(3)
	for i := 0; i < 10; i++ {
		inputCh <- fmt.Sprintf("%d", i)
	}
	fmt.Println("Добавляем и запускаем ещё 2 исполнителя")
	wp.AddWorkersAndStart(2)
	for i := 0; i < 10; i++ {
		inputCh <- fmt.Sprintf("%d", i)
	}
	fmt.Println("Удаляем 4 исполнителя")
	wp.DropWorkers(4)

	for i := 0; i < 10; i++ {
		inputCh <- fmt.Sprintf("%d", i)
	}

}
