package resource

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"text/template"

	"darvaza.org/core"
	"darvaza.org/x/web/consts"
)

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
