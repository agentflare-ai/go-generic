package generic

import (
	"fmt"
	"sync"
)

type SyncPool[T any] sync.Pool

func (p *SyncPool[T]) Get() T {
	v, ok := (*sync.Pool)(p).Get().(T)
	if !ok {
		var dv T
		panic(fmt.Errorf("expected %T, got %T", dv, v))
	}
	return v
}

func (p *SyncPool[T]) Put(x T) {
	(*sync.Pool)(p).Put(x)
}
