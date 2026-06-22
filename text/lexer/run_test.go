package lexer_test

import (
	"errors"
	"testing"

	"darvaza.org/core"
	"darvaza.org/x/text/lexer"
)

var errBoom = errors.New("boom")

const failMark = 42

var _ core.TestCase = chainCase{}

func TestRun(t *testing.T) {
	t.Run("nil start", runTestRunNilStart)
	t.Run("chain", runTestRunChain)
	t.Run("error propagates", runTestRunError)
	t.Run("with cursor", runTestRunWithCursor)
}

func runTestRunNilStart(t *testing.T) {
	t.Helper()
	err := lexer.Run[*int](nil, nil)
	core.AssertNoError(t, err, "nil start")
}

type runCounter struct{ n int }

func sum(values ...int) int {
	total := 0
	for _, v := range values {
		total += v
	}
	return total
}

func newChainCounter(deltas ...int) (c *runCounter, start lexer.StateFn[*runCounter]) {
	states := make([]lexer.StateFn[*runCounter], len(deltas))
	for i := len(deltas) - 1; i >= 0; i-- {
		delta := deltas[i]
		var next lexer.StateFn[*runCounter]
		if i+1 < len(states) {
			next = states[i+1]
		}
		states[i] = func(p *runCounter) (lexer.StateFn[*runCounter], error) {
			p.n += delta
			return next, nil
		}
	}
	if len(states) > 0 {
		start = states[0]
	}
	return &runCounter{}, start
}

type chainCase struct {
	name   string
	deltas []int
}

func (tc chainCase) Name() string { return tc.name }

func (tc chainCase) Test(t *testing.T) {
	t.Helper()
	c, start := newChainCounter(tc.deltas...)
	err := lexer.Run(c, start)
	core.AssertNoError(t, err, "run")
	core.AssertEqual(t, sum(tc.deltas...), c.n, "count")
}

func newChainCase(name string, deltas ...int) chainCase {
	return chainCase{name: name, deltas: deltas}
}

func chainCases() []chainCase {
	return []chainCase{
		newChainCase("empty"),
		newChainCase("single", 42),
		newChainCase("pair", 1, 10),
		newChainCase("triple", 1, 10, 100),
	}
}

func runTestRunChain(t *testing.T) {
	t.Helper()
	core.RunTestCases(t, chainCases())
}

func newFailScenario() (p *int, fail lexer.StateFn[*int]) {
	fail = func(state *int) (lexer.StateFn[*int], error) {
		*state = failMark
		return nil, errBoom
	}
	return new(int), fail
}

func runTestRunError(t *testing.T) {
	t.Helper()
	p, fail := newFailScenario()
	err := lexer.Run(p, fail)
	core.AssertErrorIs(t, err, errBoom, "propagates")
	core.AssertEqual(t, failMark, *p, "state threaded to fail")
}

type runParser struct {
	*lexer.Cursor
	count int
}

func runConsumeOnce(p *runParser) (lexer.StateFn[*runParser], error) {
	if _, ok := p.Consume(); !ok {
		return nil, nil
	}
	p.count++
	return nil, nil
}

func runConsumeTwice(p *runParser) (lexer.StateFn[*runParser], error) {
	if _, ok := p.Consume(); !ok {
		return nil, nil
	}
	p.count++
	return runConsumeOnce, nil
}

func runTestRunWithCursor(t *testing.T) {
	t.Helper()
	p := &runParser{Cursor: lexer.New("ñ€𐍈")}
	err := lexer.Run(p, runConsumeTwice)
	core.AssertNoError(t, err, "run")
	core.AssertEqual(t, 2, p.count, "transitions")
	r, ok := p.Peek()
	core.AssertEqual(t, true, ok, "remaining")
	core.AssertEqual(t, '𐍈', r, "remaining rune")
}
