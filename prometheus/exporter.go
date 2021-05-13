package prometheus

import (
	"context"
	"time"

	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	client_prometheus "github.com/prometheus/client_golang/prometheus"
	otel_instrumentation_host "go.opentelemetry.io/contrib/instrumentation/host"
	otel_instrumentation_runtime "go.opentelemetry.io/contrib/instrumentation/runtime"
	otel_exporters_prometheus "go.opentelemetry.io/otel/exporters/metric/prometheus"
	otel_metric_controller "go.opentelemetry.io/otel/sdk/metric/controller/basic"
	otel_resource "go.opentelemetry.io/otel/sdk/resource"
)

var (
	defaultPromRegistry  = client_prometheus.NewRegistry()
	grpcServerPrometheus = grpc_prometheus.NewServerMetrics()
	grpcClientPrometheus = grpc_prometheus.NewClientMetrics()
)

func init() {
	client_prometheus.DefaultGatherer = defaultPromRegistry
	client_prometheus.DefaultRegisterer = defaultPromRegistry

	grpcServerPrometheus.EnableHandlingTimeHistogram()
	grpcClientPrometheus.EnableClientHandlingTimeHistogram()

	err := defaultPromRegistry.Register(grpcServerPrometheus)
	if err != nil {
		return
	}
	err = defaultPromRegistry.Register(grpcClientPrometheus)
	if err != nil {
		return
	}
}

func InitExporter() (*otel_exporters_prometheus.Exporter, error) {
	resource, err := otel_resource.New(
		context.Background(),
	)

	if err != nil {
		return nil, err
	}

	return otel_exporters_prometheus.InstallNewPipeline(
		otel_exporters_prometheus.Config{
			Registry:                   defaultPromRegistry,
			DefaultHistogramBoundaries: []float64{.0005, 0.0075, 0.001, 0.002, 0.003, 0.004, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
		},
		otel_metric_controller.WithResource(resource),
		otel_metric_controller.WithCollectPeriod(10*time.Second),
	)
}

func WithRuntime() error {
	err := otel_instrumentation_runtime.Start(
		otel_instrumentation_runtime.WithMinimumReadMemStatsInterval(time.Second),
	)
	if err != nil {
		return err
	}
	err = otel_instrumentation_host.Start()
	return err
}
