package atomic

import "sync/atomic"

// These aliases exist because we shadow the sync/atomic package. They let
// callers reach the standard-library atomic types through a single import
// of darvaza.org/x/sync/atomic.

// Bool aliases sync/atomic.Bool.
type Bool = atomic.Bool

// Int32 aliases sync/atomic.Int32.
type Int32 = atomic.Int32

// Int64 aliases sync/atomic.Int64.
type Int64 = atomic.Int64

// Uint32 aliases sync/atomic.Uint32.
type Uint32 = atomic.Uint32

// Uint64 aliases sync/atomic.Uint64.
type Uint64 = atomic.Uint64

// Uintptr aliases sync/atomic.Uintptr.
type Uintptr = atomic.Uintptr

// Value aliases sync/atomic.Value.
type Value = atomic.Value

// Pointer aliases sync/atomic.Pointer.
type Pointer[T any] = atomic.Pointer[T]
