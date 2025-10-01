package generic

import (
	"fmt"
	"sync/atomic"
)

type Atomic[T any] struct {
	load           func() T
	store          func(x T)
	swap           func(x T) T
	compareAndSwap func(old, new T) bool
}

func (a Atomic[T]) Load() T {
	if a.load == nil {
		var v T
		return v
	}
	return a.load()
}

func (a Atomic[T]) Store(x T) {
	if a.store == nil {
		return
	}
	a.store(x)
}

func (a Atomic[T]) Swap(x T) T {
	if a.swap == nil {
		return x
	}
	return a.swap(x)
}

func (a Atomic[T]) CompareAndSwap(old, new T) bool {
	if a.compareAndSwap == nil {
		return false
	}
	return a.compareAndSwap(old, new)
}

func MakeAtomic[T any](maybeDefaultValue ...T) Atomic[T] {
	var a atomic.Value
	if len(maybeDefaultValue) > 0 {
		a.Store(maybeDefaultValue[0])
	}
	return Atomic[T]{
		load: func() T {
			v, ok := a.Load().(T)
			if !ok {
				var dv T
				panic(fmt.Errorf("expected %T, got %T", dv, v))
			}
			return v
		},
		store: func(x T) { a.Store(x) },
		swap: func(x T) T {
			v, ok := a.Swap(x).(T)
			if !ok {
				var dv T
				panic(fmt.Errorf("expected %T, got %T", dv, v))
			}
			return v
		},
		compareAndSwap: func(old, new T) bool { return a.CompareAndSwap(old, new) },
	}
}
