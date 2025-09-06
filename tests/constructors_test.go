package flow_test

import (
	. "github.com/MirrexOne/Flow"
	"testing"
)

func TestConstructors(t *testing.T) {
	t.Run("NewFlow from slice", func(t *testing.T) {
		slice := []int{1, 2, 3, 4, 5}
		result := NewFlow(slice).Collect()
		if len(result) != 5 {
			t.Errorf("Expected 5 elements, got %d", len(result))
		}
	})

	t.Run("Of with variadic arguments", func(t *testing.T) {
		result := Of(1, 2, 3, 4, 5).Collect()
		if len(result) != 5 {
			t.Errorf("Expected 5 elements, got %d", len(result))
		}

		// Test with strings
		strings := Of("hello", "world").Collect()
		if len(strings) != 2 {
			t.Errorf("Expected 2 strings, got %d", len(strings))
		}
	})

	t.Run("Values with variadic arguments", func(t *testing.T) {
		result := Values(10, 20, 30).Collect()
		if len(result) != 3 {
			t.Errorf("Expected 3 elements, got %d", len(result))
		}
		if result[0] != 10 || result[1] != 20 || result[2] != 30 {
			t.Errorf("Unexpected values: %v", result)
		}
	})

	t.Run("FromSlice", func(t *testing.T) {
		slice := []string{"a", "b", "c"}
		result := FromSlice(slice).Collect()
		if len(result) != 3 {
			t.Errorf("Expected 3 elements, got %d", len(result))
		}
		if result[0] != "a" || result[1] != "b" || result[2] != "c" {
			t.Errorf("Unexpected values: %v", result)
		}
	})

	t.Run("Single value", func(t *testing.T) {
		result := Single(42).Collect()
		if len(result) != 1 || result[0] != 42 {
			t.Errorf("Expected [42], got %v", result)
		}
	})

	t.Run("Empty flow", func(t *testing.T) {
		result := Empty[int]().Collect()
		if len(result) != 0 {
			t.Errorf("Expected empty slice, got %v", result)
		}
	})

	t.Run("Range", func(t *testing.T) {
		result := Range(1, 6).Collect()
		expected := []int{1, 2, 3, 4, 5}
		if len(result) != len(expected) {
			t.Errorf("Expected %d elements, got %d", len(expected), len(result))
		}
		for i, v := range result {
			if v != expected[i] {
				t.Errorf("At index %d: expected %d, got %d", i, expected[i], v)
			}
		}
	})

	t.Run("From (backward compatibility)", func(t *testing.T) {
		// Should still work for backward compatibility
		result := From([]int{1, 2, 3}).Collect()
		if len(result) != 3 {
			t.Errorf("Expected 3 elements, got %d", len(result))
		}
	})
}

func TestVariadicConstructorsWithDifferentTypes(t *testing.T) {
	t.Run("Of with structs", func(t *testing.T) {
		type Person struct {
			Name string
			Age  int
		}

		people := Of(
			Person{"Alice", 25},
			Person{"Bob", 30},
			Person{"Charlie", 35},
		)

		count := people.Count()
		if count != 3 {
			t.Errorf("Expected 3 people, got %d", count)
		}
	})

	t.Run("Values with pointers", func(t *testing.T) {
		a, b, c := 1, 2, 3
		result := Values(&a, &b, &c).Collect()
		if len(result) != 3 {
			t.Errorf("Expected 3 pointers, got %d", len(result))
		}
		if *result[0] != 1 || *result[1] != 2 || *result[2] != 3 {
			t.Errorf("Unexpected pointer values")
		}
	})

	t.Run("Zero variadic arguments", func(t *testing.T) {
		// Of with no arguments should create empty flow
		result := Of[int]().Collect()
		if len(result) != 0 {
			t.Errorf("Expected empty flow, got %v", result)
		}

		// Values with no arguments
		result2 := Values[string]().Collect()
		if len(result2) != 0 {
			t.Errorf("Expected empty flow, got %v", result2)
		}
	})
}
