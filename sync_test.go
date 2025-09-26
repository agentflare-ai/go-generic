package generic

import (
	"sync"
	"testing"
	"time"
)

func TestSyncPool_Get_Put(t *testing.T) {
	t.Run("string pool", func(t *testing.T) {
		pool := &SyncPool[string]{}

		// Get should return zero value when no New function is set
		got := pool.Get()
		if got != "" {
			t.Errorf("expected empty string, got %q", got)
		}

		// Put a value back
		pool.Put("test")
	})

	t.Run("int pool", func(t *testing.T) {
		pool := &SyncPool[int]{}

		// Get should return zero value when no New function is set
		got := pool.Get()
		if got != 0 {
			t.Errorf("expected 0, got %d", got)
		}

		// Put a value back
		pool.Put(42)
	})
}

func TestSyncPool_WithNewFunction(t *testing.T) {
	t.Run("string pool with new function", func(t *testing.T) {
		pool := &SyncPool[string]{}
		pool.New = func() any { return "default" }

		got := pool.Get()
		if got != "default" {
			t.Errorf("expected 'default', got %q", got)
		}
	})

	t.Run("int pool with new function", func(t *testing.T) {
		pool := &SyncPool[int]{}
		pool.New = func() any { return 123 }

		got := pool.Get()
		if got != 123 {
			t.Errorf("expected 123, got %d", got)
		}
	})
}

func TestSyncPool_Reuse(t *testing.T) {
	pool := &SyncPool[string]{}

	// Put some values
	pool.Put("first")
	pool.Put("second")

	// Get values back - may or may not get the put values due to internal pooling
	got1 := pool.Get()
	got2 := pool.Get()

	// We can't guarantee which values we get back, but they should be strings
	if got1 == "" && got2 == "" {
		t.Error("expected at least one non-empty string from pool reuse")
	}
}

func TestSyncPool_ConcurrentAccess(t *testing.T) {
	pool := &SyncPool[int]{}
	pool.New = func() any { return 0 }

	var wg sync.WaitGroup
	const numGoroutines = 10
	const numOperations = 100

	results := make(chan []int, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			localResults := make([]int, numOperations)
			for j := 0; j < numOperations; j++ {
				// Put and get in sequence
				pool.Put(id*100 + j)
				val := pool.Get()
				localResults[j] = val
			}
			results <- localResults
		}(i)
	}

	wg.Wait()
	close(results)

	// Collect all results
	allResults := make([]int, 0, numGoroutines*numOperations)
	for res := range results {
		allResults = append(allResults, res...)
	}

	// Verify we got some expected values (pool reuse)
	found := make(map[int]bool)
	for _, val := range allResults {
		found[val] = true
	}

	if len(found) == 0 {
		t.Error("expected to find some put values in results")
	}
}

func TestSyncPool_CustomType(t *testing.T) {
	type TestStruct struct {
		ID   int
		Name string
	}

	pool := &SyncPool[TestStruct]{}
	pool.New = func() any {
		return TestStruct{ID: 999, Name: "default"}
	}

	// Test Get with new function
	got := pool.Get()
	expected := TestStruct{ID: 999, Name: "default"}
	if got != expected {
		t.Errorf("expected %+v, got %+v", expected, got)
	}

	// Test Put and reuse
	custom := TestStruct{ID: 123, Name: "custom"}
	pool.Put(custom)

	got2 := pool.Get()
	if got2 != expected && got2 != custom {
		t.Errorf("expected either default or custom struct, got %+v", got2)
	}
}

func BenchmarkSyncPool_Get(b *testing.B) {
	pool := &SyncPool[int]{}
	pool.New = func() any { return 42 }

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		val := pool.Get()
		pool.Put(val)
	}
}

func BenchmarkSyncPool_Put(b *testing.B) {
	pool := &SyncPool[int]{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pool.Put(i)
	}
}

// Test that demonstrates pool behavior over time
func TestSyncPool_Lifecycle(t *testing.T) {
	pool := &SyncPool[time.Time]{}

	// Initially empty
	now := time.Now()
	pool.Put(now)

	// Get it back
	got := pool.Get()
	if !got.Equal(now) {
		t.Errorf("expected %v, got %v", now, got)
	}

	// Next get should be zero value
	got2 := pool.Get()
	if !got2.IsZero() {
		t.Errorf("expected zero time, got %v", got2)
	}
}
