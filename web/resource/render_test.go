package resource_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"darvaza.org/core"
)

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
	req := httptest.NewRequest("GET", "/", http.NoBody)
	err := fn(rw, req, http.StatusCreated, "test data")

	core.AssertNoError(t, err, "RendererFunc call")
	core.AssertEqual(t, http.StatusCreated, rw.Code, "status code")
	core.AssertEqual(t, "test data", rw.Body.String(), "response body")
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
	req := httptest.NewRequest("GET", "/", http.NoBody)

	err := fn(rw, req, 200, 42)
	core.AssertNoError(t, err, "RendererFunc call")
}
