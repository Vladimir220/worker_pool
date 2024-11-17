package worker_pool

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestSimpleWorker(t *testing.T) {
	in_fun := make(chan string)
	out_fun := make(chan string)
	stop_signal := make(chan struct{})
	sw := SimpleWorkerCreator(0, new(sync.WaitGroup), in_fun, out_fun)
	var res string

	defer func() {
		close(stop_signal)
		time.Sleep(500 * time.Millisecond)
		close(in_fun)
		close(out_fun)
	}()

	go sw.Start()
	go chanSpeaker(in_fun, "Hello world", 1, stop_signal)

	t.Logf("(#1) Checking for output values:")
	cnt, cncl := context.WithTimeout(context.Background(), 2*time.Second)
	defer cncl()
	select {
	case res = <-out_fun:
		t.Logf("%c Received: « %s »", c_ok, res)
	case <-cnt.Done():
		t.Errorf("%c No response received", c_error)
		t.FailNow()
	}

	t.Logf("(#2) Checking the equality of expected and received data:")
	if res == `Message from Worker [id:0]: "Hello world"` {
		t.Logf("%c Equal", c_ok)
	} else {
		t.Errorf("%c Not equal", c_error)
	}

	t.Logf("(#3) Checking stop when there is no input:")
	sw.Stop()
	go chanSpeaker(in_fun, "Hello world", 1, stop_signal)
	cnt, cncl = context.WithTimeout(context.Background(), 2*time.Second)
	defer cncl()
	select {
	case <-out_fun:
		t.Errorf("%c Not stopped", c_error)
	case <-cnt.Done():
		t.Logf("%c Stopped", c_ok)
	}

	t.Logf("(#4) Checking for stop when there is input:")
	go sw.Start()
	go chanSpeaker(in_fun, "Hello world", 10000, stop_signal)
	sw.Stop()
	cnt, cncl = context.WithTimeout(context.Background(), 2*time.Second)
	defer cncl()
	select {
	case <-out_fun:
		t.Errorf("%c Not stopped", c_error)
	case <-cnt.Done():
		t.Logf("%c Stopped", c_ok)
	}
}
