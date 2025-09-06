package benchmarks_test

import (
	flow "github.com/MirrexOne/Flow"
	"testing"
)

// Benchmark comparing Flow with traditional for loop
func BenchmarkFlowVsLoop(b *testing.B) {
	data := make([]int, 1000)
	for i := range data {
		data[i] = i
	}

	b.Run("Traditional Loop", func(b *testing.B) {
		for b.Loop() {
			sum := 0
			for _, v := range data {
				if v%2 == 0 {
					sum += v * v
				}
			}
			_ = sum
		}
	})

	b.Run("Flow API", func(b *testing.B) {
		for b.Loop() {
			sum := flow.From(data).
				Filter(func(x int) bool { return x%2 == 0 }).
				Map(func(x int) int { return x * x }).
				Reduce(0, func(acc, x int) int { return acc + x })
			_ = sum
		}
	})

	b.Run("Flow Lazy Evaluation", func(b *testing.B) {
		for b.Loop() {
			// Only take first 10 even numbers - lazy evaluation benefit
			result := flow.From(data).
				Filter(func(x int) bool { return x%2 == 0 }).
				Map(func(x int) int { return x * x }).
				Take(10).
				Collect()
			_ = result
		}
	})
}

// Benchmark for different Flow operations
func BenchmarkFlowOperations(b *testing.B) {
	data := make([]int, 10000)
	for i := range data {
		data[i] = i
	}

	b.Run("Filter", func(b *testing.B) {
		for b.Loop() {
			flow.From(data).
				Filter(func(x int) bool { return x%2 == 0 }).
				Count()
		}
	})

	b.Run("Map", func(b *testing.B) {
		for b.Loop() {
			flow.From(data).
				Map(func(x int) int { return x * 2 }).
				Count()
		}
	})

	b.Run("Take", func(b *testing.B) {
		for b.Loop() {
			flow.From(data).
				Take(100).
				Collect()
		}
	})

	b.Run("Distinct", func(b *testing.B) {
		smallData := make([]int, 100)
		for i := range smallData {
			smallData[i] = i % 10 // Many duplicates
		}
		b.ResetTimer()

		for b.Loop() {
			flow.Distinct(flow.From(smallData)).Count()
		}
	})
}

// Benchmark lazy vs eager evaluation
func BenchmarkLazyVsEager(b *testing.B) {
	b.Run("Lazy - Take 10 from infinite", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			flow.Infinite(func(x int) int { return x * x }).
				Take(10).
				Collect()
		}
	})

	b.Run("Eager - Generate 1000 then take 10", func(b *testing.B) {
		for b.Loop() {
			data := make([]int, 1000)
			for j := range data {
				data[j] = j * j
			}
			result := data[:10]
			_ = result
		}
	})
}

// Benchmark for Combine operations
func BenchmarkCombineOperations(b *testing.B) {
	data1 := make([]int, 1000)
	data2 := make([]int, 1000)
	for i := range data1 {
		data1[i] = i
		data2[i] = i * 2
	}

	b.Run("Combine", func(b *testing.B) {
		for b.Loop() {
			flow.Combine(flow.From(data1), flow.From(data2)).Count()
		}
	})

	b.Run("CombineWith", func(b *testing.B) {
		for b.Loop() {
			flow.CombineWith(flow.From(data1), flow.From(data2), func(a, b int) int {
				return a + b
			}).Count()
		}
	})

	b.Run("Merge", func(b *testing.B) {
		for b.Loop() {
			flow.Merge(
				flow.From(data1[:100]),
				flow.From(data2[:100]),
			).Count()
		}
	})
}

// Benchmark for GroupBy operations
func BenchmarkGroupByOperations(b *testing.B) {
	data := make([]int, 1000)
	for i := range data {
		data[i] = i
	}

	b.Run("GroupBy mod 10", func(b *testing.B) {
		for b.Loop() {
			groups := flow.GroupBy(flow.From(data), func(x int) int {
				return x % 10
			})
			_ = groups
		}
	})

	b.Run("GroupBy with counting", func(b *testing.B) {
		for b.Loop() {
			groups := flow.GroupBy(flow.From(data), func(x int) int {
				return x % 10
			})
			for _, group := range groups {
				_ = len(group)
			}
		}
	})
}

// Benchmark for Chunk operations
func BenchmarkChunkOperations(b *testing.B) {
	data := make([]int, 10000)
	for i := range data {
		data[i] = i
	}

	b.Run("Chunk size 10", func(b *testing.B) {
		for b.Loop() {
			flow.Chunk(flow.From(data), 10).Count()
		}
	})

	b.Run("Chunk size 100", func(b *testing.B) {
		for b.Loop() {
			flow.Chunk(flow.From(data), 100).Count()
		}
	})

	b.Run("Chunk size 1000", func(b *testing.B) {
		for b.Loop() {
			flow.Chunk(flow.From(data), 1000).Count()
		}
	})
}
