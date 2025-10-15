package web

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"darvaza.org/core"

	"darvaza.org/x/web/consts"
)

// TestCase interface validations
var (
	_ core.TestCase = basicStatusTestCase{}
	_ core.TestCase = wrapperStatusTestCase{}
	_ core.TestCase = retryStatusTestCase{}
)

// basicStatusTestCase tests NewStatus* functions that take no parameters
type basicStatusTestCase struct {
	name         string
	factory      func() *HTTPError
	expectedCode int
}

func (tc basicStatusTestCase) Name() string {
	return tc.name
}

func (tc basicStatusTestCase) Test(t *testing.T) {
	t.Helper()

	err := tc.factory()
	core.AssertNotNil(t, err, "HTTPError")
	core.AssertEqual(t, tc.expectedCode, err.HTTPStatus(), "status code")
	core.AssertNil(t, err.Err, "wrapped error")
	core.AssertNil(t, err.Hdr, "headers")
}

func newBasicStatusTestCase(name string, factory func() *HTTPError,
	expectedCode int) basicStatusTestCase {
	return basicStatusTestCase{
		name:         name,
		factory:      factory,
		expectedCode: expectedCode,
	}
}

// wrapperStatusTestCase tests NewStatus* functions that wrap an error
type wrapperStatusTestCase struct {
	name         string
	factory      func(error) *HTTPError
	inputErr     error
	expectedCode int
}

func (tc wrapperStatusTestCase) Name() string {
	return tc.name
}

func (tc wrapperStatusTestCase) Test(t *testing.T) {
	t.Helper()

	err := tc.factory(tc.inputErr)
	core.AssertNotNil(t, err, "HTTPError")
	core.AssertEqual(t, tc.expectedCode, err.HTTPStatus(), "status code")

	if tc.inputErr != nil {
		core.AssertNotNil(t, err.Err, "wrapped error")
		core.AssertErrorIs(t, err.Err, tc.inputErr, "error chain")
	}
}

func newWrapperStatusTestCase(name string, factory func(error) *HTTPError,
	inputErr error, expectedCode int) wrapperStatusTestCase {
	return wrapperStatusTestCase{
		name:         name,
		factory:      factory,
		inputErr:     inputErr,
		expectedCode: expectedCode,
	}
}

// retryStatusTestCase tests NewStatus* functions with Retry-After header
type retryStatusTestCase struct {
	name           string
	factory        func(time.Duration) *HTTPError
	retryAfter     time.Duration
	expectedCode   int
	expectedHeader string
}

func (tc retryStatusTestCase) Name() string {
	return tc.name
}

func (tc retryStatusTestCase) Test(t *testing.T) {
	t.Helper()

	err := tc.factory(tc.retryAfter)
	core.AssertNotNil(t, err, "HTTPError")
	core.AssertEqual(t, tc.expectedCode, err.HTTPStatus(), "status code")
	core.AssertNotNil(t, err.Hdr, "headers")

	retryAfter := err.Hdr.Get(consts.RetryAfter)
	core.AssertEqual(t, tc.expectedHeader, retryAfter, "Retry-After header")
}

func newRetryStatusTestCase(name string, factory func(time.Duration) *HTTPError,
	retryAfter time.Duration, expectedCode int,
	expectedHeader string) retryStatusTestCase {
	return retryStatusTestCase{
		name:           name,
		factory:        factory,
		retryAfter:     retryAfter,
		expectedCode:   expectedCode,
		expectedHeader: expectedHeader,
	}
}

// Test functions

func TestBasicStatusHelpers(t *testing.T) {
	testCases := []basicStatusTestCase{
		newBasicStatusTestCase("NotModified", NewStatusNotModified, http.StatusNotModified),
		newBasicStatusTestCase("Unauthorized", NewStatusUnauthorized, http.StatusUnauthorized),
		newBasicStatusTestCase("Forbidden", NewStatusForbidden, http.StatusForbidden),
		newBasicStatusTestCase("NotFound", NewStatusNotFound, http.StatusNotFound),
		newBasicStatusTestCase("NotAcceptable", NewStatusNotAcceptable, http.StatusNotAcceptable),
		newBasicStatusTestCase("Conflict", NewStatusConflict, http.StatusConflict),
		newBasicStatusTestCase("Gone", NewStatusGone, http.StatusGone),
		newBasicStatusTestCase("PreconditionFailed", NewStatusPreconditionFailed, http.StatusPreconditionFailed),
		newBasicStatusTestCase("NotImplemented", NewStatusNotImplemented, http.StatusNotImplemented),
		newBasicStatusTestCase("GatewayTimeout", NewStatusGatewayTimeout, http.StatusGatewayTimeout),
	}

	core.RunTestCases(t, testCases)
}

func TestWrapperStatusHelpers(t *testing.T) {
	testErr := errors.New("test error")

	testCases := []wrapperStatusTestCase{
		newWrapperStatusTestCase("BadRequest with error",
			NewStatusBadRequest, testErr, http.StatusBadRequest),
		newWrapperStatusTestCase("BadRequest with nil",
			NewStatusBadRequest, nil, http.StatusBadRequest),
		newWrapperStatusTestCase("UnsupportedMediaType with error",
			NewStatusUnsupportedMediaType, testErr, http.StatusUnsupportedMediaType),
		newWrapperStatusTestCase("UnprocessableEntity with error",
			NewStatusUnprocessableEntity, testErr, http.StatusUnprocessableEntity),
		newWrapperStatusTestCase("InternalServerError with error",
			NewStatusInternalServerError, testErr, http.StatusInternalServerError),
		newWrapperStatusTestCase("BadGateway with error",
			NewStatusBadGateway, testErr, http.StatusBadGateway),
	}

	core.RunTestCases(t, testCases)
}

func TestRetryStatusHelpers(t *testing.T) {
	testCases := []retryStatusTestCase{
		newRetryStatusTestCase("TooManyRequests 60 seconds",
			NewStatusTooManyRequests, 60*time.Second, http.StatusTooManyRequests, "60"),
		newRetryStatusTestCase("TooManyRequests 1 minute",
			NewStatusTooManyRequests, 1*time.Minute, http.StatusTooManyRequests, "60"),
		newRetryStatusTestCase("TooManyRequests rounds up",
			NewStatusTooManyRequests, 500*time.Millisecond, http.StatusTooManyRequests, "1"),
		newRetryStatusTestCase("TooManyRequests zero",
			NewStatusTooManyRequests, 0, http.StatusTooManyRequests, "0"),
		newRetryStatusTestCase("TooManyRequests negative",
			NewStatusTooManyRequests, -10*time.Second, http.StatusTooManyRequests, "0"),
		newRetryStatusTestCase("ServiceUnavailable 120 seconds",
			NewStatusServiceUnavailable, 120*time.Second, http.StatusServiceUnavailable, "120"),
		newRetryStatusTestCase("ServiceUnavailable rounds up",
			NewStatusServiceUnavailable, 1500*time.Millisecond, http.StatusServiceUnavailable, "2"),
	}

	core.RunTestCases(t, testCases)
}

// Test wrapper idempotency (wrapping HTTPError returns same error)
func TestWrapperIdempotency(t *testing.T) {
	t.Run("BadRequest", func(t *testing.T) {
		original := NewStatusBadRequest(errors.New("test"))
		wrapped := NewStatusBadRequest(original)
		core.AssertSame(t, original, wrapped, "same instance")
	})

	t.Run("UnsupportedMediaType", func(t *testing.T) {
		original := NewStatusUnsupportedMediaType(errors.New("test"))
		wrapped := NewStatusUnsupportedMediaType(original)
		core.AssertSame(t, original, wrapped, "same instance")
	})

	t.Run("UnprocessableEntity", func(t *testing.T) {
		original := NewStatusUnprocessableEntity(errors.New("test"))
		wrapped := NewStatusUnprocessableEntity(original)
		core.AssertSame(t, original, wrapped, "same instance")
	})

	t.Run("InternalServerError", func(t *testing.T) {
		original := NewStatusInternalServerError(errors.New("test"))
		wrapped := NewStatusInternalServerError(original)
		core.AssertSame(t, original, wrapped, "same instance")
	})

	t.Run("BadGateway", func(t *testing.T) {
		original := NewStatusBadGateway(errors.New("test"))
		wrapped := NewStatusBadGateway(original)
		core.AssertSame(t, original, wrapped, "same instance")
	})
}
