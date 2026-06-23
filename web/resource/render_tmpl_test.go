package resource

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"text/template"

	"darvaza.org/core"
	"darvaza.org/x/web/consts"
)

// TestCase interface validations for template renderer test cases
var _ core.TestCase = renderTemplateTestCase{}

// Test interface validations - compile-time verification
var _ TemplateRenderer[string] = (*testTemplateRenderer)(nil)
var _ TemplateRendererWithCode[string] = (*testTemplateRendererWithCode)(nil)

type renderTemplateTestCase struct {
	data           any
	expectedBody   string
	templateString string
	name           string
	method         string
	inputCode      int
	expectedCode   int
	expectError    bool
	nilTemplate    bool
}

func (tc renderTemplateTestCase) Name() string {
	return tc.name
}

func (tc renderTemplateTestCase) Test(t *testing.T) {
	t.Helper()

	rw := httptest.NewRecorder()
	req := httptest.NewRequest(tc.method, "/", http.NoBody)

	var tmpl *template.Template
	if !tc.nilTemplate {
		var err error
		tmpl, err = template.New("test").Parse(tc.templateString)
		core.AssertMustNoError(t, err, "template parse")
	}

	err := RenderTemplate(rw, req, tc.inputCode, tmpl, tc.data)

	if tc.expectError {
		core.AssertError(t, err, "RenderTemplate error")
		return
	}

	core.AssertNoError(t, err, "RenderTemplate")
	core.AssertEqual(t, tc.expectedCode, rw.Code, "status code")
	core.AssertEqual(t, tc.expectedBody, rw.Body.String(), "response body")

	// The implementation writes only the status line for HEAD, so
	// it emits no Content-Length; assert its absence. For the other
	// methods Content-Length must match the body length.
	if tc.method == consts.HEAD {
		core.AssertEqual(t, "", rw.Header().Get(consts.ContentLength), "content length")
	} else {
		contentLength := rw.Header().Get(consts.ContentLength)
		expectedLength := len(tc.expectedBody)
		core.AssertEqual(t, fmt.Sprintf("%d", expectedLength), contentLength, "content length")
	}
}

// Factory functions for test cases
func newRenderTemplateTestCase(name, templateString string, data any, expectedBody string,
	code int) renderTemplateTestCase {
	return renderTemplateTestCase{
		name:           name,
		templateString: templateString,
		data:           data,
		expectedBody:   expectedBody,
		inputCode:      code,
		expectedCode:   code,
		method:         consts.GET,
		expectError:    false,
		nilTemplate:    false,
	}
}

// newRenderTemplateTestCaseHead builds a HEAD-method case. HEAD
// responses carry no body, so the template and data are immaterial
// fixtures; only the status mapping matters, and RenderTemplate may
// adjust it (e.g. 200 → 204), so inputCode and expectedCode are
// supplied independently.
func newRenderTemplateTestCaseHead(name string, inputCode,
	expectedCode int) renderTemplateTestCase {
	return renderTemplateTestCase{
		name:           name,
		templateString: "{{.}}",
		data:           "head",
		expectedBody:   "",
		inputCode:      inputCode,
		expectedCode:   expectedCode,
		method:         consts.HEAD,
		expectError:    false,
		nilTemplate:    false,
	}
}

// newRenderTemplateTestCaseNilTemplate builds a GET case whose template
// is nil, so RenderTemplate fails before any rendering.
func newRenderTemplateTestCaseNilTemplate(name string) renderTemplateTestCase {
	return renderTemplateTestCase{
		name:           name,
		templateString: "",
		data:           "",
		expectedBody:   "",
		inputCode:      http.StatusOK,
		expectedCode:   http.StatusOK,
		method:         consts.GET,
		expectError:    true,
		nilTemplate:    true,
	}
}

// newRenderTemplateTestCaseExecError builds a GET case whose template
// parses but fails during execution, so the error surfaces through
// RenderTemplate rather than from a nil template.
func newRenderTemplateTestCaseExecError(name, templateString string,
	data any) renderTemplateTestCase {
	return renderTemplateTestCase{
		name:           name,
		templateString: templateString,
		data:           data,
		expectedBody:   "",
		inputCode:      http.StatusOK,
		expectedCode:   http.StatusOK,
		method:         consts.GET,
		expectError:    true,
		nilTemplate:    false,
	}
}

//revive:disable-next-line:argument-limit
func newRenderTemplateTestCaseNormalization(name, templateString string, data any,
	expectedBody string, inputCode, expectedCode int) renderTemplateTestCase {
	return renderTemplateTestCase{
		name:           name,
		templateString: templateString,
		data:           data,
		expectedBody:   expectedBody,
		inputCode:      inputCode,
		expectedCode:   expectedCode,
		method:         consts.GET,
		expectError:    false,
		nilTemplate:    false,
	}
}

func renderTemplateTestCases() []renderTemplateTestCase {
	return []renderTemplateTestCase{
		newRenderTemplateTestCase("simple template", "Hello {{.}}", "World",
			"Hello World", http.StatusOK),
		newRenderTemplateTestCase("status created", "Data: {{.}}", "test",
			"Data: test", http.StatusCreated),
		newRenderTemplateTestCase("status accepted", "{{.}}", "accepted",
			"accepted", http.StatusAccepted),
		// GET keeps a non-2xx status and still renders the body — the
		// mirror of "HEAD keeps not found", which suppresses it.
		newRenderTemplateTestCase("status not found renders body", "{{.}}",
			"missing", "missing", http.StatusNotFound),
		newRenderTemplateTestCase("complex template",
			"Name: {{.Name}}, Age: {{.Age}}", testData{Name: "John", Age: 25},
			"Name: John, Age: 25", http.StatusOK),
		newRenderTemplateTestCaseNormalization("zero status code defaults to 200",
			"{{.}}", "default", "default", 0, http.StatusOK),
		newRenderTemplateTestCaseNormalization("negative status code becomes 500",
			"{{.}}", "error", "error", -1, http.StatusInternalServerError),
		// HEAD adjusts a 200 to 204 and leaves every other status
		// untouched, always with an empty body. Normalisation runs
		// first, so a 0 reaches the adjustment as 200 and becomes 204,
		// while a negative settles at 500 and is left alone.
		newRenderTemplateTestCaseHead("HEAD adjusts 200 to 204",
			http.StatusOK, http.StatusNoContent),
		newRenderTemplateTestCaseHead("HEAD normalises 0 then adjusts to 204",
			0, http.StatusNoContent),
		newRenderTemplateTestCaseHead("HEAD negative becomes 500",
			-1, http.StatusInternalServerError),
		newRenderTemplateTestCaseHead("HEAD keeps 204",
			http.StatusNoContent, http.StatusNoContent),
		newRenderTemplateTestCaseHead("HEAD keeps created",
			http.StatusCreated, http.StatusCreated),
		newRenderTemplateTestCaseHead("HEAD keeps accepted",
			http.StatusAccepted, http.StatusAccepted),
		newRenderTemplateTestCaseHead("HEAD keeps not found",
			http.StatusNotFound, http.StatusNotFound),
		newRenderTemplateTestCaseNilTemplate("nil template"),
		newRenderTemplateTestCaseExecError("template execution error",
			"{{.NonExistentField}}", "simple string"),
	}
}

// TestRenderTemplate tests the RenderTemplate function
func TestRenderTemplate(t *testing.T) {
	core.RunTestCases(t, renderTemplateTestCases())
}

// TestDoRenderTemplate tests the doRenderTemplate function directly
func TestDoRenderTemplate(t *testing.T) {
	t.Run("nil template", runTestDoRenderTemplateNil)
	t.Run("template execution error", runTestDoRenderTemplateExecutionError)
	t.Run("successful render", runTestDoRenderTemplateSuccess)
}

func runTestDoRenderTemplateNil(t *testing.T) {
	rw := httptest.NewRecorder()
	err := doRenderTemplate(rw, nil, http.StatusOK, "test")

	core.AssertError(t, err, "nil template error")
	core.AssertErrorIs(t, err, core.ErrInvalid, "error type")
}

func runTestDoRenderTemplateExecutionError(t *testing.T) {
	// Create a template that will fail during execution
	tmpl, err := template.New("test").Parse("{{.NonExistentField}}")
	core.AssertMustNoError(t, err, "template parse")

	rw := httptest.NewRecorder()
	err = doRenderTemplate(rw, tmpl, http.StatusOK, "simple string")

	core.AssertError(t, err, "template execution error")
}

func runTestDoRenderTemplateSuccess(t *testing.T) {
	tmpl, err := template.New("test").Parse("Hello {{.}}")
	core.AssertMustNoError(t, err, "template parse")

	rw := httptest.NewRecorder()
	err = doRenderTemplate(rw, tmpl, http.StatusAccepted, "World")

	core.AssertNoError(t, err, "doRenderTemplate")
	core.AssertEqual(t, http.StatusAccepted, rw.Code, "status code")
	core.AssertEqual(t, "Hello World", rw.Body.String(), "response body")
	core.AssertEqual(t, "11", rw.Header().Get(consts.ContentLength), "content length")
}

// TestWithTemplate tests the WithTemplate helper function
func TestWithTemplate(t *testing.T) {
	t.Run("creates option function", runTestWithTemplateCreatesOption)
	t.Run("with custom media type", runTestWithTemplateCustomMediaType)
}

func runTestWithTemplateCreatesOption(t *testing.T) {
	fn := func(rw http.ResponseWriter, _ *http.Request, code int, data string) error {
		rw.WriteHeader(code)
		_, err := rw.Write([]byte("template: " + data))
		return err
	}

	option := WithTemplate("text/csv", fn)
	core.AssertNotNil(t, option, "WithTemplate returns option")

	// Test that the option can be applied to a resource
	r := newResource[string](nil)
	err := option(r)
	core.AssertNoError(t, err, "option application")

	// Verify the renderer was registered
	renderer := r.getRendererWithCode("text/csv")
	core.AssertNotNil(t, renderer, "renderer registered")

	// Test the renderer works
	rw := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", http.NoBody)
	err = renderer(rw, req, http.StatusOK, "test")

	core.AssertNoError(t, err, "renderer execution")
	core.AssertEqual(t, "template: test", rw.Body.String(), "renderer output")
}

func runTestWithTemplateCustomMediaType(t *testing.T) {
	fn := func(rw http.ResponseWriter, _ *http.Request, code int, data string) error {
		rw.Header().Set("Content-Type", "application/xml")
		rw.WriteHeader(code)
		_, err := rw.Write([]byte("<data>" + data + "</data>"))
		return err
	}

	option := WithTemplate("application/xml", fn)
	r := newResource[string](nil)
	err := option(r)
	core.AssertNoError(t, err, "option application")

	renderer := r.getRendererWithCode("application/xml")
	core.AssertNotNil(t, renderer, "XML renderer registered")

	rw := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", http.NoBody)
	err = renderer(rw, req, http.StatusOK, "test")

	core.AssertNoError(t, err, "XML renderer execution")
	core.AssertEqual(t, "<data>test</data>", rw.Body.String(), "XML output")
	core.AssertEqual(t, "application/xml", rw.Header().Get("Content-Type"), "content type")
}

// TestTemplateInterfaces tests the template renderer interfaces
func TestTemplateInterfaces(t *testing.T) {
	t.Run("TemplateRenderer interface", runTestTemplateRenderer)
	t.Run("TemplateRendererWithCode interface", runTestTemplateRendererWithCode)
}

func runTestTemplateRenderer(t *testing.T) {
	renderer := &testTemplateRenderer{data: "legacy template test"}

	rw := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", http.NoBody)

	err := renderer.RenderTemplate(rw, req, "test data")
	core.AssertNoError(t, err, "legacy template renderer")
	core.AssertEqual(t, "LEGACY_TEMPLATE: test data", rw.Body.String(), "legacy output")
}

func runTestTemplateRendererWithCode(t *testing.T) {
	renderer := &testTemplateRendererWithCode{data: "code-aware template test"}

	rw := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", http.NoBody)

	err := renderer.RenderTemplate(rw, req, http.StatusCreated, "test data")
	core.AssertNoError(t, err, "code-aware template renderer")
	core.AssertEqual(t, http.StatusCreated, rw.Code, "status code")
	core.AssertEqual(t, "CODE_TEMPLATE: test data", rw.Body.String(), "code-aware output")
}

// Test data types and helper implementations

type testData struct {
	Name string
	Age  int
}

type testTemplateRenderer struct {
	data string
}

func (*testTemplateRenderer) RenderTemplate(rw http.ResponseWriter, _ *http.Request, data string) error {
	rw.Header().Set("Content-Type", "text/plain")
	_, err := rw.Write([]byte("LEGACY_TEMPLATE: " + data))
	return err
}

type testTemplateRendererWithCode struct {
	data string
}

func (*testTemplateRendererWithCode) RenderTemplate(rw http.ResponseWriter, _ *http.Request, code int,
	data string) error {
	rw.Header().Set("Content-Type", "text/plain")
	rw.WriteHeader(code)
	_, err := rw.Write([]byte("CODE_TEMPLATE: " + data))
	return err
}
