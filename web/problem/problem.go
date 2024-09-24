// Package problem implements [RFC7807]
//
// [RFC7807]: https://datatracker.ietf.org/doc/html/rfc7807
//
// [RFC3986]: https://datatracker.ietf.org/doc/html/rfc3986
package problem

import (
	"net/http"

	"darvaza.org/core"
	"darvaza.org/x/web"
)

const (
	// JSON contains the Content-Type for RFC7807 in JSON format
	JSON = "application/problem+json; charset=utf-8"
)

var (
	_ web.Error    = (*Problem)(nil)
	_ http.Handler = (*Problem)(nil)
	_ web.Handler  = (*Problem)(nil)
)

// Problem represents an error as described by RPC-7807.
type Problem struct {
	Hdr http.Header `json:"-"`

	// The Problem Details JSON Object [§3]
	//
	// [§3]: https://www.rfc-editor.org/rfc/rfc7807#section-3
	// [§3.1]: https://www.rfc-editor.org/rfc/rfc7807#section-3.1
	//
	// [...]
	// Note that this requires each of the sub-problems to be similar enough
	// to use the same HTTP status code.  If they do not, the 207 (Multi-
	// Status) [RFC4918] code could be used to encapsulate multiple status
	// messages.

	// Members of a Problem Details Object [§3.1]
	//
	// A problem details object can have the following members:

	// "type" (string) - A URI reference [RFC3986] that identifies the
	// problem type.  This specification encourages that, when
	// dereferenced, it provide human-readable documentation for the
	// problem type (e.g., using HTML [W3C.REC-html5-20141028]).  When
	// this member is not present, its value is assumed to be
	// "about:blank".
	//
	// Consumers MUST use the "type" string as the primary identifier for
	// the problem type; [...]
	// Consumers SHOULD NOT automatically dereference the type URI.
	Type string `json:"type,omitempty" default:"about:blank"`

	// "title" (string) - A short, human-readable summary of the problem
	// type.  It SHOULD NOT change from occurrence to occurrence of the
	// problem, except for purposes of localization (e.g., using
	// proactive content negotiation; see [RFC7231], Section 3.4).
	//
	// [...] the "title" string is advisory and included only
	// for users who are not aware of the semantics of the URI and do not
	// have the ability to discover them (e.g., offline log analysis).
	Title string `json:"title"`

	// "status" (number) - The HTTP status code ([RFC7231], Section 6)
	// generated by the origin server for this occurrence of the problem.
	//
	// The "status" member, if present, is only advisory; it conveys the
	// HTTP status code used for the convenience of the consumer.
	// Generators MUST use the same status code in the actual HTTP response,
	// to assure that generic HTTP software that does not understand this
	// format still behaves correctly.  See Section 5 for further caveats
	// regarding its use.
	//
	// Consumers can use the status member to determine what the original
	// status code used by the generator was, in cases where it has been
	// changed (e.g., by an intermediary or cache), and when message bodies
	// persist without HTTP information.  Generic HTTP software will still
	// use the HTTP status code.
	Status int `json:"status,omitempty" validate:"gt=0"`

	// "detail" (string) - A human-readable explanation specific to this
	// occurrence of the problem.
	//
	// The "detail" member, if present, ought to focus on helping the client
	// correct the problem, rather than giving debugging information.
	//
	// Consumers SHOULD NOT parse the "detail" member for information;
	// extensions are more suitable and less error-prone ways to obtain such
	// information.
	Detail string `json:"detail,omitempty"`

	// "instance" (string) - A URI reference that identifies the specific
	// occurrence of the problem.  It may or may not yield further
	// information if dereferenced.
	//
	// Note that both "type" and "instance" accept relative URIs; this means
	// that they must be resolved relative to the document's base URI, as
	// per [RFC3986], Section 5.
	Instance string `json:"instance,omitempty" validate:"uri"`

	// 400
	//
	InvalidParams []*InvalidParam `json:"invalid-params,omitempty"`
}

// New creates a new [Problem]
func New(problemStatus int, problemType, problemTitle string) *Problem {
	return &Problem{
		Status: problemStatus,
		Type:   problemType,
		Title:  problemTitle,
	}
}

// OK tells if there are no problems.
func (e *Problem) OK() bool {
	return !e.notOK()
}

func (e *Problem) notOK() bool {
	const ok = true

	switch {
	case e == nil:
		return ok
	case e.Status < 0, e.Status >= 300, len(e.Title) > 0, e.hasSubProblems():
		return !ok
	default:
		return ok
	}
}

// AsError returns itself as an standard error if there is
// a problem.
func (e *Problem) AsError() error {
	if e.notOK() {
		return e
	}
	return nil
}

// Unwrap returns all sub-problems
func (e *Problem) Unwrap() []error {
	if e != nil {
		return core.AsErrors(e.InvalidParams)
	}
	return nil
}

func (e *Problem) hasSubProblems() bool {
	for _, ep := range e.InvalidParams {
		if ep != nil {
			return true
		}
	}
	return false
}

// HTTPStatus returns the associated HTTP Status Code
func (e *Problem) HTTPStatus() int {
	switch {
	case e == nil, e.Status < 0:
		return http.StatusInternalServerError
	case e.Status != 0:
		return e.Status
	case e.OK():
		return http.StatusOK
	default:
		// it should have Status
		return http.StatusInternalServerError
	}
}

// Header returns the HTTP headers associated to the [Problem], if any.
func (e *Problem) Header() http.Header {
	if e != nil && len(e.Hdr) > 0 {
		return e.Hdr
	}
	return nil
}

// Error returns the title of the error.
func (e *Problem) Error() string {
	var code int

	switch {
	case e == nil:
		return "nil error receiver"
	case e.Title != "":
		return e.Title
	case e.Status < 0:
		code = http.StatusInternalServerError
	case e.Status != 0:
		code = e.Status
	case e.hasSubProblems():
		// it should have Status and Title!
		code = http.StatusInternalServerError
	default:
		code = http.StatusOK
	}

	return web.ErrorText(code)
}

func (e *Problem) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	ServeHTTP(e, rw, req)
}

// TryServeHTTP returns the Problem if it's an error, or render a 204 if it isn't.
func (e *Problem) TryServeHTTP(rw http.ResponseWriter, req *http.Request) error {
	return TryServeHTTP(e, rw, req)
}
