package buffer_test

import (
	"fmt"
	"io"
	"testing"

	"darvaza.org/core"

	"darvaza.org/x/text/buffer"
)

// failingWriter is an [io.Writer] that always fails with
// [io.ErrShortWrite]. Used to exercise error-propagation paths.
type failingWriter struct{}

func (failingWriter) Write(_ []byte) (int, error) { return 0, io.ErrShortWrite }

// ── NewBuffer capacity (table-driven) ─────────────────────────────

var _ core.TestCase = capacityTestCase{}

type capacityTestCase struct {
	name     string
	input    string
	capacity int
}

func newCapacityTestCase(name string, capacity int, input string) capacityTestCase {
	return capacityTestCase{
		name:     name,
		input:    input,
		capacity: capacity,
	}
}

func (tc capacityTestCase) Name() string { return tc.name }

func (tc capacityTestCase) Test(t *testing.T) {
	t.Helper()
	buf := buffer.New(tc.capacity)
	core.AssertMustNotNil(t, buf, "buffer")
	buf.WriteStrings(tc.input)
	core.AssertEqual(t, tc.input, buf.String(), "content")
}

func TestNewBuffer(t *testing.T) {
	cases := []capacityTestCase{
		newCapacityTestCase("positive capacity", 100, "test"),
		newCapacityTestCase("zero capacity", 0, "zero"),
		newCapacityTestCase("negative capacity", -10, "negative"),
	}
	core.RunTestCases(t, cases)
}

// ── Buffer scenarios ──────────────────────────────────────────────

func TestBuffer(t *testing.T) {
	t.Run("method chaining", runTestMethodChaining)
	t.Run("reset", runTestReset)
	t.Run("write strings", runTestWriteStrings)
	t.Run("write bytes", runTestWriteBytes)
	t.Run("write runes", runTestWriteRunes)
	t.Run("print methods", runTestPrintMethods)
	t.Run("print spacing", runTestPrintSpacing)
	t.Run("grow", runTestGrow)
	t.Run("grow non-positive", runTestGrowNonPositive)
	t.Run("io writer", runTestIOWriter)
	t.Run("write to", runTestWriteTo)
	t.Run("write to error", runTestWriteToError)
	t.Run("bytes aliasing", runTestBytesAliasing)
	t.Run("bytes empty", runTestBytesEmpty)
	t.Run("print chain identity", runTestPrintChainIdentity)
	t.Run("write string singular", runTestWriteStringSingular)
	t.Run("write byte singular", runTestWriteByteSingular)
	t.Run("write rune singular", runTestWriteRuneSingular)
	t.Run("complex chaining", runTestComplexChaining)
}

func runTestMethodChaining(t *testing.T) {
	t.Helper()
	var buf buffer.Buffer
	result := buf.WriteStrings("hello", " ", "world").
		WriteRunes('!').
		Printf(" number: %d", 42).
		WriteBytes([]byte(" end"))
	core.AssertSame(t, &buf, result, "chain identity")
	core.AssertEqual(t, "hello world! number: 42 end", buf.String(), "content")
}

func runTestReset(t *testing.T) {
	t.Helper()
	var buf buffer.Buffer
	buf.WriteStrings("some text")
	core.AssertMustEqual(t, 9, buf.Len(), "pre-reset len")
	result := buf.Reset()
	core.AssertSame(t, &buf, result, "reset identity")
	core.AssertEqual(t, 0, buf.Len(), "post-reset len")
	core.AssertEqual(t, "", buf.String(), "post-reset string")
}

func runTestWriteStrings(t *testing.T) {
	t.Helper()
	var buf buffer.Buffer
	buf.WriteStrings("a", "b", "c")
	core.AssertEqual(t, "abc", buf.String(), "content")
}

func runTestWriteBytes(t *testing.T) {
	t.Helper()
	var buf buffer.Buffer
	buf.WriteBytes([]byte("hello"), []byte(" "), []byte("world"))
	core.AssertEqual(t, "hello world", buf.String(), "content")
}

func runTestWriteRunes(t *testing.T) {
	t.Helper()
	var buf buffer.Buffer
	buf.WriteRunes('H', 'e', 'l', 'p', '!', '🚀')
	core.AssertEqual(t, "Help!🚀", buf.String(), "content")
}

func runTestPrintMethods(t *testing.T) {
	t.Helper()
	var buf buffer.Buffer
	buf.Print("hello").
		Print(" ").
		Println("world").
		Printf("number: %d", 123)
	core.AssertEqual(t, "hello world\nnumber: 123", buf.String(), "content")
}

func runTestPrintSpacing(t *testing.T) {
	t.Helper()
	var buf buffer.Buffer

	buf.Print("a", "b")
	core.AssertEqual(t, "ab", buf.String(), "Print two strings (no space)")

	buf.Reset().Print("k=", 42)
	core.AssertEqual(t, "k=42", buf.String(), "Print string+non-string (no space)")

	buf.Reset().Print(1, 2)
	core.AssertEqual(t, "1 2", buf.String(), "Print two non-strings (space)")

	buf.Reset().Println("a", "b")
	core.AssertEqual(t, "a b\n", buf.String(), "Println two strings (space + newline)")

	buf.Reset().Println(1, 2)
	core.AssertEqual(t, "1 2\n", buf.String(), "Println two non-strings (space + newline)")
}

func runTestGrow(t *testing.T) {
	t.Helper()
	var buf buffer.Buffer
	result := buf.Grow(100)
	core.AssertSame(t, &buf, result, "grow identity")
	buf.WriteStrings("test")
	core.AssertEqual(t, "test", buf.String(), "content")
}

func runTestGrowNonPositive(t *testing.T) {
	t.Helper()
	var buf buffer.Buffer
	core.AssertNoPanic(t, func() { buf.Grow(0) }, "grow zero")
	core.AssertNoPanic(t, func() { buf.Grow(-10) }, "grow negative")
}

func runTestIOWriter(t *testing.T) {
	t.Helper()
	var buf buffer.Buffer
	n, err := buf.Write([]byte("test"))
	core.AssertNoError(t, err, "write")
	core.AssertEqual(t, 4, n, "bytes written")
	core.AssertEqual(t, 4, buf.Len(), "len")
	core.AssertSliceEqual(t, []byte("test"), buf.Bytes(), "bytes")
}

func runTestWriteTo(t *testing.T) {
	t.Helper()
	var src, dst buffer.Buffer
	src.WriteStrings("payload")
	n, err := src.WriteTo(&dst)
	core.AssertNoError(t, err, "writeTo")
	core.AssertEqual(t, int64(7), n, "bytes copied")
	core.AssertEqual(t, "payload", dst.String(), "dst content")
	core.AssertEqual(t, 0, src.Len(), "src drained")
	src.WriteStrings("again")
	core.AssertEqual(t, "again", src.String(), "src reusable after drain")
}

func runTestWriteToError(t *testing.T) {
	t.Helper()
	var buf buffer.Buffer
	buf.WriteStrings("payload")
	n, err := buf.WriteTo(failingWriter{})
	core.AssertError(t, err, "writeTo error propagated")
	core.AssertEqual(t, int64(0), n, "bytes reported")
	core.AssertEqual(t, "payload", buf.String(), "buffer intact after error")
}

func runTestBytesAliasing(t *testing.T) {
	t.Helper()
	var buf buffer.Buffer
	buf.WriteStrings("hello")
	s := buf.String()
	b1 := buf.Bytes()
	b2 := buf.Bytes()
	core.AssertEqual(t, "hello", string(b1), "initial content")
	core.AssertEqual(t, s, string(b1), "Bytes content matches String")
	// Bytes aliases internal storage (see [buffer.Buffer.Bytes]);
	// successive calls must return slices over the same backing array.
	core.AssertSame(t, b1, b2, "Bytes aliases stable storage")
}

func runTestBytesEmpty(t *testing.T) {
	t.Helper()
	var buf buffer.Buffer
	b := buf.Bytes()
	core.AssertNotNil(t, b, "non-nil slice")
	core.AssertEqual(t, 0, len(b), "empty len")
}

func runTestPrintChainIdentity(t *testing.T) {
	t.Helper()
	var buf buffer.Buffer
	core.AssertSame(t, &buf, buf.Print("x"), "Print identity")
	core.AssertSame(t, &buf, buf.Println("y"), "Println identity")
	core.AssertSame(t, &buf, buf.Printf("%s", "z"), "Printf identity")
}

func runTestWriteStringSingular(t *testing.T) {
	t.Helper()
	var buf buffer.Buffer
	n, err := buf.WriteString("hello")
	core.AssertNoError(t, err, "writeString")
	core.AssertEqual(t, 5, n, "bytes written")
	core.AssertEqual(t, "hello", buf.String(), "content")
}

func runTestWriteByteSingular(t *testing.T) {
	t.Helper()
	var buf buffer.Buffer
	err := buf.WriteByte('Z')
	core.AssertNoError(t, err, "writeByte")
	core.AssertEqual(t, "Z", buf.String(), "content")
}

func runTestWriteRuneSingular(t *testing.T) {
	t.Helper()
	var buf buffer.Buffer
	n, err := buf.WriteRune('🚀')
	core.AssertNoError(t, err, "writeRune")
	core.AssertEqual(t, 4, n, "utf-8 bytes")
	core.AssertEqual(t, "🚀", buf.String(), "content")
}

func runTestComplexChaining(t *testing.T) {
	t.Helper()
	var buf buffer.Buffer
	buf.Grow(100).
		WriteStrings(`{"name":"`).
		WriteStrings("John Doe").
		WriteStrings(`","age":`).
		Printf("%d", 30).
		WriteStrings(`,"active":`).
		Print(true).
		WriteRunes('}')
	core.AssertEqual(t, `{"name":"John Doe","age":30,"active":true}`, buf.String(), "json")

	buf.Reset().WriteStrings("new content")
	core.AssertEqual(t, "new content", buf.String(), "after reset")
}

// ── Benchmarks ────────────────────────────────────────────────────────

func BenchmarkBuffer_Chaining(b *testing.B) {
	b.ReportAllocs()
	i := 0
	for b.Loop() {
		var buf buffer.Buffer
		buf.WriteStrings("hello", " ", "world").
			Printf(" %d", i).
			WriteRunes('!')
		i++
	}
}

func BenchmarkBuffer_JSON(b *testing.B) {
	b.ReportAllocs()
	i := 0
	for b.Loop() {
		var buf buffer.Buffer
		buf.WriteStrings(`{"id":"`).
			Printf("%d", i).
			WriteStrings(`","name":"item_`).
			Printf("%d", i).
			WriteStrings(`","value":`).
			Printf("%d", i*10).
			WriteRunes('}')
		i++
	}
}

// ── Examples ──────────────────────────────────────────────────────────

func ExampleBuffer_chaining() {
	var buf buffer.Buffer

	buf.WriteStrings("Hello", " ").
		Print("World").
		Printf(" - %d", 2024)

	_, _ = fmt.Println(buf.String())
	// Output: Hello World - 2024
}

func ExampleBuffer_json() {
	var buf buffer.Buffer

	buf.WriteStrings(`{"name":"`).
		WriteStrings("Alice").
		WriteStrings(`","score":`).
		Printf("%d", 95).
		WriteRunes('}')

	_, _ = fmt.Println(buf.String())
	// Output: {"name":"Alice","score":95}
}
