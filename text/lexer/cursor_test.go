package lexer_test

import (
	"testing"

	"darvaza.org/core"
	"darvaza.org/x/text/lexer"
)

var (
	_ core.TestCase = peekCase{}
	_ core.TestCase = consumeCase{}
)

// peekCase asserts the first rune Peek returns for a given source.
type peekCase struct {
	name string
	src  string
	want rune
	ok   bool
}

func (tc peekCase) Name() string { return tc.name }

func (tc peekCase) Test(t *testing.T) {
	t.Helper()
	c := lexer.New(tc.src)
	got, ok := c.Peek()
	core.AssertEqual(t, tc.ok, ok, "ok")
	if tc.ok {
		core.AssertEqual(t, tc.want, got, "rune")
	}
	core.AssertEqual(t, !tc.ok, c.Done(), "Done agrees with !ok")
}

//revive:disable-next-line:flag-argument
func newPeekCase(name, src string, want rune, ok bool) peekCase {
	return peekCase{name: name, src: src, want: want, ok: ok}
}

func peekCases() []peekCase {
	return []peekCase{
		newPeekCase("empty", "", 0, false),
		newPeekCase("ascii", "abc", 'a', true),
		// cspell:disable-next-line
		newPeekCase("utf8 2-byte", "ñabc", 'ñ', true),
		newPeekCase("utf8 3-byte", "€abc", '€', true),
		// cspell:disable-next-line
		newPeekCase("utf8 4-byte", "𐍈abc", '𐍈', true),
	}
}

func TestPeek(t *testing.T) {
	core.RunTestCases(t, peekCases())
}

// consumeCase asserts the rune sequence produced by repeated
// [lexer.Cursor.Consume] calls until the source is exhausted.
type consumeCase struct {
	want []rune
	name string
	src  string
}

func (tc consumeCase) Name() string { return tc.name }

func (tc consumeCase) Test(t *testing.T) {
	t.Helper()
	c := lexer.New(tc.src)
	var got []rune
	for {
		r, ok := c.Consume()
		if !ok {
			break
		}
		got = append(got, r)
	}
	core.AssertSliceEqual(t, tc.want, got, "runes")
	core.AssertEqual(t, true, c.Done(), "Done after exhaustion")
	_, ok := c.Consume()
	core.AssertEqual(t, false, ok, "Consume past end")
}

func newConsumeCase(name, src string, want ...rune) consumeCase {
	return consumeCase{name: name, src: src, want: want}
}

func consumeCases() []consumeCase {
	return []consumeCase{
		newConsumeCase("empty", ""),
		newConsumeCase("ascii", "abc", 'a', 'b', 'c'),
		newConsumeCase("mixed widths", "ñ€𐍈x", 'ñ', '€', '𐍈', 'x'),
	}
}

func TestConsume(t *testing.T) {
	core.RunTestCases(t, consumeCases())
}

func TestAdvance(t *testing.T) {
	c := lexer.New("ñ€𐍈x")
	c.Advance(0)
	r, _ := c.Peek()
	core.AssertEqual(t, 'ñ', r, "no-op advance")
	c.Advance(1)
	r, _ = c.Peek()
	core.AssertEqual(t, '€', r, "advance one rune")
	c.Advance(2)
	r, _ = c.Peek()
	core.AssertEqual(t, 'x', r, "advance two runes")
	c.Advance(10)
	core.AssertEqual(t, true, c.Done(), "over-advance clamps to end")
	c.Advance(-1)
	core.AssertEqual(t, true, c.Done(), "negative advance is no-op")
}

func TestEmit(t *testing.T) {
	c := lexer.New("hello")
	c.Keep("greeting=")
	r, _ := c.Consume()
	core.AssertEqual(t, 'h', r, "consumed h")
	c.KeepRune('H')
	c.EmitRest()
	core.AssertEqual(t, "greeting=Hello", c.Emitted(), "emitted")
	core.AssertEqual(t, true, c.Done(), "EmitRest reaches end")
	c.Reset()
	core.AssertEqual(t, "", c.Emitted(), "after reset")
}

func TestEmitRestEmpty(t *testing.T) {
	c := lexer.New("")
	c.EmitRest()
	core.AssertEqual(t, "", c.Emitted(), "no input → empty emit")
	core.AssertEqual(t, true, c.Done(), "still at end")
}
