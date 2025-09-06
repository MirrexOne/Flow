package flow

// MapTo transforms each element to a different type.
// This is a lazy operation - the mapper is not called until the stream is consumed.
// Since Go doesn't support method-level type parameters, this is a standalone function.
//
// Example:
//
//	strings := flow.MapTo(flow.Range(1, 6), func(x int) string {
//	    return fmt.Sprintf("Number: %d", x)
//	})
func MapTo[T, R any](f Flow[T], mapper func(T) R) Flow[R] {
	return Flow[R]{
		source: func(yield func(R) bool) {
			for val := range f.source {
				if !yield(mapper(val)) {
					return
				}
			}
		},
	}
}

// Distinct removes duplicate elements from the stream.
// Requires the type to be comparable.
// This is a lazy operation but requires memory to track seen elements.
//
// Example:
//
//	unique := flow.Distinct(flow.NewFlow([]int{1, 2, 2, 3, 3, 3, 4}))
func Distinct[T comparable](f Flow[T]) Flow[T] {
	return Flow[T]{
		source: func(yield func(T) bool) {
			seen := make(map[T]bool)
			for val := range f.source {
				if !seen[val] {
					seen[val] = true
					if !yield(val) {
						return
					}
				}
			}
		},
	}
}

// FlatMap transforms each element to a Flow and flattens the results.
// Useful for working with nested structures.
//
// Example:
//
//	words := flow.NewFlow([]string{"hello", "world"})
//	letters := flow.FlatMap(words, func(word string) flow.Flow[rune] {
//	    return flow.NewFlow([]rune(word))
//	})
func FlatMap[T, R any](f Flow[T], mapper func(T) Flow[R]) Flow[R] {
	return Flow[R]{
		source: func(yield func(R) bool) {
			for val := range f.source {
				subFlow := mapper(val)
				for subVal := range subFlow.source {
					if !yield(subVal) {
						return
					}
				}
			}
		},
	}
}

// Chunk groups elements into slices of specified size.
// The last chunk may have fewer elements if the stream size is not divisible by the chunk size.
//
// Example:
//
//	chunks := flow.Chunk(flow.Range(1, 11), 3)
//	// Produces: [1,2,3], [4,5,6], [7,8,9], [10]
func Chunk[T any](f Flow[T], size int) Flow[[]T] {
	if size <= 0 {
		panic("chunk size must be positive")
	}

	return Flow[[]T]{
		source: func(yield func([]T) bool) {
			chunk := make([]T, 0, size)
			for val := range f.source {
				chunk = append(chunk, val)
				if len(chunk) == size {
					chunkCopy := make([]T, len(chunk))
					copy(chunkCopy, chunk)
					if !yield(chunkCopy) {
						return
					}
					chunk = chunk[:0]
				}
			}
			if len(chunk) > 0 {
				yield(chunk)
			}
		},
	}
}

// Combine merges two flows into pairs.
// The resulting flow ends when either input flow ends.
//
// Example:
//
//	names := flow.NewFlow([]string{"Alice", "Bob"})
//	ages := flow.NewFlow([]int{25, 30})
//	pairs := flow.Combine(names, ages)
//	// Produces: {First: "Alice", Second: 25}, {First: "Bob", Second: 30}
func Combine[T, U any](f1 Flow[T], f2 Flow[U]) Flow[Pair[T, U]] {
	return Flow[Pair[T, U]]{
		source: func(yield func(Pair[T, U]) bool) {
			var vals1 []T
			var vals2 []U

			for val := range f1.source {
				vals1 = append(vals1, val)
			}
			for val := range f2.source {
				vals2 = append(vals2, val)
			}

			minLen := min(len(vals2), len(vals1))

			for i := range minLen {
				if !yield(Pair[T, U]{First: vals1[i], Second: vals2[i]}) {
					return
				}
			}
		},
	}
}

// Pair represents a pair of values.
// Used by the Combine function.
type Pair[T, U any] struct {
	First  T
	Second U
}

// CombineWith merges two flows using a custom combiner function.
// This provides more flexibility than Combine by allowing custom result types.
// The resulting flow ends when either input flow ends.
//
// Example:
//
//	names := flow.NewFlow([]string{"Alice", "Bob"})
//	ages := flow.NewFlow([]int{25, 30})
//	people := flow.CombineWith(names, ages, func(name string, age int) string {
//	    return fmt.Sprintf("%s is %d years old", name, age)
//	})
//	// Produces: "Alice is 25 years old", "Bob is 30 years old"
func CombineWith[T, U, R any](f1 Flow[T], f2 Flow[U], combiner func(T, U) R) Flow[R] {
	return Flow[R]{
		source: func(yield func(R) bool) {
			var vals1 []T
			var vals2 []U

			for val := range f1.source {
				vals1 = append(vals1, val)
			}
			for val := range f2.source {
				vals2 = append(vals2, val)
			}

			minLen := min(len(vals2), len(vals1))

			for i := range minLen {
				if !yield(combiner(vals1[i], vals2[i])) {
					return
				}
			}
		},
	}
}

// Merge combines multiple flows into a single flow.
// Unlike Combine, this concatenates flows sequentially rather than pairing elements.
// Elements from all flows are yielded in the order they appear.
// Can be called without arguments, in which case it returns an empty flow.
//
// Example:
//
//	flow1 := flow.NewFlow([]int{1, 2, 3})
//	flow2 := flow.NewFlow([]int{4, 5, 6})
//	flow3 := flow.NewFlow([]int{7, 8, 9})
//	merged := flow.Merge(flow1, flow2, flow3)
//	// Produces: 1, 2, 3, 4, 5, 6, 7, 8, 9
//
//	// Can also chain with method:
//	merged2 := flow1.Merge(flow2, flow3)
//	// Produces: 1, 2, 3, 4, 5, 6, 7, 8, 9
func Merge[T any](flows ...Flow[T]) Flow[T] {
	if len(flows) == 0 {
		return Empty[T]()
	}

	return Flow[T]{
		source: func(yield func(T) bool) {
			for _, f := range flows {
				for val := range f.source {
					if !yield(val) {
						return
					}
				}
			}
		},
	}
}

// GroupBy groups elements by a key function.
// Returns a map where keys are the result of the keyFunc and values are slices of elements.
// This is a terminal operation that consumes the entire stream.
//
// Example:
//
//	people := []Person{{Name: "Alice", Age: 25}, {Name: "Bob", Age: 30}, {Name: "Charlie", Age: 25}}
//	byAge := flow.GroupBy(flow.NewFlow(people), func(p Person) int { return p.Age })
//	// Result: map[25:[{Alice 25} {Charlie 25}] 30:[{Bob 30}]]
func GroupBy[T any, K comparable](f Flow[T], keyFunc func(T) K) map[K][]T {
	result := make(map[K][]T)
	for val := range f.source {
		key := keyFunc(val)
		result[key] = append(result[key], val)
	}
	return result
}

// GroupByFlow is a lazy version of GroupBy that returns a Flow of groups.
// Each group is represented as a KeyValue pair containing the key and slice of values.
// This is useful when you want to process groups lazily.
//
// Example:
//
//	people := []Person{{Name: "Alice", Age: 25}, {Name: "Bob", Age: 30}, {Name: "Charlie", Age: 25}}
//	groups := flow.GroupByFlow(flow.NewFlow(people), func(p Person) int { return p.Age })
//	groups.ForEach(func(kv KeyValue[int, []Person]) {
//	    fmt.Printf("Age %d: %v\n", kv.Key, kv.Value)
//	})
func GroupByFlow[T any, K comparable](f Flow[T], keyFunc func(T) K) Flow[KeyValue[K, []T]] {
	return Flow[KeyValue[K, []T]]{
		source: func(yield func(KeyValue[K, []T]) bool) {
			groups := GroupBy(f, keyFunc)
			for key, values := range groups {
				if !yield(KeyValue[K, []T]{Key: key, Value: values}) {
					return
				}
			}
		},
	}
}

// KeyValue represents a key-value pair.
// Used by GroupByFlow and other key-value operations.
type KeyValue[K comparable, V any] struct {
	Key   K
	Value V
}

// Partition splits a flow into two based on a predicate.
// Returns two slices: elements that match the predicate and elements that don't.
// This is a terminal operation that consumes the entire stream.
//
// Example:
//
//	evens, odds := flow.Partition(flow.Range(1, 11), func(x int) bool { return x%2 == 0 })
//	// evens: [2, 4, 6, 8, 10]
//	// odds: [1, 3, 5, 7, 9]
func Partition[T any](f Flow[T], predicate func(T) bool) (matching []T, notMatching []T) {
	for val := range f.source {
		if predicate(val) {
			matching = append(matching, val)
		} else {
			notMatching = append(notMatching, val)
		}
	}
	return
}

// Window creates sliding windows of elements.
// Each window contains 'size' elements, and windows overlap by 'size-step' elements.
// If step equals size, windows don't overlap (tumbling windows).
//
// Example:
//
//	// Sliding window with size=3, step=1
//	windows := flow.Window(flow.Range(1, 6), 3, 1)
//	// Produces: [1,2,3], [2,3,4], [3,4,5]
//
//	// Tumbling window with size=3, step=3
//	windows := flow.Window(flow.Range(1, 10), 3, 3)
//	// Produces: [1,2,3], [4,5,6], [7,8,9]
func Window[T any](f Flow[T], size, step int) Flow[[]T] {
	if size <= 0 {
		panic("window size must be positive")
	}
	if step <= 0 {
		panic("window step must be positive")
	}

	return Flow[[]T]{
		source: func(yield func([]T) bool) {
			var buffer []T
			for val := range f.source {
				buffer = append(buffer, val)

				for len(buffer) >= size {
					window := make([]T, size)
					copy(window, buffer[:size])

					if !yield(window) {
						return
					}

					if step >= len(buffer) {
						buffer = nil
					} else {
						buffer = buffer[step:]
					}
				}
			}
		},
	}
}
