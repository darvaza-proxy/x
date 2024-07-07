package resource

import (
	"bytes"
	"encoding/json"
	"net/http"
)

// RenderJSON encodes the data as JSON and sends it to the client after setting
// Content-Type and Content-Length.  For HEAD only Content-Type is set.
func RenderJSON(rw http.ResponseWriter, req *http.Request, data any) error {
	SetHeader(rw, ContentType, JSON)

	if req.Method == HEAD {
		// done
		return nil
	}

	b, err := json.Marshal(data)
	if err != nil {
		return err
	}

	SetHeader(rw, ContentLength, "%v", len(b))

	buf := bytes.NewBuffer(b)
	_, err = buf.WriteTo(rw)
	return err
}
