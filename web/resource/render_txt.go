package resource

import "net/http"

func txtRendererOf[T any](x any) (HandlerFunc[T], bool) {
	if v, ok := x.(interface {
		RenderTXT(http.ResponseWriter, *http.Request, T) error
	}); ok {
		return v.RenderTXT, true
	}

	return nil, false
}
