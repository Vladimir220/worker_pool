# Модуль "worker_pool" #
## Автор: Трофимов Владимир ##
---
### Содержание ###
- [Установка](#установка)
- [Введение](#введение)
- [Описание модуля](#описание-модуля)
- [Рекомендации по расширению функционала](#рекомендации-по-расширению-функционала)
- [Инструкция по работе с базовым workerPool](#инструкция-по-работе-с-базовым-workerPool)
- [Пример работы с базовым workerPool](#пример-работы-с-базовым-workerPool)
- [Инструкция по работе с outStreamWorkerPool (proxy-расширением для workerPool)](#инструкция-по-работе-с-outStreamWorkerPool-proxy-расширением-для-workerPool)
- [Пример работы с outStreamWorkerPool (proxy-расширением для workerPool)](#пример-работы-с-outStreamWorkerPool-proxy-расширением-для-workerPool)
---
### Установка ###
```
go get github.com/Vladimir220/worker_pool@v0.0.5
```
### Введение ###
Модуль worker_pool представляет собой инструмент для управления worker-ами (миньонами🙂) с целью параллельной обработки потока данных по заданному алгоритму.

Архитектура модуля предполагает возможность его функционального расширения в соответствии с "Open/Closed Principle" (SOLID). 
Для возможности расширения предлагается использовать интерфейсы, фабричные методы, паттерн проектирования "Посредник" ("Proxy"). Подробнее об этом в разделе "Рекомендации по расширению функционала".
### Описание модуля ###
#### Диаграмма классов для лучшего понимания структуры модуля: ####
![1](https://github.com/Vladimir220/worker_pool/blob/main/pics/class_diagram.jpg)

Последовательно опишу всё интерфейсы и структуры (классы):
- IWorker: **экспортируемый** интерфейс для описания семейства различных взаимозаменяемых worker-ов
- simpleWorker: **неэкспортируемая** структура, представляющая <ins>пример</ins> реализации интерфейса IWorker; эта структура получает данные через входной канал и на выходной канал подаёт сообщение типа `"Message from Worker [id:%d]: \"%s\""`, где s --- полученные данные
- IWorkerPool: **экспортируемый** интерфейс для описания семейства различных взаимозаменяемых worker pool-ов
- workerPool: **неэкспортируемая** структура, представляющая базовую реализацию интерфейса IWorkerPool (хотя остаётся возможность определить другую базовую реализацию); эта структура управляет работой объектов типов, реалезующих интерфейс IWorker
- outStreamWorkerPool: **неэкспортируемая** структура, представляющая <ins>пример</ins> расширения базового функционала worker pool (по умолчанию это структура workerPool) через паттерн проектирования "Посредник" ("Proxy"); эта структура управляет работой объектов типов, реалезующих интерфейс IWorker, а также прослушивает выходной канал и перенаправляет оттуда сообщения в выходной поток io.Writer

Для получения в основной программе экземпляров неэкспортируемых структур используются фабричные методы:
- SimpleWorkerCreator() --- для создания simpleWorker (пример фабрики для примера-структуры)
- WorkerPoolCraetor() --- для создания workerPool
- OutStreamWorkerPoolCraetor() --- для создания OutStreamWorkerPoolCraetor (пример фабрики для примера-структуры)

Обращаю внимание, что workerPool имеет поле workerCreator, которое представляет собой функцию-фабрику для создания worker-ов; это поле должно иметь функциональный тип следующего вида:

`type WorkerCreatorFunc func(int, chan struct{}, *sync.WaitGroup, <-chan string, chan<- string) IWorker`

Связи между структурами:
- Между IWorker и workerPool установлена связь-агрегация, поскольку workerPool с помощью указанного выше функционального поля создаёт массив из элементов IWorker, которые управляются workerPool, однако имеют независимость за счёт их запуска в отдельных горутинах
- Между workerPool и outStreamWorkerPool (пример расширения) установлена связь-композиция, поскольку outStreamWorkerPool управляет строго одним workerPool на протяжении всего своего жизненного цикла
### Рекомендации по расширению функционала ###
Продуманы следующие направления расширений функционала:
- Для добавления новой реализации worker-а создайте структуру, реализующую интерфейс IWorker, и фабричный метод, создающий и инициализирующий эту структуру.
- Для добавления новой реализации базового worker pool (по умолчанию workerPool) создайте структуру, реализующую интерфейс IWorkerPool, и фабричный метод, создающий и инициализирующий эту структуру.
- Для расширения функционала базового worker pool (по умолчанию workerPool) используйте для него proxy-обёртку, которая также надо наследовать от интерфейса IWorkerPool. Для новой структуры также надо сделать фабричный метод, создающий и инициализирующий эту структуру.
### Инструкция по работе с базовым workerPool ###
Создайте workerPool через фабричный метод: 

`WorkerPoolCraetor(wg *sync.WaitGroup, inputCh <-chan string, outputCh chan<- string, workerCreator WorkerCreatorFunc) workerPool`

Сразу решите, как будете останавливать workerPool: самостоятельно или с помощью defer. Важно остановить workerPool, иначе worker-ы продолжат работу после завершения основного кода.

Если вы останавливаете workerPool в другой горутине, то в основном коде можете дождаться окончание работы с помощью: `wg.Wait()`, который был ранее передан в workerPool.

Функции workerPool:
- AddAndStartWorkers(count int) --- добавление count worker-ов и их запуск в горутинах
- DropWorkers(count int): error --- удаление count worker-ов, логика остановки worker-ов описана в реализациях интерфейса IWorker (для simpleWorker их работа будет остановлена либо когда они в режиме ожидания входных данных, либо когда они завершили свою работу с данными)
- Stop() --- остановка всех worker-ов, логика остановки worker-ов как и у DropWorkers
- GetNumOfWorkers() int --- возвращает количество активных worker-ов
### Пример работы с базовым workerPool ###
```
package main

import (
	"fmt"
	"sync"

	wpImp "github.com/Vladimir220/worker_pool"
)

func main() {
	inputCh := make(chan string)
	outputCh := make(chan string)
	wg := &sync.WaitGroup{}
	wp := wpImp.WorkerPoolCraetor(wg, inputCh, outputCh, wpImp.SimpleWorkerCreator)

	defer wp.Stop()

	isStop := false
	stopSignal := make(chan struct{})

	defer func() {
		isStop = true
		close(stopSignal)
	}()

	go func() {

		for {
			select {
			case d := <-outputCh:
				{
					fmt.Println(d)

					if isStop {
						return
					}
				}
			case <-stopSignal:
				return
			}
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
```
#### Результат выполнения: ####
![2](https://github.com/Vladimir220/worker_pool/blob/main/pics/Exmp_res.JPG)

### Инструкция по работе с outStreamWorkerPool (proxy-расширением для workerPool) ###
Для работы с outStreamWorkerPool создайте outStreamWorkerPool через фабричный метод: 

`OutStreamWorkerPoolCraetor(wg *sync.WaitGroup, inputCh <-chan string, workerCreator WorkerCreatorFunc, writer io.Writer) *outStreamWorkerPool`

OutStreamWorkerPool имеет такие же методы, что и workerPool, поскольку они работают по общему интерфейсу IWorkerPool. Однако OutStreamWorkerPool отличается от workerPool способом вывода результатов работы: данные отправляются в поток io.Writer, указанный пользователем в фабричном методе OutStreamWorkerPoolCraetor.

Так же как и с workerPool, не забудьте остановить и OutStreamWorkerPool в конце работы программы.

### Пример работы с outStreamWorkerPool (proxy-расширением для workerPool) ###
```
package main

import (
	"fmt"
	"os"
	"sync"

	wpImp "github.com/Vladimir220/worker_pool"
)

func main() {
	inputCh := make(chan string)
	wg := &sync.WaitGroup{}

	oswp := wpImp.OutStreamWorkerPoolCraetor(wg, inputCh, wpImp.SimpleWorkerCreator, os.Stdout)
	defer oswp.Stop()

	fmt.Println("Тест OutStreamWorkerPoolCraetor")
	fmt.Println("Добавляем и запускаем 3 исполнителя")
	oswp.AddWorkersAndStart(3)
	for i := 0; i < 10; i++ {
		inputCh <- fmt.Sprintf("%d", i)
	}
	fmt.Println("Добавляем и запускаем ещё 2 исполнителя")
	oswp.AddWorkersAndStart(2)
	for i := 0; i < 10; i++ {
		inputCh <- fmt.Sprintf("%d", i)
	}
	fmt.Println("Удаляем 4 исполнителя")
	oswp.DropWorkers(4)

	for i := 0; i < 10; i++ {
		inputCh <- fmt.Sprintf("%d", i)
	}
}
```

Результат работы программы будет такой же, как и в примере работы с workerPool. Однако теперь для считывания результата не нужно считывать выходной канал, результат автоматически будет записан в поток io.Writer, указанный при создании OutStreamWorkerPoolCraetor (в нашем случае --- os.Stdout).
