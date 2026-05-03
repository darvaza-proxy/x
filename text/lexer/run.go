package lexer

// StateFn is one state in a state-function machine. It returns the
// next state, or nil to stop. A non-nil error terminates [Run].
//
// P is the parser-state type; it is threaded through every
// transition. Embed *[Cursor] in P to share the scanning primitives.
type StateFn[P any] func(P) (StateFn[P], error)

// Run drives the state machine starting at start, threading p
// through each transition until a state returns nil or an error.
//
// A nil start is a no-op.
func Run[P any](p P, start StateFn[P]) error {
	for fn := start; fn != nil; {
		next, err := fn(p)
		if err != nil {
			return err
		}
		fn = next
	}
	return nil
}
