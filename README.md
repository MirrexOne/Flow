# Flow - Lazy Stream Processing for Go

A powerful, functional stream processing library for Go with lazy evaluation.

## Features

- **Lazy evaluation** - operations are only executed when needed
- **Universal ForEach** - accepts ANY function through reflection
- **Chainable operations** - fluent API for elegant code
- **Zero dependencies** - only uses Go standard library

## Installation

```bash
go get github.com/MirrexOne/Flow
```

Requires Go 1.23+ for iterator support.

## Quick Start

```go
package main

import (
    "fmt"
    . "github.com/MirrexOne/Flow"
)

func main() {
    // Multiple ways to create flows:
    
    // From slice
    s := []int{1, 2, 3, 4, 5}
    NewFlow(s).ForEach(fmt.Print)  // Output: 12345
    
    // Variadic constructor
    Of(1, 2, 3, 4, 5).ForEach(func(x int) {
        fmt.Print(x, " ")
    })  // Output: 1 2 3 4 5
    
    // Complex pipeline
    Values(1, 2, 3, 4, 5).
        Filter(func(x int) bool { return x%2 == 0 }).
        Map(func(x int) int { return x * x }).
        ForEach(func(x int) {
            fmt.Printf("Result: %d\n", x)
        })
}
```

## Performance

### Some Benchmarks

```
Key Metrics (AMD Ryzen 5 7600X) on my PC:
┌───────────────────────────────────────────────┐
│ Simple Operations (10 items)                │
│   Traditional Loop:     3 ns/op   0 allocs  │
│   Flow Reduce:         91 ns/op   2 allocs  │
│   > Overhead: ~90ns            │
│                                              │
│ Complex Pipeline (100 items)                │
│   Traditional:         39 ns/op   0 allocs  │
│   Flow Pipeline:      758 ns/op   6 allocs  │
│   Flow Lazy:          311 ns/op   9 allocs  │
│   > Lazy is 2.4x faster than full pipeline │
└───────────────────────────────────────────────┘
```

### Why Flow is Fast

1. **Lazy Evaluation** - Only processes what's needed
2. **Zero-copy operations** - Minimal memory allocations
3. **Optimized hot paths** - Critical sections hand-tuned

## API Overview

### Stream Creation

```go
// Universal constructor - works with any input type
FlowOf[int](42)                         // Single value
FlowOf[int](1, 2, 3, 4, 5)              // Variadic arguments  
FlowOf[int]([]int{1, 2, 3})             // From slice
FlowOf[int]([3]int{1, 2, 3})            // From array
FlowOf[int](ch)                         // From channel
FlowOf[int](anotherFlow)                // From existing Flow
FlowOf[string](map[string]int{...})     // Map keys
FlowOf[int](map[string]int{...})        // Map values

// Alternative constructors
NewFlow([]int{1, 2, 3})        // From slice
Of(1, 2, 3)                    // Variadic arguments
Values(1, 2, 3)                // Alternative variadic
Single(42)                     // Single value
Empty[int]()                   // Empty flow

// Generators
Range(1, 10)                   // Numbers from 1 to 9
Infinite(func(i int) T)        // Infinite stream
FromChannel(ch)                // From channel
FromFunc(generator)            // Custom generator

// Backward compatibility
From([]int{1, 2, 3})           // Alias for NewFlow
```

### Intermediate Operations (Lazy)

These operations return a new `Flow` and are not executed until a terminal operation is called:

```go
.Filter(predicate)             // Keep matching elements
.Map(mapper)                   // Transform elements (same type)
.Take(n)                       // First n elements
.Skip(n)                       // Skip first n elements
.TakeWhile(predicate)          // Take while condition is true
.SkipWhile(predicate)          // Skip while condition is true
.Peek(action)                  // Debug/side effects
.Concat(other)                 // Append another flow
.Merge(others...)              // Merge multiple flows

// Standalone functions (type transformations)
MapTo(flow, mapper)            // Transform to different type
Distinct(flow)                 // Remove duplicates
FlatMap(flow, mapper)          // Flatten nested flows
Chunk(flow, size)              // Group into fixed-size chunks
Window(flow, size, step)       // Sliding/tumbling windows
```

### Terminal Operations (Execute)

These operations consume the stream and produce a result:

```go
.ForEach(fn)                   // Execute function for each element (ANY function!)
.ForEachFunc(fn)               // Type-safe version (faster)
.Collect()                     // Gather into slice
.Count()                       // Count elements
.Reduce(initial, reducer)      // Combine elements
.First()                       // Get first element
.Last()                        // Get last element
.AnyMatch(predicate)           // Check if any match
.AllMatch(predicate)           // Check if all match
.NoneMatch(predicate)          // Check if none match
.FindFirst(predicate)          // Find first matching element
.ToChannel(bufferSize)         // Convert to channel

// Standalone terminal operations
GroupBy(flow, keyFunc)         // Group by key into map
Partition(flow, predicate)     // Split into matching/non-matching
```

## Complete Examples

### Basic Pipeline
```go
Range(1, 11).
    Filter(func(x int) bool { return x%2 == 0 }).
    Map(func(x int) int { return x * x }).
    Take(3).
    ForEach(func(x int) { fmt.Println(x) })
// Output: 4, 16, 36
```

### Working with Structs
```go
type Person struct {
    Name string
    Age  int
}

Of(
    Person{"Alice", 25},
    Person{"Bob", 30},
    Person{"Charlie", 35},
).
    Filter(func(p Person) bool { return p.Age > 25 }).
    ForEach(func(p Person) {
        fmt.Printf("%s is %d\n", p.Name, p.Age)
    })
```

### Infinite Streams
```go
// Fibonacci sequence
Infinite(func(i int) int {
    if i < 2 { return i }
    a, b := 0, 1
    for j := 2; j <= i; j++ {
        a, b = b, a+b
    }
    return b
}).Take(10).Collect()
// [0, 1, 1, 2, 3, 5, 8, 13, 21, 34]
```

### Advanced Operations
```go
// FlatMap
words := Of("hello", "world")
FlatMap(words, func(word string) Flow[rune] {
    return NewFlow([]rune(word))
}).Collect()  // ['h','e','l','l','o','w','o','r','l','d']

// Chunk
Chunk(Range(1, 11), 3).ForEach(func(chunk []int) {
    fmt.Println(chunk)
})  // [1,2,3] [4,5,6] [7,8,9] [10]

// Combine (formerly Zip)
names := Of("Alice", "Bob")
ages := Of(25, 30)
Combine(names, ages).ForEach(func(pair Pair[string, int]) {
    fmt.Printf("%s: %d\n", pair.First, pair.Second)
})

// CombineWith for custom combination
CombineWith(names, ages, func(name string, age int) string {
    return fmt.Sprintf("%s is %d", name, age)
}).ForEach(fmt.Println)

// Merge multiple flows
flow1 := From([]int{1, 2, 3})
flow2 := From([]int{4, 5, 6})
flow1.Merge(flow2).Collect()  // [1, 2, 3, 4, 5, 6]

// GroupBy operation
people := []Person{{Name: "Alice", Age: 25}, {Name: "Bob", Age: 30}, {Name: "Charlie", Age: 25}}
byAge := GroupBy(From(people), func(p Person) int { return p.Age })
// Result: map[25:[{Alice 25} {Charlie 25}] 30:[{Bob 30}]]

// Partition operation
evens, odds := Partition(Range(1, 11), func(x int) bool { return x%2 == 0 })
// evens: [2, 4, 6, 8, 10], odds: [1, 3, 5, 7, 9]

// Window operation (sliding windows)
Window(Range(1, 6), 3, 1).ForEach(func(window []int) {
    fmt.Println(window)
})  // [1,2,3] [2,3,4] [3,4,5]
```

## Performance Deep Dive

### Detailed Benchmarks

| Operation | Time | Memory | Allocs | vs Loop |
|-----------|------|--------|--------|----------|
| **Best Case - Small Data (10 items)** |
| Traditional loop | 3 ns | 0 B | 0 | 1.0x |
| Flow.Reduce() | 91 ns | 64 B | 2 | 30x |
| Flow.Take(3) | 247 ns | 280 B | 7 | 82x |
| **Real World - Medium Data (100 items)** |
| Traditional loop | 39 ns | 0 B | 0 | 1.0x |
| Flow full pipeline | 758 ns | 192 B | 6 | 19x |
| Flow with lazy eval | 311 ns | 344 B | 9 | 8x |
| **Large Data (1000+ items)** |
| Filter | 3,896 ns | 112 B | 4 | - |
| Map | 4,652 ns | 112 B | 4 | - |
| Distinct | 1,767 ns | 704 B | 11 | - |
| Chunk(10) | 7,965 ns | 8,232 B | 106 | - |

### When to Use Flow vs Loops

```go
// USE FLOW when:
// - Readability matters
// - Complex transformations
// - Lazy evaluation needed
// - Working with streams
result := flow.From(data).
    Filter(isValid).
    Map(transform).
    Take(10).
    Collect()

// USE LOOPS when:
// - Ultra-hot path (< 100ns)
// - Simple iteration
// - Zero-alloc required
for _, v := range data {
    sum += v
}
```

## License

MIT License

## Author

Created by MirrexOne
