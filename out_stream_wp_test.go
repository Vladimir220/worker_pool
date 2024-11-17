package worker_pool

import (
	"regexp"
	"sync"
	"testing"
	"time"
)

func TestOutStreamWP(t *testing.T) {
	//t.Parallel()
	in_fun := make(chan string)
	stop_signal := make(chan struct{})
	var buff buffer
	wp := OutStreamWorkerPoolCraetor(new(sync.WaitGroup), in_fun, SimpleWorkerCreator, &buff)
	var res string

	defer func() {
		close(stop_signal)
		close(in_fun)
	}()

	t.Logf("(#1) Checking the addition of workers and receiving output from them:")
	t.Logf("Adding three workers...")
	go wp.AddWorkersAndStart(3)
	time.Sleep(1 * time.Second)
	workers_ids := Set{}
	workers_ids.Add("0")
	workers_ids.Add("1")
	workers_ids.Add("2")
	go chanSpeaker(in_fun, "Hello world", 3, stop_signal)
	time.Sleep(2 * time.Second)
	received_ids := Set{}
	reg := regexp.MustCompile(`id:(\d+)`)
	time.Sleep(2 * time.Second)

	if len(buff) != 3 {
		t.Errorf("%c No response received", c_error)
		t.FailNow()
	}
	for i := 0; i < 3; i++ {
		res = buff[i]
		t.Logf("%c Received: « %s »", c_ok, res)
		match := reg.FindStringSubmatch(res)
		if match == nil {
			t.Errorf("%c Unexpected output", c_error)
			t.FailNow()
		}
		received_ids.Add(match[1])
	}
	if received_ids.Equals(workers_ids) {
		t.Logf("%c All workers have worked (3/3)", c_ok)
	} else {
		t.Errorf("%c Not all workers worked", c_error)
		t.FailNow()
	}
	buff = buff[:0]

	t.Logf("(#2) Checking the receipt of the number of active workers:")
	num := wp.GetNumOfWorkers()
	if num == 3 {
		t.Logf("%c %d", c_ok, num)
	} else {
		t.Errorf("%c %d", c_error, num)
		t.FailNow()
	}

	t.Logf("(#3) Checking worker removal:")
	t.Logf("Removing two workers...")
	delete(workers_ids, "1")
	delete(workers_ids, "2")
	wp.DropWorkers(2)
	go chanSpeaker(in_fun, "Hello world", 3, stop_signal)
	time.Sleep(2 * time.Second)
	received_ids = Set{}
	if len(buff) != 3 {
		t.Errorf("%c No response received", c_error)
		t.FailNow()
	}
	for i := 0; i < 3; i++ {
		res = buff[i]
		t.Logf("%c Received: « %s »", c_ok, res)
		match := reg.FindStringSubmatch(res)
		if match == nil {
			t.Errorf("%c Unexpected output", c_error)
			t.FailNow()
		}
		received_ids.Add(match[1])
	}
	if received_ids.Equals(workers_ids) {
		t.Logf("%c Only one worker worked", c_ok)
	} else {
		t.Errorf("%c Not only one worker worked", c_error)
		t.FailNow()
	}
	buff = buff[:0]

	t.Logf("(#4) Checking for removal of incorrect number of workers:")
	err := wp.DropWorkers(10)
	if err != nil {
		t.Logf("%c Error detected", c_ok)
	} else {
		t.Errorf("%c No error detected", c_error)
	}

	t.Logf("(#5) Checking the pool stop:")
	wp.Stop()
	go chanSpeaker(in_fun, "Hello world", 1, stop_signal)
	time.Sleep(2 * time.Second)
	if len(buff) != 1 {
		t.Logf("%c Stopped", c_ok)
	} else {
		t.Errorf("%c Not stopped", c_error)
	}
}
