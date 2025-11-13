package generic

import (
	"context"
	"errors"
	"sync"
)

var ErrEmptyQueue = errors.New("queue is empty")

type FiFo[T any] struct {
	mu    sync.Mutex
	items []T
}

type Queue[T any] interface {
	Get(ctx context.Context) (T, error)
	Put(ctx context.Context, item T) error
	Size() int
}

func NewFiFo[T any]() *FiFo[T] {
	return &FiFo[T]{}
}

func (q *FiFo[T]) Get(ctx context.Context) (T, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	for len(q.items) == 0 {
		q.mu.Unlock()
		select {
		case <-ctx.Done():
			q.mu.Lock()
			var zero T
			return zero, ctx.Err()
		default:
			q.mu.Lock()
			continue
		}
	}

	item := q.items[0]
	q.items = q.items[1:]
	return item, nil
}

func (q *FiFo[T]) Put(ctx context.Context, item T) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		q.items = append(q.items, item)
		return nil
	}
}

func (q *FiFo[T]) Size() int {
	q.mu.Lock()
	defer q.mu.Unlock()

	return len(q.items)
}
