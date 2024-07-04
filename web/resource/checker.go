package resource

import (
	"net/http"

	"darvaza.org/x/web"
)

// Checker is a resource that knows to validate its requests
type Checker interface {
	Check(*http.Request) (*http.Request, error)
}

// TChecker is a resource that knows how validate its requests,
// and returns the relevant data
type TChecker[T any] interface {
	Check(*http.Request) (*http.Request, *T, error)
}

// CheckerFunc is the signature of a function that pre-validates
// requests and returns relevant data
type CheckerFunc[T any] func(*http.Request) (*http.Request, *T, error)

// DefaultChecker is happy with any request that can resolve a
// valid path but it doesn't do any alteration to the request
// or its context.
func DefaultChecker[T any](req *http.Request) (*http.Request, *T, error) {
	_, err := web.Resolve(req)
	if err != nil {
		return nil, nil, err
	}
	return req, nil, nil
}

func checkerOf[T any](x any) (CheckerFunc[T], bool) {
	switch v := x.(type) {
	case TChecker[T]:
		return v.Check, true
	case Checker:
		fn := func(req *http.Request) (*http.Request, *T, error) {
			req2, err := v.Check(req)
			return req2, nil, err
		}
		return fn, true
	default:
		return nil, false
	}
}
