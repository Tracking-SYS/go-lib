package http

import (
	"context"
	"net/http"

	"github.com/Tracking-SYS/go-lib/contextLib"
)

var (
	CtxUserMetadata   = contextLib.NewKey("user_metadata")
	CtxHTTPHeader     = contextLib.NewKey("header")
	CtxHTTPQueryParam = contextLib.NewKey("query_params")
	CtxDeviceInfo     = contextLib.NewKey("device_info")
	CtxWidgetType     = contextLib.NewKey("widget_type")
)

type RequestExtractorConf struct {
	fn  func(req *http.Request) interface{}
	key contextLib.Key
}
type MiddleWareFunc func(http.Handler) http.Handler

func NewRequestExtractorConf(
	key contextLib.Key,
	fn func(req *http.Request) interface{},
) *RequestExtractorConf {
	return &RequestExtractorConf{fn, key}
}

// NewHTTPRequestExtractor creates new value extractor from a *http.Request
// and attach into context
func NewRequestExtractor(exts ...*RequestExtractorConf) MiddleWareFunc {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			for _, ex := range exts {
				ctxValue := ex.fn(r)
				if ctxValue != nil {
					ctx = context.WithValue(ctx, ex.key, ctxValue)
				}
			}
			next.ServeHTTP(w, r.WithContext(ctx))
		}
		return http.HandlerFunc(fn)
	}
}
