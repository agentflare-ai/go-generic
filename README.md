# Generic

A Go package providing generic utilities and data structures.

## Overview

This package contains reusable generic types and utilities for Go applications, with a focus on type safety and performance.

## Features

- **Atomic[T]**: A type-safe atomic value with **value semantics**. Unlike `atomic.Value`, supports safe pass-by-value while maintaining shared atomic storage through closures.
- **SyncPool[T]**: A type-safe wrapper around `sync.Pool` that provides compile-time type safety for pooled objects.

## Usage

### Atomic[T]

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

**Performance**: ~1-2ns overhead vs `atomic.Value` for type safety and value semantics.

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

## Installation

```bash
go get github.com/agentflare-ai/generic
```

## Requirements

- Go 1.18 or later (for generics support)

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
