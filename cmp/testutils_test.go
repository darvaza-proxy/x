package cmp

// Fixture names shared by the person-matching tests
const (
	nameAlice   = "Alice"
	nameBob     = "Bob"
	nameCharlie = "Charlie"
	nameDavid   = "David"
)

// Expected panic messages for nil function tests
var (
	expectedNilCompFuncErr = "nil comparison function"
	expectedNilCondFuncErr = "nil condition function"
)
