package list_test

import (
	"fmt"

	"darvaza.org/x/container/list"
)

func ExampleNew() {
	// Create a new list with initial values
	l := list.New(1, 2, 3)
	fmt.Println("Length:", l.Len())
	fmt.Println("Values:", l.Values())
	// Output:
	// Length: 3
	// Values: [1 2 3]
}

func ExampleList_PushFront() {
	l := list.New[string]()
	l.PushBack("world")
	l.PushFront("hello")
	fmt.Println(l.Values())
	// Output: [hello world]
}

func ExampleList_PushBack() {
	l := list.New[int]()
	l.PushBack(1)
	l.PushBack(2)
	l.PushBack(3)
	fmt.Println(l.Values())
	// Output: [1 2 3]
}

func ExampleList_Front() {
	l := list.New("first", "second", "third")
	if v, ok := l.Front(); ok {
		fmt.Println("Front:", v)
	}
	// Output: Front: first
}

func ExampleList_Back() {
	l := list.New("first", "second", "third")
	if v, ok := l.Back(); ok {
		fmt.Println("Back:", v)
	}
	// Output: Back: third
}

func ExampleList_ForEach() {
	l := list.New(1, 2, 3, 4, 5)
	sum := 0
	l.ForEach(func(v int) bool {
		sum += v
		return true // continue iteration
	})
	fmt.Println("Sum:", sum)
	// Output: Sum: 15
}

func ExampleList_ForEach_earlyTermination() {
	l := list.New(1, 2, 3, 4, 5)
	var collected []int
	l.ForEach(func(v int) bool {
		collected = append(collected, v)
		return v < 3 // stop when v >= 3
	})
	fmt.Println("Collected:", collected)
	// Output: Collected: [1 2 3]
}

func ExampleList_DeleteMatchFn() {
	l := list.New(1, 2, 3, 4, 5)
	// Delete even numbers
	l.DeleteMatchFn(func(v int) bool {
		return v%2 == 0
	})
	fmt.Println("After deletion:", l.Values())
	// Output: After deletion: [1 3 5]
}

func ExampleList_FirstMatchFn() {
	l := list.New(1, 2, 3, 4, 5)
	// Find first even number
	if v, ok := l.FirstMatchFn(func(n int) bool {
		return n%2 == 0
	}); ok {
		fmt.Println("First even:", v)
	}
	// Output: First even: 2
}

func ExampleList_Copy() {
	l := list.New(1, 2, 3, 4, 5)
	// Copy only even numbers, doubled
	copied := l.Copy(func(v int) (int, bool) {
		if v%2 == 0 {
			return v * 2, true
		}
		return 0, false
	})
	fmt.Println("Original:", l.Values())
	fmt.Println("Copied:", copied.Values())
	// Output:
	// Original: [1 2 3 4 5]
	// Copied: [4 8]
}

func ExampleList_Clone() {
	l := list.New("a", "b", "c")
	cloned := l.Clone()
	cloned.PushBack("d")
	fmt.Println("Original:", l.Values())
	fmt.Println("Cloned:", cloned.Values())
	// Output:
	// Original: [a b c]
	// Cloned: [a b c d]
}
