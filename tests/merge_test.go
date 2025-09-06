package flow_test

import (
	"testing"

	. "github.com/MirrexOne/Flow"
)

func TestMergeMethod(t *testing.T) {
	t.Run("Merge with no arguments", func(t *testing.T) {
		flow1 := NewFlow([]int{1, 2, 3})
		result := flow1.Merge().Collect()

		expected := []int{1, 2, 3}
		if len(result) != len(expected) {
			t.Errorf("Expected %v, got %v", expected, result)
		}
		for i, v := range result {
			if v != expected[i] {
				t.Errorf("At index %d: expected %d, got %d", i, expected[i], v)
			}
		}
	})

	t.Run("Merge with one flow", func(t *testing.T) {
		flow1 := NewFlow([]int{1, 2, 3})
		flow2 := NewFlow([]int{4, 5, 6})
		result := flow1.Merge(flow2).Collect()

		expected := []int{1, 2, 3, 4, 5, 6}
		if len(result) != len(expected) {
			t.Errorf("Expected %v, got %v", expected, result)
		}
		for i, v := range result {
			if v != expected[i] {
				t.Errorf("At index %d: expected %d, got %d", i, expected[i], v)
			}
		}
	})

	t.Run("Merge with multiple flows", func(t *testing.T) {
		flow1 := NewFlow([]int{1, 2, 3})
		flow2 := NewFlow([]int{4, 5, 6})
		flow3 := NewFlow([]int{7, 8, 9})
		flow4 := NewFlow([]int{10})

		result := flow1.Merge(flow2, flow3, flow4).Collect()

		expected := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
		if len(result) != len(expected) {
			t.Errorf("Expected %v, got %v", expected, result)
		}
		for i, v := range result {
			if v != expected[i] {
				t.Errorf("At index %d: expected %d, got %d", i, expected[i], v)
			}
		}
	})

	t.Run("Merge with empty flows", func(t *testing.T) {
		flow1 := NewFlow([]int{1, 2, 3})
		empty := Empty[int]()
		flow3 := NewFlow([]int{4, 5})

		result := flow1.Merge(empty, flow3).Collect()

		expected := []int{1, 2, 3, 4, 5}
		if len(result) != len(expected) {
			t.Errorf("Expected %v, got %v", expected, result)
		}
		for i, v := range result {
			if v != expected[i] {
				t.Errorf("At index %d: expected %d, got %d", i, expected[i], v)
			}
		}
	})

	t.Run("Chain Merge operations", func(t *testing.T) {
		flow1 := NewFlow([]int{1, 2})
		flow2 := NewFlow([]int{3, 4})
		flow3 := NewFlow([]int{5, 6})

		result := flow1.Merge(flow2).Merge(flow3).Collect()

		expected := []int{1, 2, 3, 4, 5, 6}
		if len(result) != len(expected) {
			t.Errorf("Expected %v, got %v", expected, result)
		}
		for i, v := range result {
			if v != expected[i] {
				t.Errorf("At index %d: expected %d, got %d", i, expected[i], v)
			}
		}
	})

	t.Run("Merge with transformations", func(t *testing.T) {
		flow1 := Range(1, 4) // 1, 2, 3
		flow2 := Range(4, 7) // 4, 5, 6

		result := CollectAny(flow1.
			Merge(flow2).
			Filter(func(x int) bool { return x%2 == 0 }).
			Map(func(x int) int { return x * 2 }))

		// Original: 1, 2, 3, 4, 5, 6
		// After filter (even only): 2, 4, 6
		// After map (*2): 4, 8, 12
		expected := []any{4, 8, 12}
		if len(result) != len(expected) {
			t.Errorf("Expected %v, got %v", expected, result)
		}
		for i, v := range result {
			if v != expected[i] {
				t.Errorf("At index %d: expected %d, got %d", i, expected[i], v)
			}
		}
	})
}

func TestMergeFunction(t *testing.T) {
	t.Run("Merge function with no arguments", func(t *testing.T) {
		result := Merge[int, int]().Collect()

		if len(result) != 0 {
			t.Errorf("Expected empty result, got %v", result)
		}
	})

	t.Run("Merge function with single flow", func(t *testing.T) {
		flow1 := NewFlow([]int{1, 2, 3})
		result := Merge(flow1).Collect()

		expected := []int{1, 2, 3}
		if len(result) != len(expected) {
			t.Errorf("Expected %v, got %v", expected, result)
		}
		for i, v := range result {
			if v != expected[i] {
				t.Errorf("At index %d: expected %d, got %d", i, expected[i], v)
			}
		}
	})

	t.Run("Merge function with multiple flows", func(t *testing.T) {
		flow1 := NewFlow([]string{"a", "b"})
		flow2 := NewFlow([]string{"c", "d"})
		flow3 := NewFlow([]string{"e"})

		result := Merge(flow1, flow2, flow3).Collect()

		expected := []string{"a", "b", "c", "d", "e"}
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
