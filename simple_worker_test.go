package worker_pool

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestSimpleWorker(t *testing.T) {
	//t.Parallel()
	in_fun := make(chan string)
	out_fun := make(chan string)
	stopSignal := make(chan struct{})
	wgSpeaker := new(sync.WaitGroup)
	wgWp1 := new(sync.WaitGroup)
	sw1 := SimpleWorkerCreator(0, wgWp1, in_fun, out_fun)
	wgWp2 := new(sync.WaitGroup)
	sw2 := SimpleWorkerCreator(0, wgWp2, in_fun, out_fun)
	wgWp3 := new(sync.WaitGroup)
	sw3 := SimpleWorkerCreator(0, wgWp3, in_fun, out_fun)
	var res string

	defer func() {
		close(stopSignal)
		sw1.Stop()
		sw2.Stop()
		sw3.Stop()
		wgSpeaker.Wait()
		wgWp1.Wait()
		wgWp2.Wait()
		wgWp3.Wait()
		close(in_fun)
		close(out_fun)
	}()

	wgWp1.Add(1)
	go sw1.Start()
	wgSpeaker.Add(1)
	go chanSpeaker(in_fun, "Hello world", 1, stopSignal, wgSpeaker)

	t.Logf("(#1) Checking for output values:")
	cnt, cncl := context.WithTimeout(context.Background(), 2*time.Second)
	defer cncl()
	select {
	case res = <-out_fun:
		t.Logf("%c Received: « %s »", Ok, res)
	case <-cnt.Done():
		t.Errorf("%c No response received", Error)
		t.FailNow()
	}

	t.Logf("(#2) Checking the equality of expected and received data:")
	if res == `Message from Worker [id:0]: "Hello world"` {
		t.Logf("%c Equal", Ok)
	} else {
		t.Errorf("%c Not equal", Error)
	}

	t.Logf("(#3) Checking stop when there is no input:")
	sw1.Stop()
	wgWp1.Wait()
	wgSpeaker.Add(1)
	go chanSpeaker(in_fun, "Hello world", 1, stopSignal, wgSpeaker)
	cnt, cncl = context.WithTimeout(context.Background(), 2*time.Second)
	defer cncl()
	select {
	case <-out_fun:
		t.Errorf("%c Not stopped", Error)
	case <-cnt.Done():
		t.Logf("%c Stopped", Ok)
	}

	t.Logf("(#4) Checking for stop when there is input but there in not output:")
	wgWp2.Add(1)
	go sw2.Start()
	wgSpeaker.Add(1)
	go chanSpeaker(in_fun, "Hello world", 10000, stopSignal, wgSpeaker)
	sw2.Stop()
	start := time.Now()
	wgWp2.Wait()
	duration := time.Since(start).Seconds()
	t.Logf("%f", duration)
	cnt, cncl = context.WithTimeout(context.Background(), 2*time.Second)
	defer cncl()
	select {
	case <-out_fun:
		t.Errorf("%c Not stopped", Error)
	case <-cnt.Done():
		t.Logf("%c Stopped", Ok)
	}

	t.Logf("(#5) Checking for stop when there is input and output:")
	wgWp3.Add(1)
	go sw3.Start()
	go func() {
		for i := 0; i < 5; i++ {
			<-out_fun
		}
	}()
	sw3.Stop()
	wgWp3.Wait()
	cnt, cncl = context.WithTimeout(context.Background(), 2*time.Second)
	defer cncl()
	select {
	case <-out_fun:
		t.Errorf("%c Not stopped", Error)
	case <-cnt.Done():
		t.Logf("%c Stopped", Ok)
	}
}
