package resource

import (
	"net/http"
	"testing"

	"darvaza.org/core"
)

// TestResource tests Resource struct functionality
func TestResource(t *testing.T) {
	t.Run("rc field presence", runTestResourceRcFieldPresence)
	t.Run("rc field independence", runTestResourceRcFieldIndependence)
}

func runTestResourceRcFieldPresence(_ *testing.T) {
	var r Resource[string]
	// Test that rc field can be accessed without compilation error
	_ = r.rc // This is expected to be nil for zero value
}

func runTestResourceRcFieldIndependence(t *testing.T) {
	r := newResource[string](nil)

	// Add to rc map
	r.rc["test"] = testRenderer

	// Verify other maps are unaffected
	core.AssertEqual(t, 0, len(r.h), "h field remains empty")
	core.AssertEqual(t, 0, len(r.r), "r field remains empty")
	core.AssertEqual(t, 1, len(r.rc), "rc field has one item")
}

// TestNewResource tests newResource function
func TestNewResource(t *testing.T) {
	t.Run("all fields initialized", runTestNewResourceAllFieldsInitialized)
	t.Run("with handler", runTestNewResourceWithHandler)
}

func runTestNewResourceAllFieldsInitialized(t *testing.T) {
	r := newResource[string](nil)

	core.AssertNotNil(t, r.h, "h field initialized")
	core.AssertNotNil(t, r.r, "r field initialized")
	core.AssertNotNil(t, r.rc, "rc field initialized")

	// Test all maps are empty initially
	core.AssertEqual(t, 0, len(r.h), "h field empty initially")
	core.AssertEqual(t, 0, len(r.r), "r field empty initially")
	core.AssertEqual(t, 0, len(r.rc), "rc field empty initially")
}

func runTestNewResourceWithHandler(t *testing.T) {
	handler := &testHandler{}
	r := newResource[string](handler)

	core.AssertNotNil(t, r.rc, "rc field initialized with handler")
	core.AssertEqual(t, 0, len(r.rc), "rc field empty with handler")
}

// testHandler is a simple handler for testing
type testHandler struct{}

func (*testHandler) Get(_ http.ResponseWriter, _ *http.Request) error {
	return nil
}

// testRenderer is a test renderer function for code-aware rendering
func testRenderer(_ http.ResponseWriter, _ *http.Request, _ int, _ string) error {
	return nil
}

// writingRenderer is a test renderer that actually writes data
func writingRenderer(rw http.ResponseWriter, _ *http.Request, code int, data string) error {
	rw.WriteHeader(code)
	_, err := rw.Write([]byte(data))
	return err
}

// TestAddRendererWithCode tests the addRendererWithCode method
func TestAddRendererWithCode(t *testing.T) {
	t.Run("valid media type", runTestAddRendererWithCodeValidMediaType)
	t.Run("invalid media type", runTestAddRendererWithCodeInvalidMediaType)
	t.Run("nil renderer", runTestAddRendererWithCodeNilRenderer)
	t.Run("empty media type", runTestAddRendererWithCodeEmptyMediaType)
}

func runTestAddRendererWithCodeValidMediaType(t *testing.T) {
	r := newResource[string](nil)
	err := r.addRendererWithCode("application/json", testRenderer)
	core.AssertNoError(t, err, "add valid renderer")
	core.AssertNotNil(t, r.rc["application/json"], "renderer stored")
	core.AssertEqual(t, 1, len(r.ql), "quality list updated")
}

func runTestAddRendererWithCodeInvalidMediaType(t *testing.T) {
	r := newResource[string](nil)
	err := r.addRendererWithCode("invalid", testRenderer)
	core.AssertError(t, err, "invalid media type")
}

func runTestAddRendererWithCodeNilRenderer(t *testing.T) {
	r := newResource[string](nil)
	err := r.addRendererWithCode("text/plain", nil)
	core.AssertError(t, err, "nil renderer")
}

func runTestAddRendererWithCodeEmptyMediaType(t *testing.T) {
	r := newResource[string](nil)
	err := r.addRendererWithCode("", testRenderer)
	core.AssertError(t, err, "empty media type")
}

// TestAddRendererWithCodeQualityList tests quality list management
func TestAddRendererWithCodeQualityList(t *testing.T) {
	t.Run("add first renderer", runTestAddRendererWithCodeQualityListFirst)
	t.Run("replace same renderer", runTestAddRendererWithCodeQualityListReplace)
	t.Run("add different renderer", runTestAddRendererWithCodeQualityListDifferent)
}

func runTestAddRendererWithCodeQualityListFirst(t *testing.T) {
	r := newResource[string](nil)
	testRenderer1 := testRenderer
	err := r.addRendererWithCode("application/json", testRenderer1)
	core.AssertNoError(t, err, "add first renderer")
	core.AssertEqual(t, 1, len(r.ql), "quality list has one entry")
}

func runTestAddRendererWithCodeQualityListReplace(t *testing.T) {
	r := newResource[string](nil)
	testRenderer1 := testRenderer
	testRenderer2 := testRenderer
	_ = r.addRendererWithCode("application/json", testRenderer1)
	err := r.addRendererWithCode("application/json", testRenderer2)
	core.AssertNoError(t, err, "replace renderer")
	core.AssertEqual(t, 1, len(r.ql), "quality list still has one entry")
}

func runTestAddRendererWithCodeQualityListDifferent(t *testing.T) {
	r := newResource[string](nil)
	testRenderer1 := testRenderer
	_ = r.addRendererWithCode("application/json", testRenderer1)
	err := r.addRendererWithCode("text/plain", testRenderer1)
	core.AssertNoError(t, err, "add different renderer")
	core.AssertEqual(t, 2, len(r.ql), "quality list has two entries")
}

// TestAddRendererWithCodeCoexistWithLegacy tests coexistence with legacy renderers
func TestAddRendererWithCodeCoexistWithLegacy(t *testing.T) {
	t.Run("add legacy then code renderer", runTestAddRendererWithCodeCoexistWithLegacyAdd)
	t.Run("both renderers accessible", runTestAddRendererWithCodeCoexistWithLegacyAccess)
}

func runTestAddRendererWithCodeCoexistWithLegacyAdd(t *testing.T) {
	r := newResource[string](nil)
	legacyRenderer := func(_ http.ResponseWriter, _ *http.Request, _ string) error {
		return nil
	}
	codeRenderer := testRenderer

	err := r.addRenderer("application/json", legacyRenderer)
	core.AssertNoError(t, err, "add legacy renderer")
	core.AssertEqual(t, 1, len(r.ql), "quality list updated")

	err = r.addRendererWithCode("application/json", codeRenderer)
	core.AssertNoError(t, err, "add code renderer")
	core.AssertEqual(t, 1, len(r.ql), "quality list still has one entry")
}

func runTestAddRendererWithCodeCoexistWithLegacyAccess(t *testing.T) {
	r := newResource[string](nil)
	legacyRenderer := func(_ http.ResponseWriter, _ *http.Request, _ string) error {
		return nil
	}
	codeRenderer := testRenderer
	_ = r.addRenderer("application/json", legacyRenderer)
	_ = r.addRendererWithCode("application/json", codeRenderer)

	core.AssertNotNil(t, r.getRenderer("application/json"), "legacy renderer accessible")
	core.AssertNotNil(t, r.getRendererWithCode("application/json"), "code renderer accessible")
}

// TestWithRendererCode tests the WithRendererCode option function
func TestWithRendererCode(t *testing.T) {
	t.Run("successful option", runTestWithRendererCodeSuccessful)
	t.Run("busy resource", runTestWithRendererCodeBusy)
	t.Run("nil renderer", runTestWithRendererCodeNil)
	t.Run("invalid media type", runTestWithRendererCodeInvalid)
}

func runTestWithRendererCodeSuccessful(t *testing.T) {
	opt := WithRendererCode("application/json", testRenderer)
	r := newResource[string](nil)
	err := opt(r)
	core.AssertNoError(t, err, "WithRendererCode option")
	core.AssertNotNil(t, r.getRendererWithCode("application/json"), "renderer registered")
}

func runTestWithRendererCodeBusy(t *testing.T) {
	opt := WithRendererCode("application/json", testRenderer)
	r := newResource[string](nil)
	r.methods = []string{"GET"} // Make resource busy
	err := opt(r)
	core.AssertError(t, err, "busy resource")
}

func runTestWithRendererCodeNil(t *testing.T) {
	optNil := WithRendererCode[string]("application/json", nil)
	r2 := newResource[string](nil)
	err := optNil(r2)
	core.AssertError(t, err, "nil renderer")
}

func runTestWithRendererCodeInvalid(t *testing.T) {
	optInvalid := WithRendererCode("invalid", testRenderer)
	r3 := newResource[string](nil)
	err := optInvalid(r3)
	core.AssertError(t, err, "invalid media type")
}

// TestWithRendererCodeIntegration tests WithRendererCode with New function
func TestWithRendererCodeIntegration(t *testing.T) {
	t.Run("successful creation", runTestWithRendererCodeIntegrationSuccess)
	t.Run("error propagation", runTestWithRendererCodeIntegrationError)
}

func runTestWithRendererCodeIntegrationSuccess(t *testing.T) {
	resource, err := New[string](nil, WithRendererCode("application/json", writingRenderer))
	core.AssertNoError(t, err, "New with WithRendererCode")
	core.AssertNotNil(t, resource.getRendererWithCode("application/json"), "renderer available")
}

func runTestWithRendererCodeIntegrationError(t *testing.T) {
	_, err := New[string](nil, WithRendererCode("invalid", writingRenderer))
	core.AssertError(t, err, "New with invalid media type")
}
