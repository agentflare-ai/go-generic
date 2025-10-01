package generic

import (
	"sync"
	"testing"
)

func TestAtomicValue_Load_Store(t *testing.T) {
	t.Run("string value", func(t *testing.T) {
		var av AtomicValue[string]
		av.Store("hello")

		got := av.Load()
		if got != "hello" {
			t.Fatalf("expected 'hello', got %q", got)
		}
	})

	t.Run("int value", func(t *testing.T) {
		var av AtomicValue[int]
		av.Store(42)

		got := av.Load()
		if got != 42 {
			t.Fatalf("expected 42, got %d", got)
		}
	})

	t.Run("struct value", func(t *testing.T) {
		type TestStruct struct {
			ID   int
			Name string
		}

		var av AtomicValue[TestStruct]
		expected := TestStruct{ID: 123, Name: "test"}
		av.Store(expected)

		got := av.Load()
		if got != expected {
			t.Fatalf("expected %+v, got %+v", expected, got)
		}
	})

	t.Run("pointer value", func(t *testing.T) {
		type TestStruct struct {
			Value int
		}

		var av AtomicValue[*TestStruct]
		expected := &TestStruct{Value: 999}
		av.Store(expected)

		got := av.Load()
		if got != expected {
			t.Fatalf("expected pointer %p, got %p", expected, got)
		}
		if got.Value != 999 {
			t.Fatalf("expected Value 999, got %d", got.Value)
		}
	})
}

func TestAtomicValue_Swap(t *testing.T) {
	t.Run("swap string values", func(t *testing.T) {
		var av AtomicValue[string]
		av.Store("first")

		old := av.Swap("second")
		if old != "first" {
			t.Fatalf("expected 'first', got %q", old)
		}

		got := av.Load()
		if got != "second" {
			t.Fatalf("expected 'second', got %q", got)
		}
	})

	t.Run("swap int values", func(t *testing.T) {
		var av AtomicValue[int]
		av.Store(10)

		old := av.Swap(20)
		if old != 10 {
			t.Fatalf("expected 10, got %d", old)
		}

		got := av.Load()
		if got != 20 {
			t.Fatalf("expected 20, got %d", got)
		}
	})
}

func TestAtomicValue_CompareAndSwap(t *testing.T) {
	t.Run("successful compare and swap", func(t *testing.T) {
		var av AtomicValue[string]
		av.Store("initial")

		swapped := av.CompareAndSwap("initial", "updated")
		if !swapped {
			t.Fatal("expected successful swap")
		}

		got := av.Load()
		if got != "updated" {
			t.Fatalf("expected 'updated', got %q", got)
		}
	})

	t.Run("failed compare and swap", func(t *testing.T) {
		var av AtomicValue[string]
		av.Store("initial")

		swapped := av.CompareAndSwap("wrong", "updated")
		if swapped {
			t.Fatal("expected swap to fail")
		}

		got := av.Load()
		if got != "initial" {
			t.Fatalf("expected 'initial', got %q", got)
		}
	})

	t.Run("int compare and swap", func(t *testing.T) {
		var av AtomicValue[int]
		av.Store(100)

		// Successful swap
		swapped := av.CompareAndSwap(100, 200)
		if !swapped {
			t.Fatal("expected successful swap")
		}

		got := av.Load()
		if got != 200 {
			t.Fatalf("expected 200, got %d", got)
		}

		// Failed swap
		swapped = av.CompareAndSwap(100, 300)
		if swapped {
			t.Fatal("expected swap to fail")
		}

		got = av.Load()
		if got != 200 {
			t.Fatalf("expected 200 (unchanged), got %d", got)
		}
	})
}

func TestAtomicValue_ConcurrentAccess(t *testing.T) {
	t.Run("concurrent stores and loads", func(t *testing.T) {
		var av AtomicValue[int]
		av.Store(0)

		var wg sync.WaitGroup
		const numGoroutines = 10
		const numOperations = 100

		// Multiple goroutines storing values
		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				for j := 0; j < numOperations; j++ {
					av.Store(id*1000 + j)
					_ = av.Load()
				}
			}(i)
		}

		wg.Wait()

		// Final value should be some valid stored value
		final := av.Load()
		if final < 0 {
			t.Fatalf("unexpected final value: %d", final)
		}
	})

	t.Run("concurrent swaps", func(t *testing.T) {
		var av AtomicValue[string]
		av.Store("start")

		var wg sync.WaitGroup
		const numGoroutines = 10

		results := make(chan string, numGoroutines)

		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				old := av.Swap("goroutine")
				results <- old
			}(i)
		}

		wg.Wait()
		close(results)

		// Collect all old values
		oldValues := make([]string, 0, numGoroutines)
		for val := range results {
			oldValues = append(oldValues, val)
		}

		// At least one should have seen "start"
		foundStart := false
		for _, val := range oldValues {
			if val == "start" {
				foundStart = true
				break
			}
		}

		if !foundStart {
			t.Fatal("expected at least one goroutine to see 'start'")
		}
	})

	t.Run("concurrent compare and swap", func(t *testing.T) {
		var av AtomicValue[int]
		av.Store(0)

		var wg sync.WaitGroup
		const numGoroutines = 10

		successCount := make(chan bool, numGoroutines*10)

		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				for j := 0; j < 10; j++ {
					current := av.Load()
					swapped := av.CompareAndSwap(current, id*100+j)
					successCount <- swapped
				}
			}(i)
		}

		wg.Wait()
		close(successCount)

		// Count successful swaps
		successful := 0
		for success := range successCount {
			if success {
				successful++
			}
		}

		// Should have some successful swaps (not all will succeed due to contention)
		if successful == 0 {
			t.Fatal("expected at least some successful compare-and-swaps")
		}
	})
}

func TestAtomicValue_Multiple_Operations(t *testing.T) {
	t.Run("store, swap, compare-and-swap sequence", func(t *testing.T) {
		var av AtomicValue[string]

		// Initial store
		av.Store("first")
		if got := av.Load(); got != "first" {
			t.Fatalf("after Store: expected 'first', got %q", got)
		}

		// Swap
		old := av.Swap("second")
		if old != "first" {
			t.Fatalf("Swap returned: expected 'first', got %q", old)
		}
		if got := av.Load(); got != "second" {
			t.Fatalf("after Swap: expected 'second', got %q", got)
		}

		// CompareAndSwap success
		swapped := av.CompareAndSwap("second", "third")
		if !swapped {
			t.Fatal("CompareAndSwap should have succeeded")
		}
		if got := av.Load(); got != "third" {
			t.Fatalf("after CompareAndSwap: expected 'third', got %q", got)
		}

		// CompareAndSwap failure
		swapped = av.CompareAndSwap("wrong", "fourth")
		if swapped {
			t.Fatal("CompareAndSwap should have failed")
		}
		if got := av.Load(); got != "third" {
			t.Fatalf("after failed CompareAndSwap: expected 'third', got %q", got)
		}
	})
}

func BenchmarkAtomicValue_Load(b *testing.B) {
	var av AtomicValue[int]
	av.Store(42)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = av.Load()
	}
}

func BenchmarkAtomicValue_Store(b *testing.B) {
	var av AtomicValue[int]
	av.Store(0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		av.Store(i)
	}
}

func BenchmarkAtomicValue_Swap(b *testing.B) {
	var av AtomicValue[int]
	av.Store(0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		av.Swap(i)
	}
}

func BenchmarkAtomicValue_CompareAndSwap(b *testing.B) {
	var av AtomicValue[int]
	av.Store(0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		current := av.Load()
		av.CompareAndSwap(current, i)
	}
}

func BenchmarkAtomicValue_Concurrent_Load(b *testing.B) {
	var av AtomicValue[int]
	av.Store(42)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = av.Load()
		}
	})
}

func BenchmarkAtomicValue_Concurrent_Store(b *testing.B) {
	var av AtomicValue[int]
	av.Store(0)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			av.Store(1)
		}
	})
}

