package xstruct

import (
	"sync"
	"sync/atomic"
)

// Queue is a lock-free FIFO queue (Michael & Scott) built on atomic.Pointer.
//
// It keeps a dummy head node so enqueue/dequeue can update head/tail without
// special empty-queue CAS races.
type Queue[T any] struct {
	head  atomic.Pointer[node[T]]
	tail  atomic.Pointer[node[T]]
	once  sync.Once
}

func NewQueue[T any]() *Queue[T] {
	dummy := &node[T]{}
	q := &Queue[T]{}
	q.head.Store(dummy)
	q.tail.Store(dummy)
	q.once.Do(func() {}) // mark as initialized
	return q
}

func (q *Queue[T]) init() {
	q.once.Do(func() {
		dummy := &node[T]{}
		q.head.Store(dummy)
		q.tail.Store(dummy)
	})
}

// Enqueue adds v to the tail of the queue.
func (q *Queue[T]) Enqueue(v T) {
	q.init()
	n := &node[T]{value: v}
	for {
		tail := q.tail.Load()
		next := tail.next.Load()
		if next == nil {
			if tail.next.CompareAndSwap(nil, n) {
				q.tail.CompareAndSwap(tail, n)
				return
			}
		} else {
			q.tail.CompareAndSwap(tail, next)
		}
	}
}

// Dequeue removes and returns the front element.
// If the queue is empty, ok is false.
func (q *Queue[T]) Dequeue() (v T, ok bool) {
	q.init()
	for {
		head := q.head.Load()
		tail := q.tail.Load()
		next := head.next.Load()

		// Empty (dummy has no successor).
		if next == nil {
			var zero T
			return zero, false
		}

		// Tail is behind; help advance it.
		if head == tail {
			q.tail.CompareAndSwap(tail, next)
			continue
		}

		val := next.value
		if q.head.CompareAndSwap(head, next) {
			return val, true
		}
	}
}
