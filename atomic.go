package generic

import "sync/atomic"

type AtomicValue[T any] atomic.Value

func (a *AtomicValue[T]) Load() T {
	if a == nil {
		var v T
		return v
	}
	return (*atomic.Value)(a).Load().(T)
}

func (a *AtomicValue[T]) Store(x T) {
	if a == nil {
		return
	}
	(*atomic.Value)(a).Store(x)
}

func (a *AtomicValue[T]) Swap(x T) T {
	if a == nil {
		return x
	}
	return (*atomic.Value)(a).Swap(x).(T)
}

func (a *AtomicValue[T]) CompareAndSwap(old, new T) bool {
	if a == nil {
		return false
	}
	return (*atomic.Value)(a).CompareAndSwap(old, new)
}
