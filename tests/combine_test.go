package flow_test

import (
	"fmt"
	"testing"

	. "github.com/MirrexOne/Flow"
)

func TestCombine(t *testing.T) {
	t.Run("Combine two flows of same length", func(t *testing.T) {
		flow1 := Of(1, 2, 3)
		flow2 := Of("a", "b", "c")

		result := Combine(flow1, flow2).Collect()

		if len(result) != 3 {
			t.Errorf("Expected 3 pairs, got %d", len(result))
		}

		expected := []struct {
			First  int
			Second string
		}{
			{1, "a"},
			{2, "b"},
			{3, "c"},
		}

		for i, pair := range result {
			if pair.First != expected[i].First || pair.Second != expected[i].Second {
				t.Errorf("At index %d: expected %v, got %v", i, expected[i], pair)
			}
		}
	})

	t.Run("Combine flows of different lengths", func(t *testing.T) {
		flow1 := Of(1, 2, 3, 4, 5)
		flow2 := Of("a", "b", "c")

		result := Combine(flow1, flow2).Collect()

		// Should stop at the shorter flow
		if len(result) != 3 {
			t.Errorf("Expected 3 pairs (length of shorter flow), got %d", len(result))
		}
	})

	t.Run("Combine with empty flow", func(t *testing.T) {
		flow1 := Of(1, 2, 3)
		flow2 := Empty[string]()

		result := Combine(flow1, flow2).Collect()

		if len(result) != 0 {
			t.Errorf("Expected empty result when combining with empty flow, got %v", result)
		}
	})

}

func TestCombineWith(t *testing.T) {
	t.Run("CombineWith custom function", func(t *testing.T) {
		flow1 := Of(1, 2, 3)
		flow2 := Of(10, 20, 30)

		result := CombineWith(flow1, flow2, func(a, b int) int {
			return a + b
		}).Collect()

		expected := []int{11, 22, 33}

		if len(result) != len(expected) {
			t.Errorf("Expected %d elements, got %d", len(expected), len(result))
		}

		for i, v := range result {
			if v != expected[i] {
				t.Errorf("At index %d: expected %d, got %d", i, expected[i], v)
			}
		}
	})

	t.Run("CombineWith string concatenation", func(t *testing.T) {
		flow1 := Of("Hello", "Good", "Nice")
		flow2 := Of(" World", " Morning", " Day")

		result := CombineWith(flow1, flow2, func(a, b string) string {
			return a + b
		}).Collect()

		expected := []string{"Hello World", "Good Morning", "Nice Day"}

		for i, v := range result {
			if v != expected[i] {
				t.Errorf("At index %d: expected %s, got %s", i, expected[i], v)
			}
		}
	})

	t.Run("CombineWith different types", func(t *testing.T) {
		flow1 := Of(1, 2, 3)
		flow2 := Of("a", "b", "c")

		result := CombineWith(flow1, flow2, func(num int, str string) string {
			return fmt.Sprintf("%d-%s", num, str)
		}).Collect()

		expected := []string{"1-a", "2-b", "3-c"}

		for i, v := range result {
			if v != expected[i] {
				t.Errorf("At index %d: expected %s, got %s", i, expected[i], v)
			}
		}
	})

	t.Run("CombineWith flows of different lengths", func(t *testing.T) {
		flow1 := Range(1, 11) // 1 to 10
		flow2 := Of(100, 200, 300)

		result := CombineWith(flow1, flow2, func(a, b int) int {
			return a * b
		}).Collect()

		if len(result) != 3 {
			t.Errorf("Expected 3 elements, got %d", len(result))
		}

		expected := []int{100, 400, 900}
		for i, v := range result {
			if v != expected[i] {
				t.Errorf("At index %d: expected %d, got %d", i, expected[i], v)
			}
		}
	})

}

func TestCombineChaining(t *testing.T) {
	t.Run("Combine in a chain", func(t *testing.T) {
		flow1 := Range(1, 6)   // 1, 2, 3, 4, 5
		flow2 := Range(10, 15) // 10, 11, 12, 13, 14

		pairs := Combine(flow1, flow2).Collect()
		var sums []int
		for _, p := range pairs {
			sums = append(sums, p.First+p.Second)
		}

		result := NewFlow(sums).
			Filter(func(sum int) bool {
				return sum%2 == 0 // Only even sums
			}).
			Collect()

		if len(result) != 0 {
			t.Errorf("Expected empty result for odd sums, got %v", result)
		}

		flow3 := Range(2, 7)   // 2, 3, 4, 5, 6
		flow4 := Range(10, 15) // 10, 11, 12, 13, 14

		// Use CombineWith directly for transformation
		result2 := CombineWith(flow3, flow4, func(a, b int) int {
			return a + b
		}).
			Filter(func(sum int) bool {
				return sum%2 == 0 // Only even sums
			}).
			Collect()

		expected := []int{12, 14, 16, 18, 20}

		if len(result2) != len(expected) {
			t.Errorf("Expected %d even sums, got %d", len(expected), len(result2))
		}

		for i, v := range result2 {
			if v != expected[i] {
				t.Errorf("At index %d: expected %d, got %d", i, expected[i], v)
			}
		}
	})

	t.Run("CombineWith in a chain", func(t *testing.T) {
		names := Of("Alice", "Bob", "Charlie")
		ages := Of(25, 30, 35)

		type Person struct {
			Name string
			Age  int
		}

		people := CombineWith(names, ages, func(name string, age int) Person {
			return Person{Name: name, Age: age}
		}).
			Filter(func(p Person) bool {
				return p.Age >= 30
			}).
			Collect()

		var descriptions []string
		for _, p := range people {
			descriptions = append(descriptions, fmt.Sprintf("%s is %d years old", p.Name, p.Age))
		}
		result := descriptions

		expected := []string{
			"Bob is 30 years old",
			"Charlie is 35 years old",
		}

		if len(result) != len(expected) {
			t.Errorf("Expected %d results, got %d", len(expected), len(result))
		}

		for i, v := range result {
			if v != expected[i] {
				t.Errorf("At index %d: expected %s, got %s", i, expected[i], v)
			}
		}
	})
}
