package xstruct

import "sync/atomic"

type node[T any] struct {
	next  atomic.Pointer[node[T]]
	value T
}
