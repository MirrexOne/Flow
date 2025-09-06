package benchmarks_test

import (
	"testing"

	flow "github.com/MirrexOne/Flow"
)

// Benchmark simple operations to showcase minimal overhead
func BenchmarkSimpleOperations(b *testing.B) {
	data := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	b.Run("Traditional Loop Sum", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			sum := 0
			for _, v := range data {
				sum += v
			}
			_ = sum
		}
	})

	b.Run("Flow Reduce Sum", func(b *testing.B) {
		b.ReportAllocs()
		f := flow.NewFlow(data)
		b.ResetTimer()
		for b.Loop() {
			sum := f.Reduce(0, func(acc, x int) int { return acc + x })
			_ = sum
		}
	})

	b.Run("Flow Take First 3", func(b *testing.B) {
		b.ReportAllocs()
		f := flow.NewFlow(data)
		b.ResetTimer()
		for b.Loop() {
			result := f.Take(3).Collect()
			_ = result
		}
	})
}

// Benchmark comparing Flow with traditional for loop
func BenchmarkFlowVsLoop(b *testing.B) {
	// Smaller dataset for faster benchmarks
	data := make([]int, 100)
	for i := range data {
		data[i] = i
	}

	b.Run("Traditional Loop", func(b *testing.B) {
		b.ReportAllocs()
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
		b.ReportAllocs()
		f := flow.NewFlow(data)
		b.ResetTimer()
		for b.Loop() {
			sum := flow.ReduceAny(f.
				Filter(func(x int) bool { return x%2 == 0 }).
				Map(func(x int) int { return x * x }),
				0, func(acc, x any) any { return acc.(int) + x.(int) })
			_ = sum
		}
	})

	b.Run("Flow Lazy Evaluation", func(b *testing.B) {
		b.ReportAllocs()
		f := flow.NewFlow(data)
		b.ResetTimer()
		for b.Loop() {
			// Only take first 5 even numbers - lazy evaluation benefit
			result := f.
				Filter(func(x int) bool { return x%2 == 0 }).
				Take(5).
				Collect()
			_ = result
		}
	})
}

// Benchmark for different Flow operations
func BenchmarkFlowOperations(b *testing.B) {
	data := make([]int, 1000)
	for i := range data {
		data[i] = i
	}

	b.Run("Filter", func(b *testing.B) {
		b.ReportAllocs()
		f := flow.NewFlow(data)
		b.ResetTimer()
		for b.Loop() {
			f.Filter(func(x int) bool { return x%2 == 0 }).Count()
		}
	})

	b.Run("Map", func(b *testing.B) {
		b.ReportAllocs()
		f := flow.NewFlow(data)
		b.ResetTimer()
		for b.Loop() {
			f.Map(func(x int) int { return x * 2 }).Count()
		}
	})

	b.Run("Take", func(b *testing.B) {
		b.ReportAllocs()
		f := flow.NewFlow(data)
		b.ResetTimer()
		for b.Loop() {
			f.Take(10).Collect()
		}
	})

	b.Run("Distinct", func(b *testing.B) {
		smallData := make([]int, 100)
		for i := range smallData {
			smallData[i] = i % 10 // Many duplicates
		}
		b.ResetTimer()

		for b.Loop() {
			flow.Distinct(flow.NewFlow(smallData)).Count()
		}
	})
}

// Benchmark lazy vs eager evaluation
func BenchmarkLazyVsEager(b *testing.B) {
	b.Run("Lazy - Take 5 from infinite", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			flow.Infinite(func(x int) int { return x }).
				Take(5).
				Collect()
		}
	})

	b.Run("Eager - Generate 100 then take 5", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			data := make([]int, 100)
			for j := range data {
				data[j] = j
			}
			result := data[:5]
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
			flow.Combine(flow.NewFlow(data1), flow.NewFlow(data2)).Count()
		}
	})

	b.Run("CombineWith", func(b *testing.B) {
		for b.Loop() {
			flow.CombineWith(flow.NewFlow(data1), flow.NewFlow(data2), func(a, b int) int {
				return a + b
			}).Count()
		}
	})

	b.Run("Merge", func(b *testing.B) {
		for b.Loop() {
			flow.Merge(
				flow.NewFlow(data1[:100]),
				flow.NewFlow(data2[:100]),
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
			groups := flow.GroupBy(flow.NewFlow(data), func(x int) int {
				return x % 10
			})
			_ = groups
		}
	})

	b.Run("GroupBy with counting", func(b *testing.B) {
		for b.Loop() {
			groups := flow.GroupBy(flow.NewFlow(data), func(x int) int {
				return x % 100
			})
			for _, group := range groups {
				_ = len(group)
			}
		}
	})
}

// Benchmark for Chunk operations
func BenchmarkChunkOperations(b *testing.B) {
	data := make([]int, 1000)
	for i := range data {
		data[i] = i
	}

	b.Run("Chunk size 10", func(b *testing.B) {
		b.ReportAllocs()
		f := flow.NewFlow(data)
		b.ResetTimer()
		for b.Loop() {
			flow.Chunk(f, 10).Count()
		}
	})

	b.Run("Chunk size 100", func(b *testing.B) {
		b.ReportAllocs()
		f := flow.NewFlow(data)
		b.ResetTimer()
		for b.Loop() {
			flow.Chunk(f, 100).Count()
		}
	})

	b.Run("Chunk size 1000", func(b *testing.B) {
		b.ReportAllocs()
		f := flow.NewFlow(data)
		b.ResetTimer()
		for b.Loop() {
			flow.Chunk(f, 1000).Count()
		}
	})
}
