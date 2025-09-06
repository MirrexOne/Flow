package flow_test

import (
	"testing"

	. "github.com/MirrexOne/Flow"
)

func TestDistinct(t *testing.T) {
	t.Run("Remove duplicates", func(t *testing.T) {
		data := NewFlow([]int{1, 2, 2, 3, 3, 3, 4, 4, 4, 4})
		result := Distinct(data).Collect()

		expected := []int{1, 2, 3, 4}
		if len(result) != len(expected) {
			t.Errorf("Expected %v, got %v", expected, result)
		}
		for i, v := range result {
			if v != expected[i] {
				t.Errorf("At index %d: expected %d, got %d", i, expected[i], v)
			}
		}
	})

	t.Run("Empty flow", func(t *testing.T) {
		result := Distinct(Empty[int]()).Collect()
		if len(result) != 0 {
			t.Errorf("Expected empty result, got %v", result)
		}
	})

	t.Run("No duplicates", func(t *testing.T) {
		data := NewFlow([]int{1, 2, 3, 4, 5})
		result := Distinct(data).Collect()

		expected := []int{1, 2, 3, 4, 5}
		if len(result) != len(expected) {
			t.Errorf("Expected %v, got %v", expected, result)
		}
	})
}

func TestMapTo(t *testing.T) {
	t.Run("Int to string", func(t *testing.T) {
		data := Range(1, 4)
		result := MapTo(data, func(x int) string {
			return string('a' + rune(x-1))
		}).Collect()

		expected := []string{"a", "b", "c"}
		if len(result) != len(expected) {
			t.Errorf("Expected %v, got %v", expected, result)
		}
		for i, v := range result {
			if v != expected[i] {
				t.Errorf("At index %d: expected %s, got %s", i, expected[i], v)
			}
		}
	})
}

func TestChunk(t *testing.T) {
	t.Run("Even chunks", func(t *testing.T) {
		data := Range(1, 7)
		result := Chunk(data, 2).Collect()

		expected := [][]int{{1, 2}, {3, 4}, {5, 6}}
		if len(result) != len(expected) {
			t.Errorf("Expected %d chunks, got %d", len(expected), len(result))
		}
		for i, chunk := range result {
			if len(chunk) != len(expected[i]) {
				t.Errorf("Chunk %d: expected length %d, got %d", i, len(expected[i]), len(chunk))
			}
			for j, v := range chunk {
				if v != expected[i][j] {
					t.Errorf("Chunk %d, index %d: expected %d, got %d", i, j, expected[i][j], v)
				}
			}
		}
	})

	t.Run("Uneven chunks", func(t *testing.T) {
		data := Range(1, 8)
		result := Chunk(data, 3).Collect()

		expected := [][]int{{1, 2, 3}, {4, 5, 6}, {7}}
		if len(result) != len(expected) {
			t.Errorf("Expected %d chunks, got %d", len(expected), len(result))
		}
		if len(result[2]) != 1 {
			t.Errorf("Last chunk should have 1 element, got %d", len(result[2]))
		}
	})
}

func TestGroupBy(t *testing.T) {
	t.Run("Group by modulo", func(t *testing.T) {
		data := Range(1, 11)
		groups := GroupBy(data, func(x int) int { return x % 3 })

		if len(groups) != 3 {
			t.Errorf("Expected 3 groups, got %d", len(groups))
		}

		if len(groups[0]) != 3 {
			t.Errorf("Group 0: expected 3 elements, got %d", len(groups[0]))
		}

		if len(groups[1]) != 4 {
			t.Errorf("Group 1: expected 4 elements, got %d", len(groups[1]))
		}

		if len(groups[2]) != 3 {
			t.Errorf("Group 2: expected 3 elements, got %d", len(groups[2]))
		}
	})
}

func TestPartition(t *testing.T) {
	t.Run("Partition even/odd", func(t *testing.T) {
		data := Range(1, 11)
		evens, odds := Partition(data, func(x int) bool { return x%2 == 0 })

		expectedEvens := []int{2, 4, 6, 8, 10}
		expectedOdds := []int{1, 3, 5, 7, 9}

		if len(evens) != len(expectedEvens) {
			t.Errorf("Evens: expected %d elements, got %d", len(expectedEvens), len(evens))
		}
		if len(odds) != len(expectedOdds) {
			t.Errorf("Odds: expected %d elements, got %d", len(expectedOdds), len(odds))
		}

		for i, v := range evens {
			if v != expectedEvens[i] {
				t.Errorf("Evens[%d]: expected %d, got %d", i, expectedEvens[i], v)
			}
		}
		for i, v := range odds {
			if v != expectedOdds[i] {
				t.Errorf("Odds[%d]: expected %d, got %d", i, expectedOdds[i], v)
			}
		}
	})
}

func TestWindow(t *testing.T) {
	t.Run("Sliding window", func(t *testing.T) {
		data := Range(1, 6)
		result := Window(data, 3, 1).Collect()

		expected := [][]int{{1, 2, 3}, {2, 3, 4}, {3, 4, 5}}
		if len(result) != len(expected) {
			t.Errorf("Expected %d windows, got %d", len(expected), len(result))
		}

		for i, window := range result {
			if len(window) != 3 {
				t.Errorf("Window %d: expected size 3, got %d", i, len(window))
			}
			for j, v := range window {
				if v != expected[i][j] {
					t.Errorf("Window %d, index %d: expected %d, got %d", i, j, expected[i][j], v)
				}
			}
		}
	})

	t.Run("Tumbling window", func(t *testing.T) {
		data := Range(1, 10)
		result := Window(data, 3, 3).Collect()

		expected := [][]int{{1, 2, 3}, {4, 5, 6}, {7, 8, 9}}
		if len(result) != len(expected) {
			t.Errorf("Expected %d windows, got %d", len(expected), len(result))
		}

		for i, window := range result {
			for j, v := range window {
				if v != expected[i][j] {
					t.Errorf("Window %d, index %d: expected %d, got %d", i, j, expected[i][j], v)
				}
			}
		}
	})
}

func TestFlatMap(t *testing.T) {
	t.Run("Flatten words to characters", func(t *testing.T) {
		words := NewFlow([]string{"ab", "cd"})
		result := FlatMap(words, func(word string) Flow[rune] {
			return NewFlow([]rune(word))
		}).Collect()

		expected := []rune{'a', 'b', 'c', 'd'}
		if len(result) != len(expected) {
			t.Errorf("Expected %d characters, got %d", len(expected), len(result))
		}
		for i, v := range result {
			if v != expected[i] {
				t.Errorf("At index %d: expected %c, got %c", i, expected[i], v)
			}
		}
	})

	t.Run("Flatten ranges", func(t *testing.T) {
		ranges := NewFlow([]int{2, 3, 2})
		result := FlatMap(ranges, func(n int) Flow[int] {
			return Range(1, n+1)
		}).Collect()

		expected := []int{1, 2, 1, 2, 3, 1, 2}
		if len(result) != len(expected) {
			t.Errorf("Expected %v, got %v", expected, result)
		}
	})
}
