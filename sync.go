package generic

import (
	"fmt"
	"sync"
)

type SyncPool[T any] sync.Pool

func (p *SyncPool[T]) Get() T {
	v, ok := p.New().(T)
	if !ok {
		panic(fmt.Errorf("expected %T, got %T", v, v))
	}
	return v
}

func (p *SyncPool[T]) Put(x T) {
	p.Put(x)
}
