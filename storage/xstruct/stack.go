package xstruct

import (
	"sync/atomic"
)

// Stack is a lock-free LIFO stack (Treiber stack) built on atomic.Pointer.
type Stack[T any] struct {
	top atomic.Pointer[node[T]]
}

// Push pushes v on top of the stack.
func (s *Stack[T]) Push(v T) {
	n := &node[T]{value: v}
	for {
		old := s.top.Load()
		n.next.Store(old)
		if s.top.CompareAndSwap(old, n) {
			return
		}
	}
}

// Pop removes and returns the top element.
// If the stack is empty, ok is false.
func (s *Stack[T]) Pop() (v T, ok bool) {
	for {
		old := s.top.Load()
		if old == nil {
			var zero T
			return zero, false
		}
		nxt := old.next.Load()
		if s.top.CompareAndSwap(old, nxt) {
			return old.value, true
		}
	}
}
