package forms_test

import (
	"errors"
	"io"
	"net/url"
	"testing"

	"darvaza.org/core"

	"darvaza.org/x/web/forms"
)

var _ core.TestCase = formJSONTestCase{}

type formJSONTestCase struct {
	expected url.Values
	name     string
	input    string
	wantErr  bool
}

func (tc formJSONTestCase) Name() string {
	return tc.name
}

func (tc formJSONTestCase) Test(t *testing.T) {
	t.Helper()

	values, err := forms.UnmarshalFormJSON([]byte(tc.input))
	if tc.wantErr {
		core.AssertError(t, err, "error")
		core.AssertFalse(t, errors.Is(err, io.EOF), "io.EOF")
		core.AssertNil(t, values, "values")
		return
	}

	core.AssertNoError(t, err, "error")
	core.AssertDeepEqual(t, tc.expected, values, "values")
}

func newFormJSONTestCase(name, input string,
	expected url.Values) formJSONTestCase {
	return formJSONTestCase{
		name:     name,
		input:    input,
		expected: expected,
	}
}

// newFormJSONKeyTestCase expects a single key holding the given values,
// or no values at all.
func newFormJSONKeyTestCase(name, input, key string,
	values ...string) formJSONTestCase {
	return newFormJSONTestCase(name, input, url.Values{
		key: append([]string{}, values...),
	})
}

func newFormJSONErrorTestCase(name, input string) formJSONTestCase {
	return formJSONTestCase{
		name:    name,
		input:   input,
		wantErr: true,
	}
}

func formJSONTestCases() []formJSONTestCase {
	return core.S(
		newFormJSONTestCase("empty object", `{}`, url.Values{}),
		newFormJSONTestCase("null document", `null`, url.Values{}),
		newFormJSONTestCase("several keys", `{"a":"x","b":"y"}`,
			url.Values{"a": {"x"}, "b": {"y"}}),
		newFormJSONKeyTestCase("trailing whitespace",
			"{\"a\":\"x\"} \n", "a", "x"),

		newFormJSONKeyTestCase("string", `{"a":"x"}`, "a", "x"),
		newFormJSONKeyTestCase("empty string", `{"a":""}`, "a", ""),
		newFormJSONKeyTestCase("true", `{"a":true}`, "a", "true"),
		newFormJSONKeyTestCase("false", `{"a":false}`, "a", "false"),
		newFormJSONKeyTestCase("null", `{"a":null}`, "a"),

		newFormJSONKeyTestCase("integer", `{"a":42}`, "a", "42"),
		newFormJSONKeyTestCase("negative decimal",
			`{"a":-1.5}`, "a", "-1.5"),

		newFormJSONKeyTestCase("array",
			`{"a":["x","y"]}`, "a", "x", "y"),
		newFormJSONTestCase("empty array", `{"a":[]}`, url.Values{}),
		newFormJSONKeyTestCase("array of mixed types",
			`{"a":["x",1,true]}`, "a", "x", "1", "true"),
		newFormJSONKeyTestCase("nested arrays",
			`{"a":[["x"],["y"]]}`, "a", "x", "y"),
		newFormJSONKeyTestCase("null before a value",
			`{"a":[null,"x"]}`, "a", "x"),
		newFormJSONKeyTestCase("null after a value",
			`{"a":["x",null]}`, "a", "x"),

		newFormJSONKeyTestCase("object", `{"a":{"b":"x"}}`, "a.b", "x"),
		newFormJSONTestCase("object of several keys",
			`{"a":{"b":"x","c":"y"}}`,
			url.Values{"a.b": {"x"}, "a.c": {"y"}}),
		newFormJSONKeyTestCase("deep object",
			`{"a":{"b":{"c":"x"}}}`, "a.b.c", "x"),
		newFormJSONTestCase("empty object value",
			`{"a":{}}`, url.Values{}),
		newFormJSONKeyTestCase("array of objects",
			`{"a":[{"b":"x"},{"b":"y"}]}`, "a.b", "x", "y"),

		newFormJSONErrorTestCase("empty input", ``),
		newFormJSONErrorTestCase("truncated object", `{"a":`),
		newFormJSONErrorTestCase("second document", `{"a":"x"} {}`),
		newFormJSONErrorTestCase("trailing brace", `{"a":"x"}}`),
		newFormJSONErrorTestCase("array document", `["x"]`),
		newFormJSONErrorTestCase("string document", `"x"`),
	)
}

func TestUnmarshalFormJSON(t *testing.T) {
	core.RunTestCases(t, formJSONTestCases())
}
