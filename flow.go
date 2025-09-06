package flow

import (
	"fmt"
	"reflect"

	"iter"

	"github.com/MirrexOne/Flow/internal"
)

// Flow represents a lazy stream of elements that can be processed functionally.
// The zero value is not usable; use constructor functions like From, Range, etc.
type Flow[T any] struct {
	source iter.Seq[T]
}

// NewFlow creates a new Flow from a slice.
// The slice is not copied, so modifications to it may affect the stream.
//
// Example:
//
//	numbers := []int{1, 2, 3, 4, 5}
//	flow.NewFlow(numbers).ForEach(fmt.Println)
func NewFlow[T any](values []T) Flow[T] {
	return Flow[T]{
		source: func(yield func(T) bool) {
			for _, val := range values {
				if !yield(val) {
					return
				}
			}
		},
	}
}

// Single creates a Flow with a single value.
//
// Example:
//
//	flow.Single(42).ForEach(fmt.Println)
func Single[T any](value T) Flow[T] {
	return Flow[T]{
		source: func(yield func(T) bool) {
			yield(value)
		},
	}
}

// Empty creates an empty Flow with no elements.
//
// Example:
//
//	flow.Empty[int]().Count() // Returns 0
func Empty[T any]() Flow[T] {
	return Flow[T]{
		source: func(yield func(T) bool) {},
	}
}

// Of creates a Flow from variadic arguments.
//
// Example:
//
//	flow.Of(1, 2, 3, 4, 5).ForEach(fmt.Println)
//	flow.Of("hello", "world").ForEach(fmt.Println)
func Of[T any](values ...T) Flow[T] {
	return NewFlow(values)
}

func FromSlice[T any](values []T) Flow[T] {
	return NewFlow(values)
}

func Values[T any](values ...T) Flow[T] {
	return NewFlow(values)
}

// FlowOf creates a Flow from various input types using reflection.
// It's a universal constructor that intelligently handles:
//   - single values: FlowOf(42)
//   - variadic arguments: FlowOf(1, 2, 3)
//   - slices: FlowOf([]int{1, 2, 3})
//   - channels: FlowOf(ch)
//   - existing Flows: FlowOf(anotherFlow)
//   - arrays: FlowOf([3]int{1, 2, 3})
//   - maps (keys): FlowOf(map[string]int{"a": 1})
//
// The type parameter T must be explicitly specified.
//
// Example:
//
//	FlowOf[int](42)                     // single value
//	FlowOf[int](1, 2, 3)                // variadic
//	FlowOf[int]([]int{1, 2, 3})         // slice
//	FlowOf[string](ch)                  // channel
//	FlowOf[Person](existingFlow)        // existing flow
func FlowOf[T any](source interface{}, rest ...interface{}) Flow[T] {
	if len(rest) > 0 {
		allValues := make([]T, 0, len(rest)+1)

		if v, ok := source.(T); ok {
			allValues = append(allValues, v)
		} else {
			panic(fmt.Sprintf("FlowOf: first argument is not of type %T", *new(T)))
		}

		for i, val := range rest {
			if v, ok := val.(T); ok {
				allValues = append(allValues, v)
			} else {
				panic(fmt.Sprintf("FlowOf: argument %d is not of type %T", i+2, *new(T)))
			}
		}

		return NewFlow(allValues)
	}

	if source == nil {
		return Empty[T]()
	}

	// Try direct type assertions first (fastest)
	switch v := source.(type) {
	case Flow[T]:
		return v
	case []T:
		return NewFlow(v)
	case <-chan T:
		return FromChannel(v)
	case T:
		return Single(v)
	}

	// Use reflection for more complex types
	rv := reflect.ValueOf(source)
	rt := rv.Type()

	if rt.Kind() == reflect.Slice {
		sliceLen := rv.Len()
		result := make([]T, 0, sliceLen)
		for i := 0; i < sliceLen; i++ {
			if elem, ok := rv.Index(i).Interface().(T); ok {
				result = append(result, elem)
			} else {
				panic(fmt.Sprintf("FlowOf: slice element %d is not of type %T", i, *new(T)))
			}
		}
		return NewFlow(result)
	}

	if rt.Kind() == reflect.Array {
		arrayLen := rv.Len()
		result := make([]T, 0, arrayLen)
		for i := 0; i < arrayLen; i++ {
			if elem, ok := rv.Index(i).Interface().(T); ok {
				result = append(result, elem)
			} else {
				panic(fmt.Sprintf("FlowOf: array element %d is not of type %T", i, *new(T)))
			}
		}
		return NewFlow(result)
	}

	if rt.Kind() == reflect.Chan && rt.ChanDir() != reflect.SendDir {
		return Flow[T]{
			source: func(yield func(T) bool) {
				for {
					val, ok := rv.Recv()
					if !ok {
						break
					}
					if elem, ok := val.Interface().(T); ok {
						if !yield(elem) {
							return
						}
					} else {
						panic(fmt.Sprintf("FlowOf: channel element is not of type %T", *new(T)))
					}
				}
			},
		}
	}

	// Map: uses keys by default, falls back to values if keys don't match type T
	if rt.Kind() == reflect.Map {
		keys := rv.MapKeys()
		result := make([]T, 0, len(keys))
		for _, key := range keys {
			if elem, ok := key.Interface().(T); ok {
				result = append(result, elem)
			} else {
				if val := rv.MapIndex(key); val.IsValid() {
					if elem, ok := val.Interface().(T); ok {
						result = append(result, elem)
						continue
					}
				}
				panic(fmt.Sprintf("FlowOf: map elements are not of type %T", *new(T)))
			}
		}
		return NewFlow(result)
	}

	if val, ok := source.(T); ok {
		return Single(val)
	}

	panic(fmt.Sprintf("FlowOf: cannot convert %T to Flow[%T]", source, *new(T)))
}

// FromFunc creates a Flow from a generator function.
// The generator should call yield for each element and return when yield returns false.
//
// Example:
//
//	flow.FromFunc(func(yield func(int) bool) {
//	    for i := 0; i < 10; i++ {
//	        if !yield(i) {
//	            return
//	        }
//	    }
//	})
func FromFunc[T any](generator func(yield func(T) bool)) Flow[T] {
	return Flow[T]{source: generator}
}

// Range creates a Flow of integers from start (inclusive) to end (exclusive).
//
// Example:
//
//	flow.Range(1, 6).ForEach(fmt.Print) // Output: 12345
func Range(start, end int) Flow[int] {
	if start >= end {
		return Empty[int]()
	}
	return Flow[int]{
		source: func(yield func(int) bool) {
			for i := start; i < end; i++ {
				if !yield(i) {
					return
				}
			}
		},
	}
}

// Infinite creates an infinite Flow using a generator function.
// The generator receives the current index starting from 0.
// Use Take() or other limiting operations to avoid infinite loops.
//
// Example:
//
//	flow.Infinite(func(i int) int { return i * i }).Take(5).Collect()
//	// Returns: [0, 1, 4, 9, 16]
func Infinite[T any](generator func(index int) T) Flow[T] {
	return Flow[T]{
		source: func(yield func(T) bool) {
			i := 0
			for {
				if !yield(generator(i)) {
					return
				}
				i++
			}
		},
	}
}

// FromChannel creates a Flow from a channel.
// The Flow will consume values from the channel until it's closed.
//
// Example:
//
//	ch := make(chan int)
//	go func() {
//	    for i := 0; i < 5; i++ {
//	        ch <- i
//	    }
//	    close(ch)
//	}()
//	flow.FromChannel(ch).ForEach(fmt.Println)
func FromChannel[T any](ch <-chan T) Flow[T] {
	return Flow[T]{
		source: func(yield func(T) bool) {
			for val := range ch {
				if !yield(val) {
					return
				}
			}
		},
	}
}


// Filter returns a Flow containing only elements that match the predicate.
// This is a lazy operation - the predicate is not called until the stream is consumed.
//
// Example:
//
//	flow.Range(1, 10).Filter(func(x int) bool { return x%2 == 0 })
func (f Flow[T]) Filter(predicate func(T) bool) Flow[T] {
	return Flow[T]{
		source: func(yield func(T) bool) {
			for val := range f.source {
				if predicate(val) {
					if !yield(val) {
						return
					}
				}
			}
		},
	}
}

// Map transforms each element using the provided mapper function.
// This is a lazy operation - the mapper is not called until the stream is consumed.
//
// Example:
//
//	flow.Range(1, 6).Map(func(x int) int { return x * x })
func (f Flow[T]) Map(mapper func(T) T) Flow[T] {
	return Flow[T]{
		source: func(yield func(T) bool) {
			for val := range f.source {
				if !yield(mapper(val)) {
					return
				}
			}
		},
	}
}

// Take limits the stream to the first n elements.
// If the stream has fewer than n elements, all elements are included.
//
// Example:
//
//	flow.Infinite(func(i int) int { return i }).Take(5)
func (f Flow[T]) Take(n int) Flow[T] {
	if n <= 0 {
		return Empty[T]()
	}
	return Flow[T]{
		source: func(yield func(T) bool) {
			if n == 0 {
				return
			}
			count := 0
			for val := range f.source {
				if !yield(val) {
					return
				}
				count++
				if count >= n {
					return
				}
			}
		},
	}
}

// Skip discards the first n elements from the stream.
// If the stream has fewer than n elements, an empty stream is returned.
//
// Example:
//
//	flow.Range(1, 10).Skip(5) // Stream of 6, 7, 8, 9
func (f Flow[T]) Skip(n int) Flow[T] {
	return Flow[T]{
		source: func(yield func(T) bool) {
			count := 0
			for val := range f.source {
				if count < n {
					count++
					continue
				}
				if !yield(val) {
					return
				}
			}
		},
	}
}

// TakeWhile takes elements while the predicate is true.
// This is a lazy operation - stops when predicate returns false.
//
// Example:
//
//	flow.Range(1, 10).TakeWhile(func(x int) bool { return x < 5 })
func (f Flow[T]) TakeWhile(predicate func(T) bool) Flow[T] {
	return Flow[T]{
		source: func(yield func(T) bool) {
			for val := range f.source {
				if !predicate(val) {
					return
				}
				if !yield(val) {
					return
				}
			}
		},
	}
}

// SkipWhile skips elements while the predicate is true.
// This is a lazy operation - starts yielding when predicate returns false.
//
// Example:
//
//	flow.Range(1, 10).SkipWhile(func(x int) bool { return x < 5 })
func (f Flow[T]) SkipWhile(predicate func(T) bool) Flow[T] {
	return Flow[T]{
		source: func(yield func(T) bool) {
			skipping := true
			for val := range f.source {
				if skipping && predicate(val) {
					continue
				}
				skipping = false
				if !yield(val) {
					return
				}
			}
		},
	}
}

// Concat appends another Flow to this one.
// This is a lazy operation - the second flow is not consumed until needed.
//
// Example:
//
//	first := flow.NewFlow([]int{1, 2, 3})
//	second := flow.NewFlow([]int{4, 5, 6})
//	combined := first.Concat(second) // [1, 2, 3, 4, 5, 6]
func (f Flow[T]) Concat(other Flow[T]) Flow[T] {
	return Flow[T]{
		source: func(yield func(T) bool) {
			for val := range f.source {
				if !yield(val) {
					return
				}
			}
			for val := range other.source {
				if !yield(val) {
					return
				}
			}
		},
	}
}

// Merge combines this flow with others into a single flow.
// Similar to Concat, but can merge multiple flows at once.
// This is a lazy operation - flows are consumed sequentially.
//
// Example:
//
//	flow1 := flow.NewFlow([]int{1, 2, 3})
//	flow2 := flow.NewFlow([]int{4, 5, 6})
//	flow3 := flow.NewFlow([]int{7, 8, 9})
//	merged := flow1.Merge(flow2, flow3) // [1, 2, 3, 4, 5, 6, 7, 8, 9]
//
//	// Can also be used without arguments (returns the same flow)
//	same := flow1.Merge() // [1, 2, 3]
func (f Flow[T]) Merge(others ...Flow[T]) Flow[T] {
	if len(others) == 0 {
		return f
	}

	return Flow[T]{
		source: func(yield func(T) bool) {
			for val := range f.source {
				if !yield(val) {
					return
				}
			}
			for _, other := range others {
				for val := range other.source {
					if !yield(val) {
						return
					}
				}
			}
		},
	}
}

// Peek performs an action on each element without consuming the stream.
// Useful for debugging or side effects like logging.
// The action is called lazily as elements are consumed.
//
// Example:
//
//	flow.Range(1, 6).
//	    Peek(func(x int) { fmt.Printf("Processing: %d\n", x) }).
//	    Filter(func(x int) bool { return x%2 == 0 }).
//	    Collect()
func (f Flow[T]) Peek(action func(T)) Flow[T] {
	return Flow[T]{
		source: func(yield func(T) bool) {
			for val := range f.source {
				action(val)
				if !yield(val) {
					return
				}
			}
		},
	}
}


// ForEach executes the given function for each element in the stream.
// This is a TERMINAL operation - it consumes the stream immediately.
// Accepts ANY function through reflection for maximum flexibility.
// For better performance with known function types, use ForEachFunc.
//
// Example:
//
//	flow.NewFlow([]int{1, 2, 3}).ForEach(fmt.Print)        // Works with fmt.Print!
//	flow.NewFlow([]int{1, 2, 3}).ForEach(fmt.Println)      // Works with fmt.Println!
//	flow.NewFlow([]int{1, 2, 3}).ForEach(customFunc)       // Works with any function!
func (f Flow[T]) ForEach(fn any) {
	if err := internal.ExecuteForEach(f.source, fn); err != nil {
		panic(err)
	}
}

// ForEachFunc is a type-safe, optimized version of ForEach.
// Use this for better performance when the function type is known at compile time.
// This version doesn't use reflection and is significantly faster.
//
// Example:
//
//	flow.Range(1, 6).ForEachFunc(func(x int) {
//	    fmt.Println(x * x)
//	})
func (f Flow[T]) ForEachFunc(action func(T)) {
	for val := range f.source {
		action(val)
	}
}

// Collect gathers all elements into a slice.
// This is a TERMINAL operation - it consumes the entire stream.
//
// Example:
//
//	numbers := flow.Range(1, 6).Collect() // Returns []int{1, 2, 3, 4, 5}
func (f Flow[T]) Collect() []T {
	result := make([]T, 0, 16)
	for val := range f.source {
		result = append(result, val)
	}
	return result
}

// Count returns the number of elements in the stream.
// This is a TERMINAL operation - it consumes the entire stream.
//
// Example:
//
//	count := flow.Range(1, 100).Filter(func(x int) bool { return x%7 == 0 }).Count()
func (f Flow[T]) Count() int {
	count := 0
	for range f.source {
		count++
	}
	return count
}

// Reduce combines all elements using the reducer function.
// This is a TERMINAL operation - it consumes the entire stream.
// The initial value is used as the starting accumulator.
//
// Example:
//
//	sum := flow.Range(1, 6).Reduce(0, func(acc, x int) int { return acc + x })
//	product := flow.Range(1, 6).Reduce(1, func(acc, x int) int { return acc * x })
func (f Flow[T]) Reduce(initial T, reducer func(accumulator, element T) T) T {
	result := initial
	for val := range f.source {
		result = reducer(result, val)
	}
	return result
}

// First returns the first element if it exists.
// This is a TERMINAL operation - it may consume only one element.
//
// Example:
//
//	if val, ok := flow.Range(10, 20).First(); ok {
//	    fmt.Printf("First: %d\n", val)
//	}
func (f Flow[T]) First() (T, bool) {
	for val := range f.source {
		return val, true
	}
	var zero T
	return zero, false
}

// Last returns the last element if it exists.
// This is a TERMINAL operation - it consumes the entire stream.
//
// Example:
//
//	if val, ok := flow.Range(10, 20).Last(); ok {
//	    fmt.Printf("Last: %d\n", val)
//	}
func (f Flow[T]) Last() (T, bool) {
	var last T
	found := false
	for val := range f.source {
		last = val
		found = true
	}
	return last, found
}

// AnyMatch checks if any element matches the predicate.
// This is a TERMINAL operation - it stops at the first match.
//
// Example:
//
//	hasEven := flow.NewFlow([]int{1, 3, 5, 6}).AnyMatch(func(x int) bool { return x%2 == 0 })
func (f Flow[T]) AnyMatch(predicate func(T) bool) bool {
	for val := range f.source {
		if predicate(val) {
			return true
		}
	}
	return false
}

// AllMatch checks if all elements match the predicate.
// This is a TERMINAL operation - it stops at the first non-match.
//
// Example:
//
//	allPositive := flow.NewFlow([]int{1, 2, 3}).AllMatch(func(x int) bool { return x > 0 })
func (f Flow[T]) AllMatch(predicate func(T) bool) bool {
	for val := range f.source {
		if !predicate(val) {
			return false
		}
	}
	return true
}

// NoneMatch checks if no elements match the predicate.
// This is a TERMINAL operation - it stops at the first match.
//
// Example:
//
//	noneNegative := flow.NewFlow([]int{1, 2, 3}).NoneMatch(func(x int) bool { return x < 0 })
func (f Flow[T]) NoneMatch(predicate func(T) bool) bool {
	return !f.AnyMatch(predicate)
}

// FindFirst returns the first element matching the predicate.
// This is a TERMINAL operation - it stops at the first match.
//
// Example:
//
//	if val, ok := flow.Range(1, 20).FindFirst(func(x int) bool { return x > 10 }); ok {
//	    fmt.Printf("Found: %d\n", val)
//	}
func (f Flow[T]) FindFirst(predicate func(T) bool) (T, bool) {
	for val := range f.source {
		if predicate(val) {
			return val, true
		}
	}
	var zero T
	return zero, false
}

// ToChannel sends all elements to a new channel.
// The channel is created with the specified buffer size.
// The channel is closed after all elements are sent.
// This is a TERMINAL operation that runs in a goroutine.
//
// Example:
//
//	ch := flow.Range(1, 6).ToChannel(2)
//	for val := range ch {
//	    fmt.Println(val)
//	}
func (f Flow[T]) ToChannel(bufferSize int) <-chan T {
	ch := make(chan T, bufferSize)
	go func() {
		defer close(ch)
		for val := range f.source {
			ch <- val
		}
	}()
	return ch
}
