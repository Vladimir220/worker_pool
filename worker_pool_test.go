package worker_pool

import (
	"context"
	"regexp"
	"sync"
	"testing"
	"time"
)

func TestWorkerPool(t *testing.T) {
	//t.Parallel()
	inFun := make(chan string)
	outFun := make(chan string)
	stopSignal := make(chan struct{})
	wgWp := new(sync.WaitGroup)
	wgSpeaker := new(sync.WaitGroup)

	wp := WorkerPoolCraetor(wgWp, inFun, outFun, SimpleWorkerCreator)
	var res string

	defer func() {
		close(stopSignal)
		wp.Stop()
		wgSpeaker.Wait()
		wgWp.Wait()
		close(inFun)
		close(outFun)
	}()

	t.Logf("(#1) Checking the addition of workers and receiving output from them:")
	t.Logf("Adding three workers...")
	go wp.AddWorkersAndStart(3)
	workersIds := Set{}
	workersIds.Add("0")
	workersIds.Add("1")
	workersIds.Add("2")
	wgSpeaker.Add(1)
	go chanSpeaker(inFun, "Hello world", 3, stopSignal, wgSpeaker)
	time.Sleep(2 * time.Second)
	receivedIds := Set{}
	reg := regexp.MustCompile(`id:(\d+)`)
	cnt, cncl := context.WithTimeout(context.Background(), 10*time.Second)
	defer cncl()
	for i := 0; i < 3; i++ {
		select {
		case res = <-outFun:
			t.Logf("%c Received: « %s »", Ok, res)
			match := reg.FindStringSubmatch(res)
			if match == nil {
				t.Errorf("%c Unexpected output", Error)
				t.FailNow()
			}
			receivedIds.Add(match[1])
		case <-cnt.Done():
			t.Errorf("%c No response received", Error)
			t.FailNow()
		}
	}
	if receivedIds.Equals(workersIds) {
		t.Logf("%c All workers have worked (3/3)", Ok)
	} else {
		t.Errorf("%c Not all workers worked", Error)
		t.FailNow()
	}

	t.Logf("(#2) Checking the receipt of the number of active workers:")
	num := wp.GetNumOfWorkers()
	if num == 3 {
		t.Logf("%c %d", Ok, num)
	} else {
		t.Errorf("%c %d", Error, num)
		t.FailNow()
	}

	t.Logf("(#3) Checking worker removal:")
	t.Logf("Removing two workers...")
	delete(workersIds, "1")
	delete(workersIds, "2")
	wp.DropWorkers(2)
	wgSpeaker.Add(1)
	go chanSpeaker(inFun, "Hello world", 3, stopSignal, wgSpeaker)
	time.Sleep(2 * time.Second)
	receivedIds = Set{}
	cnt, cncl = context.WithTimeout(context.Background(), 10*time.Second)
	defer cncl()
	for i := 0; i < 3; i++ {
		select {
		case res = <-outFun:
			t.Logf("%c Received: « %s »", Ok, res)
			match := reg.FindStringSubmatch(res)
			if match == nil {
				t.Errorf("%c Unexpected output", Error)
				t.FailNow()
			}
			receivedIds.Add(match[1])
		case <-cnt.Done():
			t.Errorf("%c No response received", Error)
			t.FailNow()
		}
	}
	if receivedIds.Equals(workersIds) {
		t.Logf("%c Only one worker worked", Ok)
	} else {
		t.Errorf("%c Not only one worker worked", Error)
		t.FailNow()
	}

	t.Logf("(#4) Checking for removal of incorrect number of workers:")
	err := wp.DropWorkers(10)
	if err != nil {
		t.Logf("%c Error detected", Ok)
	} else {
		t.Errorf("%c No error detected", Error)
	}

	t.Logf("(#5) Checking the pool stop:")
	wp.Stop()
	wgSpeaker.Add(1)
	go chanSpeaker(inFun, "Hello world", 1, stopSignal, wgSpeaker)
	cnt, cncl = context.WithTimeout(context.Background(), 3*time.Second)
	defer cncl()
	select {
	case <-outFun:
		t.Errorf("%c Not stopped", Error)
	case <-cnt.Done():
		t.Logf("%c Stopped", Ok)
	}

}
