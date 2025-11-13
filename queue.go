package generic

import (
	"context"
	"errors"
)

var ErrEmptyQueue = errors.New("queue is empty")

// FiFo is a generic, channel-token queue that preserves FIFO ordering
// and supports context-aware Enqueue/Dequeue plus a stop-the-world Snapshot.
// It uses two single-slot channels:
//   - items: holds a non-empty slice when queue has elements
//   - empty: holds a token when queue is empty
//
// No mutexes are required; synchronization is via token ownership.
type FiFo[T any] struct {
	items chan []T      // cap=1; present when non-empty
	empty chan struct{} // cap=1; present when empty
}

type Queue[T any] interface {
	Put(ctx context.Context, x T) error
	TryPut(x T) bool
	Get(ctx context.Context) (T, error)
	TryGet() (T, bool)
	IsEmpty() bool
	Size() int
}

func NewFiFo[T any]() *FiFo[T] {
	q := &FiFo[T]{
		items: make(chan []T, 1),
		empty: make(chan struct{}, 1),
	}
	q.empty <- struct{}{} // start empty
	return q
}

func (q *FiFo[T]) Size() int {
	select {
	case items := <-q.items:
		defer func() { q.items <- items }()
		return len(items)
	case <-q.empty:
		defer func() { q.empty <- struct{}{} }()
		return 0
	}
}

// Enqueue appends x, respecting ctx cancellation.
//
//go:inline
func (q *FiFo[T]) Put(ctx context.Context, x T) error {
	var s []T
	select {
	case s = <-q.items:
	case <-q.empty:
	case <-ctx.Done():
		return ctx.Err()
	}
	s = append(s, x)
	q.items <- s
	return ctx.Err()
}

// TryEnqueue attempts to enqueue without blocking; returns true if successful.
//
//go:inline
func (q *FiFo[T]) TryPut(x T) bool {
	select {
	case s := <-q.items:
		s = append(s, x)
		q.items <- s
		return true
	case <-q.empty:
		s := []T{x}
		q.items <- s
		return true
	default:
		return false
	}
}

// Dequeue removes and returns the next item, or ctx error if cancelled.
//
//go:inline
func (q *FiFo[T]) Get(ctx context.Context) (T, error) {
	var zero T
	var s []T
	select {
	case s = <-q.items:
	case <-ctx.Done():
		return zero, ctx.Err()
	}
	x := s[0]
	s = s[1:]
	if len(s) == 0 {
		q.empty <- struct{}{}
	} else {
		q.items <- s
	}
	return x, nil
}

// TryDequeue attempts to dequeue without blocking; returns (zero,false) if empty.
//
//go:inline
func (q *FiFo[T]) TryGet() (T, bool) {
	var zero T
	select {
	case s := <-q.items:
		x := s[0]
		s = s[1:]
		if len(s) == 0 {
			select {
			case q.empty <- struct{}{}:
			default:
			}
		} else {
			select {
			case q.items <- s:
			default:
			}
		}
		return x, true
	default:
		return zero, false
	}
}

// IsEmpty returns true if the queue is empty. This is a non-blocking hint.
//
//go:inline
func (q *FiFo[T]) IsEmpty() bool {
	return len(q.empty) == 1
}

// Snapshot performs a brief stop-the-world capture of the current queue contents.
// It acquires the token (items or empty), clones the slice, and restores the token.
func (q *FiFo[T]) Snapshot(ctx context.Context) ([]T, error) {
	var s []T
	tookItems := false
	select {
	case s = <-q.items:
		tookItems = true
	case <-q.empty:
		s = nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
	cp := append([]T(nil), s...)
	if tookItems {
		q.items <- s
	} else {
		q.empty <- struct{}{}
	}
	return cp, nil
}
