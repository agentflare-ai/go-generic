package generic

import (
	"fmt"
	"sync"
)

type SyncPool[T any] sync.Pool

func (p *SyncPool[T]) Get() T {
	item := (*sync.Pool)(p).Get()
	if item == nil {
		var zero T
		return zero
	}
	v, ok := item.(T)
	if !ok {
		var dv T
		panic(fmt.Errorf("expected %T, got %T", dv, item))
	}
	return v
}

func (p *SyncPool[T]) Put(x T) {
	(*sync.Pool)(p).Put(x)
}
