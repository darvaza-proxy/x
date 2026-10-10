package web_test

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"darvaza.org/core"

	"darvaza.org/x/web"
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
	factory      func() *web.HTTPError
	name         string
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

func newBasicStatusTestCase(name string, factory func() *web.HTTPError,
	expectedCode int) basicStatusTestCase {
	return basicStatusTestCase{
		name:         name,
		factory:      factory,
		expectedCode: expectedCode,
	}
}

// wrapperStatusTestCase tests NewStatus* functions that wrap an error
type wrapperStatusTestCase struct {
	inputErr     error
	factory      func(error) *web.HTTPError
	name         string
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

func newWrapperStatusTestCase(name string, factory func(error) *web.HTTPError,
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
	factory        func(time.Duration) *web.HTTPError
	name           string
	expectedHeader string
	retryAfter     time.Duration
	expectedCode   int
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

func newRetryStatusTestCase(name string, factory func(time.Duration) *web.HTTPError,
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
		newBasicStatusTestCase("NotModified", web.NewStatusNotModified, http.StatusNotModified),
		newBasicStatusTestCase("Unauthorized", web.NewStatusUnauthorized, http.StatusUnauthorized),
		newBasicStatusTestCase("Forbidden", web.NewStatusForbidden, http.StatusForbidden),
		newBasicStatusTestCase("NotFound", web.NewStatusNotFound, http.StatusNotFound),
		newBasicStatusTestCase("NotAcceptable", web.NewStatusNotAcceptable, http.StatusNotAcceptable),
		newBasicStatusTestCase("Conflict", web.NewStatusConflict, http.StatusConflict),
		newBasicStatusTestCase("Gone", web.NewStatusGone, http.StatusGone),
		newBasicStatusTestCase("PreconditionFailed", web.NewStatusPreconditionFailed, http.StatusPreconditionFailed),
		newBasicStatusTestCase("NotImplemented", web.NewStatusNotImplemented, http.StatusNotImplemented),
		newBasicStatusTestCase("GatewayTimeout", web.NewStatusGatewayTimeout, http.StatusGatewayTimeout),
	}

	core.RunTestCases(t, testCases)
}

func TestWrapperStatusHelpers(t *testing.T) {
	testErr := errors.New("test error")

	testCases := []wrapperStatusTestCase{
		newWrapperStatusTestCase("BadRequest with error",
			web.NewStatusBadRequest, testErr, http.StatusBadRequest),
		newWrapperStatusTestCase("BadRequest with nil",
			web.NewStatusBadRequest, nil, http.StatusBadRequest),
		newWrapperStatusTestCase("UnsupportedMediaType with error",
			web.NewStatusUnsupportedMediaType, testErr, http.StatusUnsupportedMediaType),
		newWrapperStatusTestCase("UnprocessableEntity with error",
			web.NewStatusUnprocessableEntity, testErr, http.StatusUnprocessableEntity),
		newWrapperStatusTestCase("InternalServerError with error",
			web.NewStatusInternalServerError, testErr, http.StatusInternalServerError),
		newWrapperStatusTestCase("BadGateway with error",
			web.NewStatusBadGateway, testErr, http.StatusBadGateway),
	}

	core.RunTestCases(t, testCases)
}

func TestRetryStatusHelpers(t *testing.T) {
	testCases := []retryStatusTestCase{
		newRetryStatusTestCase("TooManyRequests 60 seconds",
			web.NewStatusTooManyRequests, 60*time.Second, http.StatusTooManyRequests, "60"),
		newRetryStatusTestCase("TooManyRequests 1 minute",
			web.NewStatusTooManyRequests, 1*time.Minute, http.StatusTooManyRequests, "60"),
		newRetryStatusTestCase("TooManyRequests rounds up",
			web.NewStatusTooManyRequests, 500*time.Millisecond, http.StatusTooManyRequests, "1"),
		newRetryStatusTestCase("TooManyRequests zero",
			web.NewStatusTooManyRequests, 0, http.StatusTooManyRequests, "0"),
		newRetryStatusTestCase("TooManyRequests negative",
			web.NewStatusTooManyRequests, -10*time.Second, http.StatusTooManyRequests, "0"),
		newRetryStatusTestCase("ServiceUnavailable 120 seconds",
			web.NewStatusServiceUnavailable, 120*time.Second, http.StatusServiceUnavailable, "120"),
		newRetryStatusTestCase("ServiceUnavailable rounds up",
			web.NewStatusServiceUnavailable, 1500*time.Millisecond, http.StatusServiceUnavailable, "2"),
	}

	core.RunTestCases(t, testCases)
}

// Test wrapper idempotency (wrapping HTTPError returns same error)
func TestWrapperIdempotency(t *testing.T) {
	t.Run("BadRequest", func(t *testing.T) {
		original := web.NewStatusBadRequest(errors.New("test"))
		wrapped := web.NewStatusBadRequest(original)
		core.AssertSame(t, original, wrapped, "same instance")
	})

	t.Run("UnsupportedMediaType", func(t *testing.T) {
		original := web.NewStatusUnsupportedMediaType(errors.New("test"))
		wrapped := web.NewStatusUnsupportedMediaType(original)
		core.AssertSame(t, original, wrapped, "same instance")
	})

	t.Run("UnprocessableEntity", func(t *testing.T) {
		original := web.NewStatusUnprocessableEntity(errors.New("test"))
		wrapped := web.NewStatusUnprocessableEntity(original)
		core.AssertSame(t, original, wrapped, "same instance")
	})

	t.Run("InternalServerError", func(t *testing.T) {
		original := web.NewStatusInternalServerError(errors.New("test"))
		wrapped := web.NewStatusInternalServerError(original)
		core.AssertSame(t, original, wrapped, "same instance")
	})

	t.Run("BadGateway", func(t *testing.T) {
		original := web.NewStatusBadGateway(errors.New("test"))
		wrapped := web.NewStatusBadGateway(original)
		core.AssertSame(t, original, wrapped, "same instance")
	})
}
