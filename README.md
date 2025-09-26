# Generic

A Go package providing generic utilities and data structures.

## Overview

This package contains reusable generic types and utilities for Go applications, with a focus on type safety and performance.

## Features

- **SyncPool[T]**: A type-safe wrapper around `sync.Pool` that provides compile-time type safety for pooled objects.

## Usage

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
