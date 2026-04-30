package xstruct

import (
	"sync"
	"sync/atomic"
	"testing"
)

func TestStack_LIFO(t *testing.T) {
	var s Stack[int]
	s.Push(1)
	s.Push(2)
	s.Push(3)

	v, ok := s.Pop()
	if !ok || v != 3 {
		t.Fatalf("pop1: %v ok=%v", v, ok)
	}
	v, ok = s.Pop()
	if !ok || v != 2 {
		t.Fatalf("pop2: %v ok=%v", v, ok)
	}
	v, ok = s.Pop()
	if !ok || v != 1 {
		t.Fatalf("pop3: %v ok=%v", v, ok)
	}
	_, ok = s.Pop()
	if ok {
		t.Fatalf("expected empty")
	}
}

func TestStack_Concurrent(t *testing.T) {
	const (
		goros = 16
		n     = 5000
	)
	var s Stack[int]
	var pushed atomic.Int64

	var wg sync.WaitGroup
	wg.Add(goros)
	for g := 0; g < goros; g++ {
		go func(base int) {
			defer wg.Done()
			for i := 0; i < n; i++ {
				s.Push(base*n + i)
				pushed.Add(1)
			}
		}(g)
	}
	wg.Wait()

	var popped int64
	for {
		_, ok := s.Pop()
		if !ok {
			break
		}
		popped++
	}

	if popped != pushed.Load() {
		t.Fatalf("popped=%d pushed=%d", popped, pushed.Load())
	}
}
