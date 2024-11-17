package worker_pool

import (
	"context"
	"regexp"
	"sync"
	"testing"
	"time"
)

func TestWorkerPool(t *testing.T) {
	in_fun := make(chan string)
	out_fun := make(chan string)
	stop_signal := make(chan struct{})
	wp := WorkerPoolCraetor(new(sync.WaitGroup), in_fun, out_fun, SimpleWorkerCreator)
	var res string

	defer func() {
		close(stop_signal)
		close(in_fun)
		close(out_fun)
	}()

	t.Logf("#1 Checking the addition of workers and receiving output from them:")
	t.Logf("Adding three workers...")
	go wp.AddWorkersAndStart(3)
	workers_ids := Set{}
	workers_ids.Add("0")
	workers_ids.Add("1")
	workers_ids.Add("2")
	go chanSpeaker(in_fun, "Hello world", 3, stop_signal)
	time.Sleep(2 * time.Second)
	received_ids := Set{}
	reg := regexp.MustCompile(`id:(\d+)`)
	cnt, cncl := context.WithTimeout(context.Background(), 10*time.Second)
	defer cncl()
	defer wp.Stop()
	for i := 0; i < 3; i++ {
		select {
		case res = <-out_fun:
			t.Logf("%c Received: « %s »", c_ok, res)
			match := reg.FindStringSubmatch(res)
			if match == nil {
				t.Errorf("%c Unexpected output", c_error)
				t.FailNow()
			}
			received_ids.Add(match[1])
		case <-cnt.Done():
			t.Errorf("%c No response received", c_error)
			t.FailNow()
		}
	}
	if received_ids.Equals(workers_ids) {
		t.Logf("%c All workers have worked (3/3)", c_ok)
	} else {
		t.Errorf("%c Not all workers worked", c_error)
		t.FailNow()
	}

	t.Logf("#2 Checking the receipt of the number of active workers:")
	num := wp.GetNumOfWorkers()
	if num == 3 {
		t.Logf("%c %d", c_ok, num)
	} else {
		t.Errorf("%c %d", c_error, num)
		t.FailNow()
	}

	t.Logf("#3 Checking worker removal:")
	t.Logf("Removing two workers...")
	delete(workers_ids, "1")
	delete(workers_ids, "2")
	wp.DropWorkers(2)
	go chanSpeaker(in_fun, "Hello world", 3, stop_signal)
	time.Sleep(2 * time.Second)
	received_ids = Set{}
	cnt, cncl = context.WithTimeout(context.Background(), 10*time.Second)
	defer cncl()
	for i := 0; i < 3; i++ {
		select {
		case res = <-out_fun:
			t.Logf("%c Received: « %s »", c_ok, res)
			match := reg.FindStringSubmatch(res)
			if match == nil {
				t.Errorf("%c Unexpected output", c_error)
				t.FailNow()
			}
			received_ids.Add(match[1])
		case <-cnt.Done():
			t.Errorf("%c No response received", c_error)
			t.FailNow()
		}
	}
	if received_ids.Equals(workers_ids) {
		t.Logf("%c Only one worker worked", c_ok)
	} else {
		t.Errorf("%c Not only one worker worked", c_error)
		t.FailNow()
	}

	t.Logf("#4 Checking the pool stop:")
	wp.Stop()
	go chanSpeaker(in_fun, "Hello world", 1, stop_signal)
	cnt, cncl = context.WithTimeout(context.Background(), 3*time.Second)
	defer cncl()
	select {
	case <-out_fun:
		t.Errorf("%c Not stopped", c_error)
	case <-cnt.Done():
		t.Logf("%c Stopped", c_ok)
	}
}
