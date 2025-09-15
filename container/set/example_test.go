package set_test

import (
	"fmt"

	"darvaza.org/core"
	"darvaza.org/x/container/set"
)

// User represents a user with ID and name.
type User struct {
	Name string
	ID   int
}

func ExampleConfig_New() {
	// Create a configuration for a set of users
	cfg := set.Config[int, int, User]{
		ItemKey:   func(u User) (int, error) { return u.ID, nil },
		Hash:      func(id int) (int, error) { return id % 10, nil },
		ItemMatch: func(k int, u User) bool { return k == u.ID },
	}

	// Create a new set with initial users
	users, err := cfg.New(
		User{ID: 1, Name: "Alice"},
		User{ID: 2, Name: "Bob"},
	)
	if err != nil {
		panic(err)
	}

	// Check if user exists
	fmt.Println("Contains user 1:", users.Contains(1))
	fmt.Println("Contains user 3:", users.Contains(3))

	// Output:
	// Contains user 1: true
	// Contains user 3: false
}

func ExampleSet_Push() {
	cfg := set.Config[string, string, string]{
		ItemKey:   func(s string) (string, error) { return s, nil },
		Hash:      func(s string) (string, error) { return s, nil },
		ItemMatch: func(k, v string) bool { return k == v },
	}

	tags, _ := cfg.New()

	// Add tags
	stored, err := tags.Push("golang")
	fmt.Printf("Added golang: %q (err: %v)\n", stored, err)

	stored, err = tags.Push("testing")
	fmt.Printf("Added testing: %q (err: %v)\n", stored, err)

	// Try to add duplicate
	_, err = tags.Push("golang")
	fmt.Println("Adding duplicate golang:", err)

	// Output:
	// Added golang: "golang" (err: <nil>)
	// Added testing: "testing" (err: <nil>)
	// Adding duplicate golang: already exists
}

func ExampleSet_Get() {
	cfg := set.Config[int, int, User]{
		ItemKey:   func(u User) (int, error) { return u.ID, nil },
		Hash:      func(id int) (int, error) { return id % 10, nil },
		ItemMatch: func(k int, u User) bool { return k == u.ID },
	}

	users, _ := cfg.New(
		User{ID: 1, Name: "Alice"},
		User{ID: 2, Name: "Bob"},
	)

	// Get existing user
	if user, err := users.Get(1); err == nil {
		fmt.Printf("User 1: %s\n", user.Name)
	}

	// Try to get non-existent user
	if _, err := users.Get(3); err != nil {
		fmt.Printf("User 3: %v\n", err)
	}

	// Output:
	// User 1: Alice
	// User 3: does not exist
}

func ExampleSet_ForEach() {
	cfg := set.Config[int, int, int]{
		ItemKey:   func(n int) (int, error) { return n, nil },
		Hash:      func(n int) (int, error) { return n % 10, nil },
		ItemMatch: func(k, v int) bool { return k == v },
	}

	numbers, _ := cfg.New(1, 2, 3, 4, 5)

	// Sum all numbers
	sum := 0
	numbers.ForEach(func(n int) bool {
		sum += n
		return true // continue iteration
	})

	fmt.Println("Sum:", sum)

	// Find all even numbers
	var evens []int
	numbers.ForEach(func(n int) bool {
		if n%2 == 0 {
			evens = append(evens, n)
		}
		return true
	})

	// Sort for consistent output
	core.SliceSortOrdered(evens)
	fmt.Printf("Even numbers: %v\n", evens)

	// Output:
	// Sum: 15
	// Even numbers: [2 4]
}

func ExampleSet_Values() {
	cfg := set.Config[string, string, string]{
		ItemKey:   func(s string) (string, error) { return s, nil },
		Hash:      func(s string) (string, error) { return s, nil },
		ItemMatch: func(k, v string) bool { return k == v },
	}

	fruits, _ := cfg.New("apple", "banana", "cherry")

	// Get all values
	allFruits := fruits.Values()
	fmt.Printf("Total fruits: %d\n", len(allFruits))

	// Note: order is not guaranteed
	found := make(map[string]bool)
	for _, fruit := range allFruits {
		found[fruit] = true
	}

	// Check specific fruits
	for _, fruit := range []string{"apple", "banana", "cherry"} {
		fmt.Printf("Has %s: %v\n", fruit, found[fruit])
	}

	// Output:
	// Total fruits: 3
	// Has apple: true
	// Has banana: true
	// Has cherry: true
}

func ExampleSet_Copy() {
	cfg := set.Config[int, int, User]{
		ItemKey:   func(u User) (int, error) { return u.ID, nil },
		Hash:      func(id int) (int, error) { return id % 10, nil },
		ItemMatch: func(k int, u User) bool { return k == u.ID },
	}

	users, _ := cfg.New(
		User{ID: 1, Name: "Alice"},
		User{ID: 2, Name: "Bob"},
		User{ID: 3, Name: "Charlie"},
		User{ID: 4, Name: "Diana"},
	)

	// Copy only users with even IDs
	evenUsers := users.Copy(nil, func(u User) bool {
		return u.ID%2 == 0
	})

	// Collect and sort for consistent output
	var names []string
	evenUsers.ForEach(func(u User) bool {
		names = append(names, u.Name)
		return true
	})
	core.SliceSortOrdered(names)

	fmt.Print("Even users:")
	for _, name := range names {
		fmt.Printf(" %s", name)
	}
	fmt.Println()

	// Output:
	// Even users: Bob Diana
}

func ExampleSet_Clone() {
	cfg := set.Config[string, string, string]{
		ItemKey:   func(s string) (string, error) { return s, nil },
		Hash:      func(s string) (string, error) { return s, nil },
		ItemMatch: func(k, v string) bool { return k == v },
	}

	original, _ := cfg.New("red", "green", "blue")
	clone := original.Clone()

	// Modify original
	_, _ = original.Push("yellow")

	fmt.Println("Original has yellow:", original.Contains("yellow"))
	fmt.Println("Clone has yellow:", clone.Contains("yellow"))

	// Output:
	// Original has yellow: true
	// Clone has yellow: false
}
