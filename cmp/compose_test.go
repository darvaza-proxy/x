package cmp

//revive:disable:cognitive-complexity

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestCompose verifies that the Compose function properly creates
// matchers that apply transformations before matching.
func TestCompose(t *testing.T) {
	t.Run("with struct fields", func(t *testing.T) {
		type Person struct {
			Name string
			Age  int
		}

		// Create a matcher for people older than 18
		isAdult := Compose(
			func(p Person) (int, bool) { return p.Age, true },
			MatchGtEq(18),
		)

		tests := []struct {
			name     string
			person   Person
			expected bool
		}{
			{"adult", Person{"Alice", 30}, true},
			{"child", Person{"Bob", 12}, false},
			{"exact boundary", Person{"Charlie", 18}, true},
			{"just under boundary", Person{"David", 17}, false},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				assert.Equal(t, tt.expected, isAdult.Match(tt.person),
					"isAdult.Match(%+v) should return %v", tt.person, tt.expected)
			})
		}

		// Create a matcher to check if a person's name starts with 'A'
		nameStartsWithA := Compose(
			func(p Person) (string, bool) { return p.Name, true },
			MatchFunc[string](func(s string) bool {
				return len(s) > 0 && s[0] == 'A'
			}),
		)

		nameTests := []struct {
			name     string
			person   Person
			expected bool
		}{
			{"starts with A", Person{"Alice", 30}, true},
			{"starts with different letter", Person{"Bob", 25}, false},
			{"empty name", Person{"", 20}, false},
		}

		for _, tt := range nameTests {
			t.Run(tt.name, func(t *testing.T) {
				assert.Equal(t, tt.expected, nameStartsWithA.Match(tt.person),
					"nameStartsWithA.Match(%+v) should return %v", tt.person, tt.expected)
			})
		}
	})

	t.Run("with type transformations", func(t *testing.T) {
		// Convert int to string and check if it has a specific length
		hasThreeDigits := Compose(
			func(i int) (string, bool) { return strconv.Itoa(i), true },
			MatchFunc[string](func(s string) bool {
				return len(s) == 3
			}),
		)

		tests := []struct {
			name     string
			input    int
			expected bool
		}{
			{"three digits", 123, true},
			{"two digits", 45, false},
			{"four digits", 1000, false},
			{"negative three digits", -123, false}, // has 4 chars due to minus sign
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				assert.Equal(t, tt.expected, hasThreeDigits.Match(tt.input),
					"hasThreeDigits.Match(%d) should return %v", tt.input, tt.expected)
			})
		}
	})

	t.Run("with accessor returning false", func(t *testing.T) {
		// Accessor that fails for even numbers
		failsForEven := Compose(
			func(i int) (int, bool) { return i, i%2 != 0 },
			MatchGt(0),
		)

		tests := []struct {
			name     string
			input    int
			expected bool
		}{
			{"odd positive", 3, true},
			{"odd negative", -5, false},  // fails the Gt(0) check
			{"even positive", 4, false},  // accessor returns false
			{"even negative", -2, false}, // accessor returns false
			{"zero", 0, false},           // accessor returns false (0 is even)
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				assert.Equal(t, tt.expected, failsForEven.Match(tt.input),
					"failsForEven.Match(%d) should return %v", tt.input, tt.expected)
			})
		}
	})

	t.Run("nested composition", func(t *testing.T) {
		type Address struct {
			Street string
			City   string
			Zip    string
		}

		type Person struct {
			Name    string
			Age     int
			Address Address
		}

		// Create a matcher that checks if a person's city starts with 'New'
		isFromNewCity := Compose(
			func(p Person) (Address, bool) { return p.Address, true },
			Compose(
				func(a Address) (string, bool) { return a.City, true },
				MatchFunc[string](func(s string) bool {
					return len(s) >= 3 && s[0:3] == "New"
				}),
			),
		)

		tests := []struct {
			name     string
			person   Person
			expected bool
		}{
			{
				"from new city",
				Person{"Alice", 30, Address{"123 Broadway", "New York", "10001"}},
				true,
			},
			{
				"from another new city",
				Person{"Bob", 25, Address{"456 Main St", "New Orleans", "70112"}},
				true,
			},
			{
				"not from new city",
				Person{"Charlie", 40, Address{"789 Oak Dr", "Boston", "02108"}},
				false,
			},
			{
				"empty city",
				Person{"David", 22, Address{"101 Pine Ave", "", "90210"}},
				false,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				assert.Equal(t, tt.expected, isFromNewCity.Match(tt.person),
					"isFromNewCity.Match(%+v) should return %v", tt.person, tt.expected)
			})
		}
	})

	t.Run("complex conditions", func(t *testing.T) {
		type Score struct {
			Value int
			Max   int
		}

		// Check if a score is a passing grade (>= 70%)
		isPassingGrade := Compose(
			func(s Score) (float64, bool) {
				if s.Max == 0 {
					return 0, false
				}
				return float64(s.Value) / float64(s.Max), true
			},
			MatchGtEq(0.7),
		)

		tests := []struct {
			name     string
			score    Score
			expected bool
		}{
			{"perfect score", Score{100, 100}, true},
			{"passing score", Score{70, 100}, true},
			{"failing score", Score{69, 100}, false},
			{"zero score", Score{0, 100}, false},
			{"invalid score", Score{50, 0}, false}, // division by zero avoided
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				assert.Equal(t, tt.expected, isPassingGrade.Match(tt.score),
					"isPassingGrade.Match(%+v) should return %v", tt.score, tt.expected)
			})
		}
	})
}

// TestComposePanic verifies that Compose panics when given nil arguments.
func TestComposePanic(t *testing.T) {
	t.Run("nil accessor function", func(t *testing.T) {
		assert.Panics(t, func() {
			Compose[int, int](nil, MatchGtEq(5))
		}, "Compose with nil accessor function should panic")
	})

	t.Run("nil matcher", func(t *testing.T) {
		assert.Panics(t, func() {
			Compose(func(i int) (int, bool) { return i, true }, nil)
		}, "Compose with nil matcher should panic")
	})
}
