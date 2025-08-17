package resource

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"darvaza.org/core"
	"darvaza.org/x/web/qlist"
)

const testJSONMediaType = "application/json;charset=utf-8"

// TestRendererFunc tests the RendererFunc type definition
func TestRendererFunc(t *testing.T) {
	// Test that RendererFunc can be assigned and called
	fn := func(rw http.ResponseWriter, _ *http.Request, code int, data string) error {
		rw.WriteHeader(code)
		_, err := rw.Write([]byte(data))
		return err
	}

	core.AssertNotNil(t, fn, "RendererFunc")

	// Test function call
	rw := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	err := fn(rw, req, http.StatusCreated, "test data")

	core.AssertNoError(t, err, "RendererFunc call")
	core.AssertEqual(t, http.StatusCreated, rw.Code, "status code")
	core.AssertEqual(t, "test data", rw.Body.String(), "response body")
}

// TestGetRendererWithCode tests the getRendererWithCode method
func TestGetRendererWithCode(t *testing.T) {
	t.Run("unregistered renderer", runTestGetRendererWithCodeUnregistered)
	t.Run("registered renderer", runTestGetRendererWithCodeRegistered)
}

func runTestGetRendererWithCodeUnregistered(t *testing.T) {
	r := &Resource[string]{
		rc: make(map[string]RendererFunc[string]),
	}
	fn := r.getRendererWithCode("application/json")
	core.AssertNil(t, fn, "unregistered renderer")
}

func runTestGetRendererWithCodeRegistered(t *testing.T) {
	r := &Resource[string]{
		rc: make(map[string]RendererFunc[string]),
	}
	r.rc["application/json"] = testRenderer

	fn := r.getRendererWithCode("application/json")
	core.AssertNotNil(t, fn, "registered renderer")
}

// TestRcFieldInitialization tests that the rc field is properly initialized
func TestRcFieldInitialization(t *testing.T) {
	t.Run("field initialized", runTestRcFieldInitializationField)
	t.Run("can add to map", runTestRcFieldInitializationAdd)
}

func runTestRcFieldInitializationField(t *testing.T) {
	r := newResource[string](nil)
	core.AssertNotNil(t, r.rc, "rc field initialized")
	core.AssertEqual(t, 0, len(r.rc), "rc field empty initially")
}

func runTestRcFieldInitializationAdd(t *testing.T) {
	r := newResource[string](nil)
	r.rc["test"] = testRenderer

	core.AssertEqual(t, 1, len(r.rc), "rc field after assignment")
}

// TestRendererFuncSignature ensures the RendererFunc signature is correct
func TestRendererFuncSignature(t *testing.T) {
	// This test ensures the RendererFunc type has the correct signature
	fn := func(rw http.ResponseWriter, _ *http.Request, code int, data int) error {
		core.AssertTrue(t, code >= 100 && code < 600, "valid status code %d", code)
		core.AssertTrue(t, data >= 0, "valid data parameter %d", data)
		rw.WriteHeader(code)
		_, err := rw.Write([]byte("test"))
		return err
	}

	rw := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)

	err := fn(rw, req, 200, 42)
	core.AssertNoError(t, err, "RendererFunc call")
}

// TestCase interface validation for data-driven tests
var _ core.TestCase = renderWithCodeTestCase{}

type renderWithCodeTestCase struct {
	acceptHeader    string
	expectedData    string
	name            string
	expectedCode    int
	hasCodeRenderer bool
	expectError     bool
}

func (tc renderWithCodeTestCase) Name() string {
	return tc.name
}

func (tc renderWithCodeTestCase) Test(t *testing.T) {
	t.Helper()

	r := &Resource[string]{
		rc: make(map[string]RendererFunc[string]),
		r:  make(map[string]HandlerFunc[string]),
	}

	// Set up quality list for JSON support
	qv, err := qlist.ParseQualityValue("application/json")
	core.AssertMustNoError(t, err, "parse quality value")
	r.ql = []qlist.QualityValue{qv}

	if tc.hasCodeRenderer {
		r.rc["application/json"] = func(rw http.ResponseWriter, _ *http.Request, code int, data string) error {
			rw.WriteHeader(code)
			_, err := rw.Write([]byte(data))
			return err
		}
	}

	rw := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Accept", tc.acceptHeader)

	err = r.RenderWithCode(rw, req, tc.expectedCode, tc.expectedData)

	if tc.expectError {
		core.AssertError(t, err, "RenderWithCode error")
		return
	}

	core.AssertNoError(t, err, "RenderWithCode")
	core.AssertEqual(t, tc.expectedCode, rw.Code, "status code")
	core.AssertEqual(t, tc.expectedData, rw.Body.String(), "response body")
}

// Factory function - success case
func newRenderWithCodeTestCaseSuccess(name, acceptHeader string,
	expectedCode int, expectedData string) renderWithCodeTestCase {
	return renderWithCodeTestCase{
		name:            name,
		acceptHeader:    acceptHeader,
		hasCodeRenderer: true,
		expectedCode:    expectedCode,
		expectedData:    expectedData,
		expectError:     false,
	}
}

// Factory function - error case
func newRenderWithCodeTestCaseError(name, acceptHeader string) renderWithCodeTestCase {
	return renderWithCodeTestCase{
		name:            name,
		acceptHeader:    acceptHeader,
		hasCodeRenderer: false,
		expectedCode:    0,
		expectedData:    "",
		expectError:     true,
	}
}

// TestRenderWithCode tests the RenderWithCode method
func TestRenderWithCode(t *testing.T) {
	t.Run("test cases", runTestRenderWithCodeCases)
	t.Run("empty rc map", runTestRenderWithCodeEmptyRC)
	t.Run("status codes", runTestRenderWithCodeStatusCodes)
}

func runTestRenderWithCodeCases(t *testing.T) {
	testCases := []renderWithCodeTestCase{
		newRenderWithCodeTestCaseSuccess("with code-aware renderer",
			"application/json", http.StatusCreated, "test data"),
		newRenderWithCodeTestCaseError("no code-aware renderer returns 406",
			"application/json"),
		newRenderWithCodeTestCaseError("invalid accept header",
			"invalid/header/format"),
		newRenderWithCodeTestCaseError("no acceptable media type",
			"application/xml"),
	}

	core.RunTestCases(t, testCases)
}

func runTestRenderWithCodeEmptyRC(t *testing.T) {
	r := &Resource[string]{
		rc: nil,
		r:  make(map[string]HandlerFunc[string]),
	}

	rw := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Accept", "application/json")

	err := r.RenderWithCode(rw, req, http.StatusOK, "test")
	core.AssertError(t, err, "RenderWithCode with nil rc")
}

func runTestRenderWithCodeStatusCodes(t *testing.T) {
	statusCodes := core.S(
		http.StatusOK,
		http.StatusCreated,
		http.StatusAccepted,
		http.StatusNoContent,
		http.StatusMovedPermanently,
		http.StatusBadRequest,
		http.StatusNotFound,
		http.StatusInternalServerError,
	)

	for _, code := range statusCodes {
		t.Run(http.StatusText(code), func(t *testing.T) {
			testStatusCodeRendering(t, code)
		})
	}
}

func testStatusCodeRendering(t *testing.T, expectedCode int) {
	t.Helper()

	r := &Resource[string]{
		rc: make(map[string]RendererFunc[string]),
	}

	// Set up quality list for JSON support
	qv, err := qlist.ParseQualityValue("application/json")
	core.AssertMustNoError(t, err, "parse quality value")
	r.ql = []qlist.QualityValue{qv}

	r.rc["application/json"] = func(rw http.ResponseWriter, _ *http.Request, receivedCode int, _ string) error {
		core.AssertEqual(t, expectedCode, receivedCode, "received status code")
		rw.WriteHeader(receivedCode)
		return nil
	}

	rw := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Accept", "application/json")

	err = r.RenderWithCode(rw, req, expectedCode, "test")
	core.AssertNoError(t, err, "RenderWithCode")
	core.AssertEqual(t, expectedCode, rw.Code, "response status code")
}

// TestJSONInterfaces tests the JSON renderer interfaces
func TestJSONInterfaces(t *testing.T) {
	t.Run("JSONRenderer interface", runTestJSONRenderer)
	t.Run("JSONRendererWithCode interface", runTestJSONRendererWithCode)
	t.Run("auto-detection prioritizes code-aware", runTestJSONAutoDetectionPriority)
}

func runTestJSONRenderer(t *testing.T) {
	// Test resource implementing legacy JSONRenderer
	resource := &testJSONResource{data: "legacy json test"}

	r := &Resource[string]{
		r: make(map[string]HandlerFunc[string]),
	}
	trySetJSONForResource(r, resource)

	// Should use legacy renderer (without code parameter)
	core.AssertNil(t, r.rc, "no code-aware renderer should be set")
	core.AssertNotNil(t, r.r, "legacy renderer should be set")

	// Check if any JSON renderer was registered
	// (the key is normalized without space after semicolon)
	found := false
	var actualKey string
	for key := range r.r {
		if key == testJSONMediaType {
			found = true
			actualKey = key
			break
		}
	}
	core.AssertTrue(t, found, "JSON legacy renderer should be registered")

	// Test that the renderer works
	rw := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	err := r.r[actualKey](rw, req, "test")
	core.AssertNoError(t, err, "legacy renderer call")
	core.AssertEqual(t, "LEGACY: test", rw.Body.String(), "response from legacy renderer")
}

func runTestJSONRendererWithCode(t *testing.T) {
	// Test resource implementing code-aware JSONRendererWithCode
	resource := &testJSONResourceWithCode{data: "code-aware json test"}

	r := &Resource[string]{
		rc: make(map[string]RendererFunc[string]),
	}
	trySetJSONForResource(r, resource)

	// Should use code-aware renderer
	core.AssertNotNil(t, r.rc, "code-aware renderer map should exist")

	// Check if any JSON renderer was registered in code-aware map
	found := false
	var actualKey string
	for key := range r.rc {
		if key == testJSONMediaType {
			found = true
			actualKey = key
			break
		}
	}
	core.AssertTrue(t, found, "JSON code-aware renderer should be registered")

	// Test that the renderer works
	rw := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	err := r.rc[actualKey](rw, req, http.StatusCreated, "test")
	core.AssertNoError(t, err, "code-aware renderer call")
	core.AssertEqual(t, http.StatusCreated, rw.Code, "status code from code-aware renderer")
	core.AssertEqual(t, "CODE_AWARE: test", rw.Body.String(), "response from code-aware renderer")
}

func runTestJSONAutoDetectionPriority(t *testing.T) {
	// Test that auto-detection correctly chooses code-aware over legacy
	// when code-aware interface is detected
	codeAwareResource := &testJSONResourceWithCode{data: "code-aware test"}

	r := &Resource[string]{
		rc: make(map[string]RendererFunc[string]),
		r:  make(map[string]HandlerFunc[string]),
	}
	trySetJSONForResource(r, codeAwareResource)

	// Should use code-aware renderer
	core.AssertNotNil(t, r.rc, "code-aware renderer map should exist")

	// Find the code-aware renderer
	var codeAwareRenderer RendererFunc[string]
	for key, renderer := range r.rc {
		if key == testJSONMediaType {
			codeAwareRenderer = renderer
			break
		}
	}
	core.AssertNotNil(t, codeAwareRenderer, "JSON code-aware renderer should be registered")

	// Legacy renderer should not be set
	_, hasLegacy := r.r[testJSONMediaType]
	core.AssertFalse(t, hasLegacy, "legacy renderer should not be set")

	// Test that it actually calls the code-aware version
	rw := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)

	err := codeAwareRenderer(rw, req, http.StatusCreated, "test")
	core.AssertNoError(t, err, "code-aware renderer call")
	core.AssertEqual(t, http.StatusCreated, rw.Code, "status code from code-aware renderer")
	core.AssertEqual(t, "CODE_AWARE: test", rw.Body.String(), "response from code-aware renderer")
}

// Test resource types for JSON interface testing

type testJSONResource struct {
	data string
}

func (*testJSONResource) RenderJSON(rw http.ResponseWriter, _ *http.Request, data string) error {
	rw.Header().Set("Content-Type", "application/json")
	_, err := rw.Write([]byte("LEGACY: " + data))
	return err
}

type testJSONResourceWithCode struct {
	data string
}

func (*testJSONResourceWithCode) RenderJSON(rw http.ResponseWriter, _ *http.Request, code int, data string) error {
	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(code)
	_, err := rw.Write([]byte("CODE_AWARE: " + data))
	return err
}

// Note: Go doesn't support method overloading, so we can't have a single type
// that implements both JSONRenderer and JSONRendererWithCode interfaces.
// The auto-detection logic checks the code-aware interface first, so we only
// need to test that priority works correctly with separate types.
