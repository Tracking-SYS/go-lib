package http

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/middleware"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/global"
)

const (
	httpLatencyName       = "http.server.requests.duration.seconds"
	httpRequestNumberName = "http.server.requests.number"
	// externalRequestsTotal   = "http.external.requests.total"
	// externalRequestsLatency = "http.external.requests.duration.seconds"
)

const (
	httpCodeLabel   = attribute.Key("code")
	httpMethodLabel = attribute.Key("method")
	httpPathLabel   = attribute.Key("path")
	httpHostLabel   = attribute.Key("host")
)

type httpMetricsObserver struct {
	latency       metric.Float64ValueRecorder
	meter         metric.Meter
	total_request metric.Int64Counter
	reqFilter     otelhttp.Filter
}

func (c httpMetricsObserver) handler(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		if c.reqFilter != nil && !c.reqFilter(r) {
			next.ServeHTTP(w, r)
			return
		}

		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		start := time.Now()
		next.ServeHTTP(ww, r)
		c.latency.Record(
			r.Context(),
			time.Since(start).Seconds(),
			httpCodeLabel.Int(ww.Status()),
			httpMethodLabel.String(r.Method),
			httpPathLabel.String(r.URL.Path),
			httpHostLabel.String(r.URL.Host),
		)

		c.total_request.Add(r.Context(),
			1,
			httpCodeLabel.Int(ww.Status()),
			httpMethodLabel.String(r.Method),
			httpPathLabel.String(r.URL.Path),
			httpHostLabel.String(r.URL.Host),
		)
	}
	return http.HandlerFunc(fn)
}

// NewHTTPMetricsObserver returns a new otel HTTPMiddleware handler.
func NewHTTPMetricsObserver(name string, reqFilter otelhttp.Filter) func(next http.Handler) http.Handler {
	var m httpMetricsObserver
	m.meter = global.Meter(name)
	m.latency = metric.Must(m.meter).NewFloat64ValueRecorder(
		httpLatencyName,
		metric.WithDescription("How long it took to process the request, partitioned by status code, method and HTTP path."),
	)

	m.total_request = metric.Must(m.meter).NewInt64Counter(
		httpRequestNumberName,
		metric.WithDescription("How many request throughput, partitioned by status code, method and HTTP path."),
	)
	m.reqFilter = reqFilter
	return m.handler
}
