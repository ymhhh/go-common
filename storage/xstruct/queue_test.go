package xstruct

import (
	"sync"
	"sync/atomic"
	"testing"
)

func TestQueue_FIFO(t *testing.T) {
	q := NewQueue[int]()
	q.Enqueue(1)
	q.Enqueue(2)
	q.Enqueue(3)

	v, ok := q.Dequeue()
	if !ok || v != 1 {
		t.Fatalf("dq1: %v ok=%v", v, ok)
	}
	v, ok = q.Dequeue()
	if !ok || v != 2 {
		t.Fatalf("dq2: %v ok=%v", v, ok)
	}
	v, ok = q.Dequeue()
	if !ok || v != 3 {
		t.Fatalf("dq3: %v ok=%v", v, ok)
	}
	_, ok = q.Dequeue()
	if ok {
		t.Fatalf("expected empty")
	}
}

func TestQueue_Concurrent(t *testing.T) {
	const (
		producers = 8
		perProd   = 1000
	)
	q := NewQueue[int]()

	var pushed atomic.Int64

	var wg sync.WaitGroup
	wg.Add(producers)
	for p := 0; p < producers; p++ {
		go func(base int) {
			defer wg.Done()
			for i := 0; i < perProd; i++ {
				q.Enqueue(base*perProd + i)
				pushed.Add(1)
			}
		}(p)
	}
	wg.Wait()

	var popped int64
	for {
		_, ok := q.Dequeue()
		if !ok {
			break
		}
		popped++
	}

	if popped != pushed.Load() {
		t.Fatalf("popped=%d pushed=%d", popped, pushed.Load())
	}
}
