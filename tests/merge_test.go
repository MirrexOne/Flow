package flow_test

import (
	. "github.com/MirrexOne/Flow"
	"testing"
)

func TestMergeMethod(t *testing.T) {
	t.Run("Merge with no arguments", func(t *testing.T) {
		flow1 := From([]int{1, 2, 3})
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
		flow1 := From([]int{1, 2, 3})
		flow2 := From([]int{4, 5, 6})
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
		flow1 := From([]int{1, 2, 3})
		flow2 := From([]int{4, 5, 6})
		flow3 := From([]int{7, 8, 9})
		flow4 := From([]int{10})

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
		flow1 := From([]int{1, 2, 3})
		flow2 := Empty[int]()
		flow3 := From([]int{4, 5})

		result := flow1.Merge(flow2, flow3).Collect()

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
		flow1 := From([]int{1, 2})
		flow2 := From([]int{3, 4})
		flow3 := From([]int{5, 6})

		// Test chaining with method
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

		result := flow1.
			Merge(flow2).
			Filter(func(x int) bool { return x%2 == 0 }).
			Map(func(x int) int { return x * 2 }).
			Collect()

		// Original: 1, 2, 3, 4, 5, 6
		// After filter (even only): 2, 4, 6
		// After map (*2): 4, 8, 12
		expected := []int{4, 8, 12}
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
		result := Merge[int]().Collect()

		if len(result) != 0 {
			t.Errorf("Expected empty result, got %v", result)
		}
	})

	t.Run("Merge function with single flow", func(t *testing.T) {
		flow1 := From([]int{1, 2, 3})
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
		flow1 := From([]string{"a", "b"})
		flow2 := From([]string{"c", "d"})
		flow3 := From([]string{"e"})

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
