// Package list provides a type-safe wrapper to the standard container/list
package list

import "container/list"

// List is a typed wrapper on top of [list.List].
type List[T any] list.List

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
func (l *List[T]) Front() T {
	if l != nil {
		if p := l.Sys().Front(); p != nil {
			if v, ok := p.Value.(T); ok {
				return v
			}
		}
	}

	return l.Zero()
}

// Back returns the last value in the list.
func (l *List[T]) Back() T {
	if l != nil {
		if p := l.Sys().Back(); p != nil {
			if v, ok := p.Value.(T); ok {
				return v
			}
		}
	}

	return l.Zero()
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
		l.unsafeForEachElement(func(_ *list.Element, v T) bool {
			return fn(v)
		})
	}
}

// DeleteMatchFn deletes elements on the list satisfying the given function.
func (l *List[T]) DeleteMatchFn(fn func(T) bool) {
	if l != nil && fn != nil {
		cb := func(elem *list.Element, v T) bool {
			if fn(v) {
				l.Sys().Remove(elem)
			}
			return true
		}

		l.unsafeForEachElement(cb)
	}
}

// PopFirstMatchFn removes and returns the first match, iterating from
// front to back.
func (l *List[T]) PopFirstMatchFn(fn func(T) bool) (T, bool) {
	var out T
	var found bool

	if l != nil && fn != nil {
		cb := func(elem *list.Element, v T) bool {
			if fn(v) {
				out = v
				found = true
				l.Sys().Remove(elem)
				return false
			}
			return true
		}

		l.unsafeForEachElement(cb)
	}

	return out, found
}

// FirstMatchFn returns the first element that satisfies the given function from
// the front to the back.
func (l *List[T]) FirstMatchFn(fn func(T) bool) (T, bool) {
	var out T
	var found bool

	if l != nil && fn != nil {
		cb := func(_ *list.Element, v T) bool {
			if fn(v) {
				out = v
				found = true
				return false
			}
			return true
		}

		l.unsafeForEachElement(cb)
	}

	return out, found
}

func (l *List[T]) unsafeForEachElement(fn func(*list.Element, T) bool) {
	var next *list.Element
	for elem := l.Sys().Front(); elem != nil; elem = next {
		next = elem.Next()
		if value, ok := elem.Value.(T); ok {
			if !fn(elem, value) {
				return
			}
		}
	}
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
