package trace

import (
	"net/http"

	"github.com/go-chi/chi"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

type Builder struct {
	*chi.Mux
	Name string
}

var (
	excludeHTTPPath = map[string]struct{}{
		"/":            {},
		"/health":      {},
		"/favicon.ico": {},
	}
)

func (b *Builder) Named(name string) *Builder {
	return b
}

func (b *Builder) Build() http.Handler {
	return otelhttp.NewHandler(b.Mux, b.Name,
		otelhttp.WithMessageEvents(otelhttp.ReadEvents, otelhttp.WriteEvents),
		otelhttp.WithFilter(ShouldLogRequest),
	)
}

func ShouldLogRequest(r *http.Request) bool {
	_, ok := excludeHTTPPath[r.URL.Path]
	return !ok
}
