# `darvaza.org/x/container/slices`

The `slices` package provides a higher level abstractions and functionality
based on the standard Go slices.

## Set[T]

`Set[T]` is a generic interface that represents a set of values implemented over a slice.

`CustomSet[T]` is a generic type that implements the `Set[T]` interface using a sorted slice
underneath. A comparison function is required to be provided at initialization time via
`NewCustomSet[T]()` or `MustCustomSet[T]()` and the initial values can optionally be included
as extra arguments.

`NewOrderedSet[T]()` is a convenience factory using a default comparison function for
ordered generic types.

For use with embedded `CustomSet[T]` variables, `InitCustomSet[T]` and `InitOrderedSet[T]`
are available.

## See also

* [`darvaza.org/x/container/set`](https://darvaza.org/x/container/set): A set implementation that uses a map internally.
* [`darvaza.org/x/container/list`](https://darvaza.org/x/container/list): A doubly linked list implementation.
* [`darvaza.org/x`](https://github.com/darvaza-proxy/x): Home of all opinionated but application independent darvaza packages.
* [`darvaza.org/core`](https://darvaza.org/core): Home of all unopinionated darvaza extensions to the standard library.
