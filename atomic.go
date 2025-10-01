package generic

import "sync/atomic"

type AtomicValue[T any] atomic.Value

func (a *AtomicValue[T]) Load() T {
	return (*atomic.Value)(a).Load().(T)
}

func (a *AtomicValue[T]) Store(x T) {
	(*atomic.Value)(a).Store(x)
}

func (a *AtomicValue[T]) Swap(x T) T {
	return (*atomic.Value)(a).Swap(x).(T)
}

func (a *AtomicValue[T]) CompareAndSwap(old, new T) bool {
	return (*atomic.Value)(a).CompareAndSwap(old, new)
}
