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
	expectedBody   string
	templateString string
	data           any
	name           string
	inputCode      int
	expectedCode   int
	method         string
	expectError    bool
	nilTemplate    bool
}

func (tc renderTemplateTestCase) Name() string {
	return tc.name
}

func (tc renderTemplateTestCase) Test(t *testing.T) {
	t.Helper()

	rw := httptest.NewRecorder()
	req := httptest.NewRequest(tc.method, "/", nil)

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

	if tc.method != consts.HEAD {
		core.AssertEqual(t, tc.expectedBody, rw.Body.String(), "response body")
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
		method:         "GET",
		expectError:    false,
		nilTemplate:    false,
	}
}

//revive:disable-next-line:argument-limit
func newRenderTemplateTestCaseMethod(name, method, templateString string, data any,
	expectedBody string, code int) renderTemplateTestCase {
	return renderTemplateTestCase{
		name:           name,
		templateString: templateString,
		data:           data,
		expectedBody:   expectedBody,
		inputCode:      code,
		expectedCode:   code,
		method:         method,
		expectError:    false,
		nilTemplate:    false,
	}
}

func newRenderTemplateTestCaseError(name string, nilTemplate bool) renderTemplateTestCase {
	return renderTemplateTestCase{
		name:           name,
		templateString: "",
		data:           "",
		expectedBody:   "",
		inputCode:      http.StatusOK,
		expectedCode:   http.StatusOK,
		method:         "GET",
		expectError:    true,
		nilTemplate:    nilTemplate,
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
		method:         "GET",
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
		newRenderTemplateTestCase("complex template",
			"Name: {{.Name}}, Age: {{.Age}}", testData{Name: "John", Age: 25},
			"Name: John, Age: 25", http.StatusOK),
		newRenderTemplateTestCaseNormalization("zero status code defaults to 200",
			"{{.}}", "default", "default", 0, http.StatusOK),
		newRenderTemplateTestCaseNormalization("negative status code becomes 500",
			"{{.}}", "error", "error", -1, http.StatusInternalServerError),
		newRenderTemplateTestCaseMethod("HEAD request", consts.HEAD,
			"{{.}}", "head", "", http.StatusNoContent),
		newRenderTemplateTestCaseMethod("HEAD with custom status", consts.HEAD,
			"{{.}}", "head", "", http.StatusAccepted),
		newRenderTemplateTestCaseError("nil template", true),
	}
}

// TestRenderTemplate tests the RenderTemplate function
func TestRenderTemplate(t *testing.T) {
	core.RunTestCases(t, renderTemplateTestCases())
}

// TestRenderTemplateStatusCodes tests various status code scenarios
func TestRenderTemplateStatusCodes(t *testing.T) {
	t.Run("status code normalization", runTestRenderTemplateStatusCodeNormalization)
	t.Run("head method status adjustment", runTestRenderTemplateHeadStatusAdjustment)
}

func runTestRenderTemplateStatusCodeNormalization(t *testing.T) {
	tmpl, err := template.New("test").Parse("{{.}}")
	core.AssertMustNoError(t, err, "template parse")

	testCases := []struct {
		name         string
		inputCode    int
		expectedCode int
	}{
		{"zero becomes 200", 0, http.StatusOK},
		{"negative becomes 500", -1, http.StatusInternalServerError},
		{"positive unchanged", http.StatusCreated, http.StatusCreated},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rw := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/", nil)

			err := RenderTemplate(rw, req, tc.inputCode, tmpl, "test")
			core.AssertNoError(t, err, "RenderTemplate")
			core.AssertEqual(t, tc.expectedCode, rw.Code, "status code")
		})
	}
}

func runTestRenderTemplateHeadStatusAdjustment(t *testing.T) {
	tmpl, err := template.New("test").Parse("{{.}}")
	core.AssertMustNoError(t, err, "template parse")

	testCases := []struct {
		name         string
		inputCode    int
		expectedCode int
	}{
		{"200 becomes 204", http.StatusOK, http.StatusNoContent},
		{"201 unchanged", http.StatusCreated, http.StatusCreated},
		{"404 unchanged", http.StatusNotFound, http.StatusNotFound},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rw := httptest.NewRecorder()
			req := httptest.NewRequest(consts.HEAD, "/", nil)

			err := RenderTemplate(rw, req, tc.inputCode, tmpl, "test")
			core.AssertNoError(t, err, "RenderTemplate")
			core.AssertEqual(t, tc.expectedCode, rw.Code, "status code")
			core.AssertEqual(t, "", rw.Body.String(), "empty body for HEAD")
		})
	}
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
	req := httptest.NewRequest("GET", "/", nil)
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
	req := httptest.NewRequest("GET", "/", nil)
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
	req := httptest.NewRequest("GET", "/", nil)

	err := renderer.RenderTemplate(rw, req, "test data")
	core.AssertNoError(t, err, "legacy template renderer")
	core.AssertEqual(t, "LEGACY_TEMPLATE: test data", rw.Body.String(), "legacy output")
}

func runTestTemplateRendererWithCode(t *testing.T) {
	renderer := &testTemplateRendererWithCode{data: "code-aware template test"}

	rw := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)

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
