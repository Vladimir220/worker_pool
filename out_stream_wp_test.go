package worker_pool

import (
	"regexp"
	"sync"
	"testing"
	"time"
)

func TestOutStreamWP(t *testing.T) {
	//t.Parallel()
	inFun := make(chan string)
	stopSignal := make(chan struct{})
	wgWp := new(sync.WaitGroup)
	wgSpeaker := new(sync.WaitGroup)
	var (
		buff buffer
		res  string
		err  error
	)
	wp := OutStreamWorkerPoolCraetor(wgWp, inFun, SimpleWorkerCreator, &buff)

	defer func() {
		close(stopSignal)
		wp.Stop()
		wgSpeaker.Wait()
		wgWp.Wait()
		close(inFun)
	}()

	t.Logf("(#1) Checking the addition of workers and receiving output from them:")
	t.Logf("Adding three workers...")
	go wp.AddWorkersAndStart(3)
	time.Sleep(1 * time.Second)
	workersIds := Set{}
	workersIds.Add("0")
	workersIds.Add("1")
	workersIds.Add("2")
	wgSpeaker.Add(1)
	go chanSpeaker(inFun, "Hello world", 3, stopSignal, wgSpeaker)
	time.Sleep(2 * time.Second)
	received_ids := Set{}
	reg := regexp.MustCompile(`id:(\d+)`)
	time.Sleep(2 * time.Second)

	if buff.len() != 3 {
		t.Errorf("%c No response received", Error)
		t.FailNow()
	}
	for i := 0; i < 3; i++ {
		res, err = buff.get(i)
		if err != nil {
			t.Errorf("%c %s", Error, err.Error())
			t.FailNow()
		}
		t.Logf("%c Received: « %s »", Ok, res)
		match := reg.FindStringSubmatch(res)
		if match == nil {
			t.Errorf("%c Unexpected output", Error)
			t.FailNow()
		}
		received_ids.Add(match[1])
	}
	if received_ids.Equals(workersIds) {
		t.Logf("%c All workers have worked (3/3)", Ok)
	} else {
		t.Errorf("%c Not all workers worked", Error)
		t.FailNow()
	}
	buff.Clear()

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
	received_ids = Set{}
	if buff.len() != 3 { // неожиданно это тоже операция чтения
		t.Errorf("%c No response received", Error)
		t.FailNow()
	}
	for i := 0; i < 3; i++ {
		res, err = buff.get(i)
		if err != nil {
			t.Errorf("%c %s", Error, err.Error())
			t.FailNow()
		}
		t.Logf("%c Received: « %s »", Ok, res)
		match := reg.FindStringSubmatch(res)
		if match == nil {
			t.Errorf("%c Unexpected output", Error)
			t.FailNow()
		}
		received_ids.Add(match[1])
	}
	if received_ids.Equals(workersIds) {
		t.Logf("%c Only one worker worked", Ok)
	} else {
		t.Errorf("%c Not only one worker worked", Error)
		t.FailNow()
	}
	buff.Clear()

	t.Logf("(#4) Checking for removal of incorrect number of workers:")
	err = wp.DropWorkers(10)
	if err != nil {
		t.Logf("%c Error detected", Ok)
	} else {
		t.Errorf("%c No error detected", Error)
	}

	t.Logf("(#5) Checking the pool stop:")
	wp.Stop()
	wgSpeaker.Add(1)
	go chanSpeaker(inFun, "Hello world", 1, stopSignal, wgSpeaker)
	time.Sleep(2 * time.Second)
	if buff.len() != 1 {
		t.Logf("%c Stopped", Ok)
	} else {
		t.Errorf("%c Not stopped", Error)
	}
}
