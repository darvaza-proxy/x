// Package consts contains constant strings used on HTTP handlers
package consts

import (
	"strings"
	"unicode"
)

const (
	// Accept is the canonical header name used for negotiating a MIME Type
	// for the content of the Request response
	Accept = "Accept"

	// AcceptEncoding is the canonical name given to the header used
	// to indicate compression options
	AcceptEncoding = "Accept-Encoding"

	// Allow is the canonical header used to indicate the supported
	// Methods.
	Allow = "Allow"

	// CacheControl is the canonical header used to specify how long
	// to cache the content.
	CacheControl = "Cache-Control"

	// ContentEncoding is the canonical Content-Encoding header
	ContentEncoding = "Content-Encoding"
	// ContentLength is the canonical Content-Length header
	ContentLength = "Content-Length"
	// ContentType is the canonical Content-Type header
	ContentType = "Content-Type"
	// ETag is the canonical ETag header
	ETag = "Etag"

	// Location is the canonical name given to the header used
	// to indicate a redirection.
	Location = "Location"
)

const (
	// GET represents the HTTP GET Method.
	GET = "GET"
	// HEAD represents the HTTP HEAD Method.
	HEAD = "HEAD"
	// POST represents the HTTP POST Method.
	POST = "POST"
	// PUT represents the HTTP PUT Method.
	PUT = "PUT"
	// DELETE represents the HTTP DELETE Method.
	DELETE = "DELETE"
	// CONNECT represents the HTTP CONNECT Method.
	CONNECT = "CONNECT"
	// OPTIONS represents the HTTP OPTIONS Method.
	OPTIONS = "OPTIONS"
	// TRACE represents the HTTP TRACE Method.
	TRACE = "TRACE"
	// PATCH represents the HTTP PATCH Method.
	PATCH = "PATCH"
)

const (
	// BIN is the standard Media Type for binary files.
	BIN = "application/octet-stream"
	// JSON is the standard Media Type for JSON content.
	JSON = "application/json; charset=utf-8"
	// HTML is the standard Media Type for HTML content.
	HTML = "text/html; charset=utf-8"
	// TXT is the standard Media Type for plain text content.
	TXT = "text/plain; charset=utf-8"

	// URLEncodedForm is the standard Media Type for HTML
	// forms by default.
	URLEncodedForm = "application/x-www-form-urlencoded"
	// MultiPartForm is the standard Media Type for HTML
	// forms that include attachments.
	MultiPartForm = "multipart/form-data"
)

// ContentTypeValue returns the first part of a Content-Type string
func ContentTypeValue(s string) string {
	i := strings.IndexFunc(s, func(r rune) bool {
		if r == ';' || unicode.IsSpace(r) {
			return true
		}
		return false
	})

	if i < 0 {
		return s
	}
	return s[:i]
}
