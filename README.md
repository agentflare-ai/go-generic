# Generic

A Go package providing generic utilities and data structures.

## Overview

This package contains reusable generic types and utilities for Go applications, with a focus on type safety and performance.

## Features

* **Atomic\[T]**: A type-safe atomic value with **value semantics**. Unlike `atomic.Value`, supports safe pass-by-value while maintaining shared atomic storage through closures.
* **SyncPool\[T]**: A type-safe wrapper around `sync.Pool` that provides compile-time type safety for pooled objects.
* **FiFo\[T]**: A thread-safe generic FIFO queue with context support and blocking semantics.

## Usage

### Atomic\[T]

`Atomic[T]` provides type-safe atomic operations with value semantics, eliminating the need for pointer passing and type assertions:

```go
package main

import (
    "github.com/agentflare-ai/generic"
)

func main() {
    // Create an atomic value with optional default
    av := generic.MakeAtomic(42)
    
    // Or without default
    av := generic.MakeAtomic[string]()
    
    // All operations are type-safe (no type assertions needed)
    av.Store("hello")
    value := av.Load() // Returns string, not interface{}
    
    old := av.Swap("world") // Returns "hello"
    swapped := av.CompareAndSwap("world", "goodbye") // Returns true
}
```

**Pass-by-value semantics** - copies share the same atomic storage:

```go
av := generic.MakeAtomic(0)

// Pass by value to functions
processAtomic(av)

// Embed in structs
type Counter struct {
    count generic.Atomic[int]
}

func (c Counter) Increment() {  // Value receiver works!
    for {
        old := c.count.Load()
        if c.count.CompareAndSwap(old, old+1) {
            break
        }
    }
}

func processAtomic(a generic.Atomic[int]) {
    a.Store(100)  // Modifies original atomic storage
}
```

With custom types:

```go
type Config struct {
    Timeout int
    Retries int
}

// Can pass by value safely
av := generic.MakeAtomic(&Config{Timeout: 30, Retries: 3})

// All copies access the same atomic value
av2 := av
av2.Store(&Config{Timeout: 60, Retries: 5})
fmt.Println(av.Load().Timeout) // 60
```

**Performance**: \~1-2ns overhead vs `atomic.Value` for type safety and value semantics.

### SyncPool

```go
package main

import (
    "github.com/agentflare-ai/generic"
)

func main() {
    // Create a pool for strings
    pool := &generic.SyncPool[string]{}

    // Put values in the pool
    pool.Put("hello")
    pool.Put("world")

    // Get values from the pool
    value := pool.Get() // May return "hello", "world", or zero value
}
```

With a New function:

```go
pool := &generic.SyncPool[int]{}
pool.New = func() any { return 42 }

value := pool.Get() // Returns 42 if pool is empty
```

### FiFo\[T]

A thread-safe generic FIFO queue that blocks when empty and supports context cancellation:

```go
package main

import (
    "context"
    "time"
    "github.com/agentflare-ai/go-generic"
)

func main() {
    // Create a queue
    queue := generic.NewFiFo[string]()

    ctx := context.Background()

    // Put items (non-blocking)
    queue.Put(ctx, "hello")
    queue.Put(ctx, "world")

    // Get items (blocks if empty)
    item1, _ := queue.Get(ctx) // "hello"
    item2, _ := queue.Get(ctx) // "world"

    // Get with timeout when empty
    ctxTimeout, _ := context.WithTimeout(ctx, 100*time.Millisecond)
    _, err := queue.Get(ctxTimeout) // Returns context.DeadlineExceeded
}
```

**Key Features:**

* **Thread-safe**: Safe for concurrent use by multiple goroutines
* **Blocking semantics**: Get() blocks when queue is empty
* **Context support**: Respects context cancellation for both operations
* **Visibility**: `Size()` reports queued item count for instrumentation
* **Generic**: Type-safe for any Go type
* **Memory efficient**: Minimal allocations, reuses underlying slice

**Performance**: \~6-8ns per operation with 45B allocation overhead for Put operations.

## Performance Comparison

Benchmark results comparing FiFo\[T] against alternative queue implementations (Apple M3 Max, Go 1.25.3):

| Implementation | Put (ns/op) | Get (ns/op) | Put (B/op) | Concurrent (ns/op) |
| -------------- | ----------- | ----------- | ---------- | ------------------ |
| **FiFo\[T]**   | 8.21        | 6.06        | 45         | 173.9              |
| Mutex + Slice  | 8.54        | 6.02        | 47         | 167.8              |
| Channel (buf)  | 20.28       | 21.41       | 0          | 169.6              |

**Tradeoffs:**

* **FiFo\[T] vs Mutex+Slice**: Nearly identical performance with FiFo having slightly lower Put latency and memory usage. FiFo provides cleaner blocking semantics and context support out of the box.

* **FiFo\[T] vs Channel**: 2.5-3x faster operations with the tradeoff of 45B allocation per Put vs channel's zero-allocation approach. FiFo blocks on empty Get() while channels would require separate synchronization.

* **Blocking vs Non-blocking**: FiFo uses blocking semantics (Get waits for items), making it suitable for producer-consumer patterns. For non-blocking use cases, consider channel-based approaches.

* **Memory**: FiFo has minimal heap allocations (45B per Put) vs mutex+slice (47B). Channels use no heap allocations but require pre-sizing buffers.

Choose FiFo\[T] when you need:

* Clean blocking producer-consumer semantics
* Context cancellation support
* Minimal memory overhead
* Type safety with generics

```bash
go get github.com/agentflare-ai/generic
```

## Requirements

* Go 1.18 or later (for generics support)

## Testing

Run the test suite:

```bash
go test ./...
```

Run with coverage:

```bash
go test -cover ./...
```

## Contributing

Contributions are welcome! Please ensure all tests pass and add tests for new functionality.

## License

This project is licensed under the MIT License.
