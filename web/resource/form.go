package resource

import (
	"net/http"

	"darvaza.org/x/web/forms"
)

// ParseForm is similar to the standard request.ParseForm() but it
// handles urlencoded, multipart and JSON.
// For nested JSON objects ParseForm uses dots to join keys.
func (*Resource[T]) ParseForm(req *http.Request, maxMemory int64) error {
	return forms.ParseForm(req, maxMemory)
}
