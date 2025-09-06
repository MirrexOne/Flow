package flow_test

import (
	. "github.com/MirrexOne/Flow"
	"testing"
)

func TestFlowCorrectness(t *testing.T) {
	t.Run("Filter", func(t *testing.T) {
		result := NewFlow([]int{1, 2, 3, 4, 5}).
			Filter(func(x int) bool { return x%2 == 0 }).
			Collect()

		expected := []int{2, 4}
		if len(result) != len(expected) {
			t.Errorf("Expected %v, got %v", expected, result)
		}
		for i := range result {
			if result[i] != expected[i] {
				t.Errorf("Expected %v, got %v", expected, result)
			}
		}
	})

	t.Run("Map", func(t *testing.T) {
		result := NewFlow([]int{1, 2, 3}).
			Map(func(x int) int { return x * 2 }).
			Collect()

		expected := []int{2, 4, 6}
		if len(result) != len(expected) {
			t.Errorf("Expected %v, got %v", expected, result)
		}
		for i := range result {
			if result[i] != expected[i] {
				t.Errorf("Expected %v, got %v", expected, result)
			}
		}
	})

	t.Run("Reduce", func(t *testing.T) {
		sum := NewFlow([]int{1, 2, 3, 4, 5}).
			Reduce(0, func(acc, x int) int { return acc + x })

		if sum != 15 {
			t.Errorf("Expected 15, got %d", sum)
		}
	})

	t.Run("Lazy Evaluation", func(t *testing.T) {
		result := Infinite(func(i int) int { return i }).
			Filter(func(x int) bool { return x > 5 }).
			Take(3).
			Collect()

		expected := []int{6, 7, 8}
		if len(result) != len(expected) {
			t.Errorf("Expected %v, got %v", expected, result)
		}
		for i := range result {
			if result[i] != expected[i] {
				t.Errorf("Expected %v, got %v", expected, result)
			}
		}
	})
}
