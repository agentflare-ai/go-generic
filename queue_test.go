package generic

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestFiFo_BasicOperations(t *testing.T) {
	q := NewFiFo[int]()

	ctx := context.Background()

	// Put some items
	if err := q.Put(ctx, 1); err != nil {
		t.Errorf("unexpected error putting item: %v", err)
	}
	if err := q.Put(ctx, 2); err != nil {
		t.Errorf("unexpected error putting item: %v", err)
	}
	if err := q.Put(ctx, 3); err != nil {
		t.Errorf("unexpected error putting item: %v", err)
	}

	// Get items back in FIFO order
	item1, err := q.Get(ctx)
	if err != nil {
		t.Errorf("unexpected error getting item: %v", err)
	}
	if item1 != 1 {
		t.Errorf("expected 1, got %d", item1)
	}

	item2, err := q.Get(ctx)
	if err != nil {
		t.Errorf("unexpected error getting item: %v", err)
	}
	if item2 != 2 {
		t.Errorf("expected 2, got %d", item2)
	}

	item3, err := q.Get(ctx)
	if err != nil {
		t.Errorf("unexpected error getting item: %v", err)
	}
	if item3 != 3 {
		t.Errorf("expected 3, got %d", item3)
	}

	// Test getting from empty queue with timeout
	ctxTimeout, cancel := context.WithTimeout(ctx, 10*time.Millisecond)
	defer cancel()
	_, err = q.Get(ctxTimeout)
	if err != context.DeadlineExceeded {
		t.Errorf("expected timeout error, got %v", err)
	}
}

func TestFiFo_Size(t *testing.T) {
	q := NewFiFo[int]()
	ctx := context.Background()

	if size := q.Size(); size != 0 {
		t.Fatalf("expected initial size 0, got %d", size)
	}

	if err := q.Put(ctx, 10); err != nil {
		t.Fatalf("put failed: %v", err)
	}
	if size := q.Size(); size != 1 {
		t.Fatalf("expected size 1, got %d", size)
	}

	if err := q.Put(ctx, 20); err != nil {
		t.Fatalf("put failed: %v", err)
	}
	if size := q.Size(); size != 2 {
		t.Fatalf("expected size 2, got %d", size)
	}

	if _, err := q.Get(ctx); err != nil {
		t.Fatalf("get failed: %v", err)
	}
	if size := q.Size(); size != 1 {
		t.Fatalf("expected size 1 after get, got %d", size)
	}

	if _, err := q.Get(ctx); err != nil {
		t.Fatalf("get failed: %v", err)
	}
	if size := q.Size(); size != 0 {
		t.Fatalf("expected size 0 after draining, got %d", size)
	}
}

func TestFiFo_StringType(t *testing.T) {
	q := NewFiFo[string]()

	ctx := context.Background()

	// Put strings
	if err := q.Put(ctx, "hello"); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if err := q.Put(ctx, "world"); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Get strings back
	s1, err := q.Get(ctx)
	if err != nil || s1 != "hello" {
		t.Errorf("expected 'hello', got %q, err: %v", s1, err)
	}

	s2, err := q.Get(ctx)
	if err != nil || s2 != "world" {
		t.Errorf("expected 'world', got %q, err: %v", s2, err)
	}

	// Test timeout on empty queue
	ctxTimeout, cancel := context.WithTimeout(ctx, 5*time.Millisecond)
	defer cancel()
	_, err = q.Get(ctxTimeout)
	if err != context.DeadlineExceeded {
		t.Errorf("expected timeout, got %v", err)
	}
}

func TestFiFo_CustomType(t *testing.T) {
	type TestStruct struct {
		ID   int
		Name string
	}

	q := NewFiFo[TestStruct]()
	ctx := context.Background()

	item1 := TestStruct{ID: 1, Name: "first"}
	item2 := TestStruct{ID: 2, Name: "second"}

	// Put custom types
	if err := q.Put(ctx, item1); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if err := q.Put(ctx, item2); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Get them back
	got1, err := q.Get(ctx)
	if err != nil || got1 != item1 {
		t.Errorf("expected %+v, got %+v, err: %v", item1, got1, err)
	}

	got2, err := q.Get(ctx)
	if err != nil || got2 != item2 {
		t.Errorf("expected %+v, got %+v, err: %v", item2, got2, err)
	}

	// Test timeout on empty queue
	ctxTimeout, cancel := context.WithTimeout(ctx, 5*time.Millisecond)
	defer cancel()
	_, err = q.Get(ctxTimeout)
	if err != context.DeadlineExceeded {
		t.Errorf("expected timeout, got %v", err)
	}
}

func TestFiFo_ContextCancellation(t *testing.T) {
	q := NewFiFo[int]()

	// Test Get with cancelled context on empty queue
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := q.Get(ctx)
	if err != context.Canceled {
		t.Errorf("expected context.Canceled, got %v", err)
	}

	// Test Put with cancelled context
	ctx2, cancel2 := context.WithCancel(context.Background())
	cancel2()

	err = q.Put(ctx2, 42)
	if err != context.Canceled {
		t.Errorf("expected context.Canceled, got %v", err)
	}

	// Test Get with cancelled context when queue has items - should still return the item
	ctx3, cancel3 := context.WithCancel(context.Background())
	q.Put(ctx3, 99) // Put an item
	cancel3()       // Cancel context - but Get should still work since item is available

	val, err := q.Get(ctx3)
	if err != nil {
		t.Errorf("expected no error when getting available item, got %v", err)
	}
	if val != 99 {
		t.Errorf("expected 99, got %d", val)
	}
}

func TestFiFo_ConcurrentAccess(t *testing.T) {
	q := NewFiFo[int]()
	ctx := context.Background()

	var wg sync.WaitGroup
	const numProducers = 5
	const numConsumers = 5
	const itemsPerProducer = 20 // Reduced for faster test

	// Start producers
	produced := make(chan int, numProducers*itemsPerProducer)
	for i := 0; i < numProducers; i++ {
		wg.Add(1)
		go func(producerID int) {
			defer wg.Done()
			for j := 0; j < itemsPerProducer; j++ {
				item := producerID*itemsPerProducer + j
				if err := q.Put(ctx, item); err != nil {
					t.Errorf("producer %d failed to put item %d: %v", producerID, item, err)
					return
				}
				produced <- item
			}
		}(i)
	}

	// Start consumers
	collected := make(chan int, numProducers*itemsPerProducer)
	for i := 0; i < numConsumers; i++ {
		wg.Add(1)
		go func(consumerID int) {
			defer wg.Done()
			for {
				// Use timeout to avoid hanging
				ctxTimeout, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
				item, err := q.Get(ctxTimeout)
				cancel()

				if err != nil {
					// Check if we've collected all expected items
					if len(collected) >= numProducers*itemsPerProducer {
						break
					}
					continue
				}
				collected <- item

				// Check if we've collected all items
				if len(collected) >= numProducers*itemsPerProducer {
					break
				}
			}
		}(i)
	}

	// Wait for all producers to finish
	wg.Wait()
	close(produced)

	// Wait a bit for consumers to finish
	time.Sleep(200 * time.Millisecond)

	// Close collected channel
	close(collected)

	// Verify we got all items
	items := make([]int, 0, numProducers*itemsPerProducer)
	for item := range collected {
		items = append(items, item)
	}

	if len(items) != numProducers*itemsPerProducer {
		t.Errorf("expected %d items, got %d", numProducers*itemsPerProducer, len(items))
	}

	// Verify all items are present (FIFO order not guaranteed with concurrent access)
	expectedItems := make(map[int]bool)
	for i := 0; i < numProducers*itemsPerProducer; i++ {
		expectedItems[i] = true
	}

	for _, item := range items {
		if !expectedItems[item] {
			t.Errorf("unexpected item: %d", item)
		}
		delete(expectedItems, item)
	}

	if len(expectedItems) > 0 {
		t.Errorf("missing items: %v", expectedItems)
	}
}

func TestFiFo_SequentialOperations(t *testing.T) {
	q := NewFiFo[int]()
	ctx := context.Background()

	// Test multiple put/get cycles
	for cycle := 0; cycle < 10; cycle++ {
		// Put items
		for i := 0; i < 5; i++ {
			if err := q.Put(ctx, cycle*10+i); err != nil {
				t.Errorf("cycle %d: failed to put item %d: %v", cycle, i, err)
			}
		}

		// Get items back
		for i := 0; i < 5; i++ {
			expected := cycle*10 + i
			got, err := q.Get(ctx)
			if err != nil {
				t.Errorf("cycle %d: failed to get item %d: %v", cycle, i, err)
			}
			if got != expected {
				t.Errorf("cycle %d: expected %d, got %d", cycle, expected, got)
			}
		}

		// Should be empty
		ctxTimeout, cancel := context.WithTimeout(ctx, 5*time.Millisecond)
		_, err := q.Get(ctxTimeout)
		cancel()
		if err != context.DeadlineExceeded {
			t.Errorf("cycle %d: expected timeout when queue empty, got %v", cycle, err)
		}
	}
}

func TestFiFo_Timeout(t *testing.T) {
	q := NewFiFo[int]()

	// Test Get with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	start := time.Now()
	_, err := q.Get(ctx)
	elapsed := time.Since(start)

	if err != context.DeadlineExceeded {
		t.Errorf("expected context.DeadlineExceeded, got %v", err)
	}

	if elapsed < 5*time.Millisecond || elapsed > 20*time.Millisecond {
		t.Errorf("timeout took %v, expected ~10ms", elapsed)
	}
}

// Benchmark implementations for comparison

type MutexQueue[T any] struct {
	mu    sync.Mutex
	items []T
}

func NewMutexQueue[T any]() *MutexQueue[T] {
	return &MutexQueue[T]{}
}

func (q *MutexQueue[T]) Get(ctx context.Context) (T, error) {
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

func (q *MutexQueue[T]) Put(ctx context.Context, item T) error {
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

type ChanQueue[T any] struct {
	items chan T
}

func NewChanQueue[T any](bufferSize int) *ChanQueue[T] {
	return &ChanQueue[T]{
		items: make(chan T, bufferSize),
	}
}

func (q *ChanQueue[T]) Get(ctx context.Context) (T, error) {
	select {
	case item := <-q.items:
		return item, nil
	case <-ctx.Done():
		var zero T
		return zero, ctx.Err()
	}
}

func (q *ChanQueue[T]) Put(ctx context.Context, item T) error {
	select {
	case q.items <- item:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Benchmarks

func BenchmarkFiFo_Put(b *testing.B) {
	q := NewFiFo[int]()
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		q.Put(ctx, i)
	}
}

func BenchmarkFiFo_Get(b *testing.B) {
	q := NewFiFo[int]()
	ctx := context.Background()

	// Pre-fill queue
	for i := 0; i < b.N; i++ {
		q.Put(ctx, i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		q.Get(ctx)
	}
}

func BenchmarkMutexQueue_Put(b *testing.B) {
	q := NewMutexQueue[int]()
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		q.Put(ctx, i)
	}
}

func BenchmarkMutexQueue_Get(b *testing.B) {
	q := NewMutexQueue[int]()
	ctx := context.Background()

	// Pre-fill queue
	for i := 0; i < b.N; i++ {
		q.Put(ctx, i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		q.Get(ctx)
	}
}

func BenchmarkChanQueue_Put(b *testing.B) {
	q := NewChanQueue[int](b.N) // Make buffer large enough
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		q.Put(ctx, i)
	}
}

func BenchmarkChanQueue_Get(b *testing.B) {
	q := NewChanQueue[int](b.N) // Make buffer large enough
	ctx := context.Background()

	// Pre-fill queue
	for i := 0; i < b.N; i++ {
		q.Put(ctx, i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		q.Get(ctx)
	}
}

// Concurrent benchmarks

func BenchmarkFiFo_Concurrent(b *testing.B) {
	q := NewFiFo[int]()
	ctx := context.Background()

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			q.Put(ctx, i)
			q.Get(ctx)
			i++
		}
	})
}

func BenchmarkMutexQueue_Concurrent(b *testing.B) {
	q := NewMutexQueue[int]()
	ctx := context.Background()

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			q.Put(ctx, i)
			q.Get(ctx)
			i++
		}
	})
}

func BenchmarkChanQueue_Concurrent(b *testing.B) {
	q := NewChanQueue[int](1024)
	ctx := context.Background()

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			q.Put(ctx, i)
			q.Get(ctx)
			i++
		}
	})
}
