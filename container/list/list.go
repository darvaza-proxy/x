// Package list provides a type-safe wrapper to the standard container/list
package list

import (
	"container/list"

	"darvaza.org/core"
)

// List is a typed wrapper on top of [list.List].
type List[T any] list.List

// New creates a new List with optional initial entries.
// It returns a pointer to the newly created List.
func New[T any](entries ...T) *List[T] {
	l := (*List[T])(list.New())
	for _, e := range entries {
		l.PushBack(e)
	}

	return l
}

// Sys returns the native [list.List]
func (l *List[T]) Sys() *list.List {
	if l == nil {
		return nil
	}
	return (*list.List)(l)
}

// Len returns the number of elements in the list
func (l *List[T]) Len() int {
	if l == nil {
		return 0
	}
	return l.Sys().Len()
}

// Zero returns the zero value of the type associated
// to the list.
func (*List[T]) Zero() T {
	var zero T
	return zero
}

// PushFront adds a value at the beginning of the list.
func (l *List[T]) PushFront(v T) {
	if l != nil {
		l.Sys().PushFront(v)
	}
}

// PushBack adds a value at the end of the list.
func (l *List[T]) PushBack(v T) {
	if l != nil {
		l.Sys().PushBack(v)
	}
}

// Front returns the first value in the list.
func (l *List[T]) Front() (T, bool) {
	var out T
	var found bool

	if l != nil {
		_, out, found = l.unsafeFirstMatchElement(nil)
	}

	return out, found
}

// Back returns the last value in the list.
func (l *List[T]) Back() (T, bool) {
	var out T
	var found bool

	if l != nil {
		_, out, found = l.unsafeFirstMatchBackwardElement(nil)
	}

	return out, found
}

// Values returns all values in the list.
func (l *List[T]) Values() []T {
	out := make([]T, 0, l.Len())
	l.ForEach(func(v T) bool {
		out = append(out, v)
		return true
	})
	return out
}

// ForEach calls a function on each element of the list, allowing safe modification of the
// list during iteration.
func (l *List[T]) ForEach(fn func(T) bool) {
	if l != nil && fn != nil {
		cb := func(_ *list.Element, v T) bool {
			return fn(v)
		}

		l.unsafeForEachElement(cb)
	}
}

// DeleteMatchFn deletes elements on the list satisfying the given function.
func (l *List[T]) DeleteMatchFn(fn func(T) bool) {
	if ll := l.Sys(); ll != nil && fn != nil {
		remove := l.unsafeGetAllMatchElement(fn)
		if len(remove) > 0 {
			for _, el := range remove {
				ll.Remove(el)
			}
		}
	}
}

// PopFirstMatchFn removes and returns the first match, iterating from
// front to back.
func (l *List[T]) PopFirstMatchFn(fn func(T) bool) (T, bool) {
	var out T
	var found bool

	if l != nil && fn != nil {
		var el *list.Element

		el, out, found = l.unsafeFirstMatchElement(fn)
		if el != nil {
			l.Sys().Remove(el)
		}
	}

	return out, found
}

// MoveToBackFirstMatchFn moves the first element that satisfies the given function
// to the back of the list, searching from front to back.
func (l *List[T]) MoveToBackFirstMatchFn(fn func(T) bool) {
	if l != nil && fn != nil {
		var el *list.Element

		el, _, _ = l.unsafeFirstMatchElement(fn)
		if el != nil {
			l.Sys().MoveToBack(el)
		}
	}
}

// MoveToFrontFirstMatchFn moves the first element that satisfies the given function
// to the front of the list, searching from front to back.
func (l *List[T]) MoveToFrontFirstMatchFn(fn func(T) bool) {
	if l != nil && fn != nil {
		var el *list.Element

		el, _, _ = l.unsafeFirstMatchElement(fn)
		if el != nil {
			l.Sys().MoveToFront(el)
		}
	}
}

// FirstMatchFn returns the first element that satisfies the given function from
// the front to the back.
func (l *List[T]) FirstMatchFn(fn func(T) bool) (T, bool) {
	var out T
	var found bool

	if l != nil && fn != nil {
		_, out, found = l.unsafeFirstMatchElement(fn)
	}

	return out, found
}

// Purge removes any element not complying with the type restriction.
// It returns the number of elements removed.
func (l *List[T]) Purge() int {
	var count int

	if ll := l.Sys(); ll != nil {
		var remove []*list.Element

		cb := func(el *list.Element) bool {
			if _, ok := el.Value.(T); !ok {
				remove = append(remove, el)
			}
			return true
		}

		core.ListForEachElement(ll, cb)

		for _, el := range remove {
			ll.Remove(el)
			count++
		}
	}

	return count
}

// Clone returns a shallow copy of the list.
func (l *List[T]) Clone() *List[T] {
	return l.Copy(nil)
}

// Copy returns a copy of the list, optionally altered or filtered.
func (l *List[T]) Copy(fn func(T) (T, bool)) *List[T] {
	if fn == nil {
		fn = func(v T) (T, bool) {
			return v, true
		}
	}

	out := new(List[T])
	if l != nil {
		l.ForEach(func(v T) bool {
			if v, ok := fn(v); ok {
				out.PushBack(v)
			}

			return true
		})
	}

	return out
}

func (l *List[T]) unsafeGetAllMatchElement(fn matchFunc[T]) []*list.Element {
	var elems []*list.Element
	_, cb := newGetAllMatchElement(fn, &elems)
	l.unsafeForEachElement(cb)
	return elems
}

func (l *List[T]) unsafeFirstMatchElement(fn matchFunc[T]) (*list.Element, T, bool) {
	var match *list.Element
	var out T

	_, _, cb := newGetMatchElement(fn, &match, &out)
	l.unsafeForEachElement(cb)
	return match, out, match != nil
}

func (l *List[T]) unsafeFirstMatchBackwardElement(fn matchFunc[T]) (*list.Element, T, bool) {
	var match *list.Element
	var value T

	_, _, cb := newGetMatchElement(fn, &match, &value)
	l.unsafeForEachBackwardElement(cb)
	return match, value, match != nil
}

func (l *List[T]) unsafeForEachElement(fn matchElemTypeFunc[T]) {
	cb := newMatchElement(fn)
	core.ListForEachElement(l.Sys(), cb)
}

func (l *List[T]) unsafeForEachBackwardElement(fn matchElemTypeFunc[T]) {
	cb := newMatchElement(fn)
	core.ListForEachBackwardElement(l.Sys(), cb)
}

// newGetAllMatchElement returns an iterator callback that will collect all matching elements.
func newGetAllMatchElement[T any](cond matchFunc[T],
	out *[]*list.Element) (*[]*list.Element, matchElemTypeFunc[T]) {
	//
	if out == nil {
		out = new([]*list.Element)
	}

	cb := func(el *list.Element, v T) bool {
		if cond == nil || cond(v) {
			*out = append(*out, el)
		}
		return true
	}

	return out, cb
}

// newGetMatchElement returns an iterator callback that will return the *list.Element and T value of the first matching
// entry.
func newGetMatchElement[T any](cond matchFunc[T],
	match **list.Element, out *T) (**list.Element, *T, matchElemTypeFunc[T]) {
	//
	if match == nil {
		match = new(*list.Element)
	}
	if out == nil {
		out = new(T)
	}

	cb := func(el *list.Element, v T) bool {
		if cond == nil || cond(v) {
			*match = el
			*out = v
		}

		return *match == nil
	}
	return match, out, cb
}

// newMatchElement returns an iterator callback that calls a helper for each (*list.Element, T) pair.
func newMatchElement[T any](fn matchElemTypeFunc[T]) matchElemFunc {
	cont := true
	return func(el *list.Element) bool {
		if value, ok := el.Value.(T); ok {
			cont = fn == nil || fn(el, value)
		}
		return cont
	}
}

type matchFunc[T any] func(T) bool
type matchElemTypeFunc[T any] func(*list.Element, T) bool
type matchElemFunc func(*list.Element) bool
