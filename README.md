# Generic

A Go package providing generic utilities and data structures.

## Overview

This package contains reusable generic types and utilities for Go applications, with a focus on type safety and performance.

## Features

- **AtomicValue[T]**: A type-safe wrapper around `atomic.Value` that provides compile-time type safety for atomic operations on generic types.
- **SyncPool[T]**: A type-safe wrapper around `sync.Pool` that provides compile-time type safety for pooled objects.

## Usage

### AtomicValue

```go
package main

import (
    "github.com/agentflare-ai/generic"
)

func main() {
    // Create an atomic value for strings
    var av generic.AtomicValue[string]
    
    // Store a value
    av.Store("hello")
    
    // Load the value
    value := av.Load() // Returns "hello"
    
    // Swap with a new value and get the old one
    old := av.Swap("world") // Returns "hello"
    
    // Compare and swap (only swaps if current value matches)
    swapped := av.CompareAndSwap("world", "goodbye") // Returns true
}
```

With custom types:

```go
type Config struct {
    Timeout int
    Retries int
}

var av generic.AtomicValue[*Config]
av.Store(&Config{Timeout: 30, Retries: 3})

// Safe concurrent access
config := av.Load()
newConfig := &Config{Timeout: 60, Retries: 5}
av.CompareAndSwap(config, newConfig)
```

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
