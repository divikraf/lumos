package observe

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/host"
	"go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

// Telemetry represents the OpenTelemetry setup
type Telemetry struct {
	config        Config
	shutdownFuncs []func(context.Context) error
}

// New creates a new Telemetry instance with the given configuration
func New(ctx context.Context, config Config) (*Telemetry, error) {
	t := &Telemetry{
		config:        config,
		shutdownFuncs: make([]func(context.Context) error, 0),
	}

	if err := t.init(ctx); err != nil {
		return nil, fmt.Errorf("failed to initialize telemetry: %w", err)
	}

	return t, nil
}

// init initializes the OpenTelemetry components
func (t *Telemetry) init(ctx context.Context) error {
	// Create resource
	res, err := t.createResource(ctx)
	if err != nil {
		slog.WarnContext(ctx, "failed to create resource", "error", err)
		// Continue with minimal resource
		res, _ = resource.New(ctx, resource.WithAttributes(
			semconv.ServiceName(t.config.Service.Name),
			semconv.DeploymentEnvironmentKey.String(t.config.Environment),
		))
	}

	// Set up propagator
	t.setupPropagator()

	// Set up tracing if enabled
	if t.config.Tracing.Enabled {
		if err := t.setupTracing(ctx, res); err != nil {
			return fmt.Errorf("failed to setup tracing: %w", err)
		}
	}

	// Set up metrics if enabled
	if t.config.Metrics.Enabled {
		if err := t.setupMetrics(ctx, res); err != nil {
			return fmt.Errorf("failed to setup metrics: %w", err)
		}
	}

	// Start infrastructure metrics if enabled
	if t.config.Metrics.Enabled {
		if err := t.startInfraMetrics(); err != nil {
			slog.WarnContext(ctx, "failed to start infrastructure metrics", "error", err)
		}
	}

	return nil
}

// createResource creates the OpenTelemetry resource
func (t *Telemetry) createResource(ctx context.Context) (*resource.Resource, error) {
	attrs := []attribute.KeyValue{
		semconv.ServiceName(t.config.Service.Name),
		semconv.DeploymentEnvironmentKey.String(t.config.Environment),
	}

	opts := []resource.Option{
		resource.WithAttributes(attrs...),
		resource.WithTelemetrySDK(),
	}

	return resource.New(ctx, opts...)
}

// setupPropagator sets up the OpenTelemetry propagator
func (t *Telemetry) setupPropagator() {
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))
}

// setupTracing sets up the tracing components
func (t *Telemetry) setupTracing(ctx context.Context, res *resource.Resource) error {
	// Create exporter
	exporter, err := t.createTraceExporter(ctx)
	if err != nil {
		return fmt.Errorf("failed to create trace exporter: %w", err)
	}

	// Create sampler
	sampler := t.createSampler()

	// Create tracer provider options
	opts := []trace.TracerProviderOption{
		trace.WithBatcher(exporter,
			trace.WithMaxExportBatchSize(t.config.Tracing.Batch.MaxExportBatchSize),
			trace.WithExportTimeout(t.config.Tracing.Batch.ExportTimeout),
			trace.WithMaxQueueSize(t.config.Tracing.Batch.MaxQueueSize),
		),
		trace.WithResource(res),
		trace.WithSampler(sampler),
	}

	// Create tracer provider
	tp := trace.NewTracerProvider(opts...)
	t.shutdownFuncs = append(t.shutdownFuncs, tp.Shutdown)
	otel.SetTracerProvider(tp)

	slog.InfoContext(ctx, "tracing initialized",
		"exporter", t.config.Tracing.Exporter.Type,
		"sampler", t.config.Tracing.Sampler.Type)

	return nil
}

// setupMetrics sets up the metrics components
func (t *Telemetry) setupMetrics(ctx context.Context, res *resource.Resource) error {
	// Create exporter
	exporter, err := t.createMetricExporter(ctx)
	if err != nil {
		return fmt.Errorf("failed to create metric exporter: %w", err)
	}

	// Create reader options
	readerOpts := []metric.PeriodicReaderOption{
		metric.WithInterval(t.config.Metrics.Reader.Interval),
		metric.WithTimeout(t.config.Metrics.Reader.Timeout),
	}

	// Create meter provider
	mp := metric.NewMeterProvider(
		metric.WithResource(res),
		metric.WithReader(metric.NewPeriodicReader(exporter, readerOpts...)),
	)
	t.shutdownFuncs = append(t.shutdownFuncs, mp.Shutdown)
	otel.SetMeterProvider(mp)

	slog.InfoContext(ctx, "metrics initialized",
		"exporter", t.config.Metrics.Exporter.Type,
		"interval", t.config.Metrics.Reader.Interval)

	return nil
}

// createTraceExporter creates the appropriate trace exporter
func (t *Telemetry) createTraceExporter(ctx context.Context) (trace.SpanExporter, error) {
	switch t.config.Tracing.Exporter.Type {
	case "otlp":
		return t.createOTLPTraceExporter(ctx)
	case "console":
		return t.createConsoleTraceExporter()
	case "none":
		return &noopTraceExporter{}, nil
	default:
		return t.createConsoleTraceExporter()
	}
}

// createMetricExporter creates the appropriate metric exporter
func (t *Telemetry) createMetricExporter(ctx context.Context) (metric.Exporter, error) {
	switch t.config.Metrics.Exporter.Type {
	case "otlp":
		return t.createOTLPMetricExporter(ctx)
	case "console":
		return t.createConsoleMetricExporter()
	case "none":
		return &noopMetricExporter{}, nil
	default:
		return t.createConsoleMetricExporter()
	}
}

// createOTLPTraceExporter creates an OTLP trace exporter
func (t *Telemetry) createOTLPTraceExporter(ctx context.Context) (trace.SpanExporter, error) {
	config := t.config.Tracing.Exporter

	if config.Protocol == "grpc" {
		opts := []otlptracegrpc.Option{
			otlptracegrpc.WithEndpoint(config.Endpoint),
			otlptracegrpc.WithTimeout(config.Timeout),
		}
		if config.Insecure {
			opts = append(opts, otlptracegrpc.WithInsecure())
		}
		if len(config.Headers) > 0 {
			opts = append(opts, otlptracegrpc.WithHeaders(config.Headers))
		}
		return otlptracegrpc.New(ctx, opts...)
	}

	// HTTP protocol
	opts := []otlptracehttp.Option{
		otlptracehttp.WithEndpoint(config.Endpoint),
		otlptracehttp.WithTimeout(config.Timeout),
	}
	if config.Insecure {
		opts = append(opts, otlptracehttp.WithInsecure())
	}
	if len(config.Headers) > 0 {
		opts = append(opts, otlptracehttp.WithHeaders(config.Headers))
	}
	return otlptracehttp.New(ctx, opts...)
}

// createOTLPMetricExporter creates an OTLP metric exporter
func (t *Telemetry) createOTLPMetricExporter(ctx context.Context) (metric.Exporter, error) {
	config := t.config.Metrics.Exporter

	if config.Protocol == "grpc" {
		opts := []otlpmetricgrpc.Option{
			otlpmetricgrpc.WithEndpoint(config.Endpoint),
			otlpmetricgrpc.WithTimeout(config.Timeout),
		}
		if config.Insecure {
			opts = append(opts, otlpmetricgrpc.WithInsecure())
		}
		if len(config.Headers) > 0 {
			opts = append(opts, otlpmetricgrpc.WithHeaders(config.Headers))
		}
		return otlpmetricgrpc.New(ctx, opts...)
	}

	// HTTP protocol
	opts := []otlpmetrichttp.Option{
		otlpmetrichttp.WithEndpoint(config.Endpoint),
		otlpmetrichttp.WithTimeout(config.Timeout),
	}
	if config.Insecure {
		opts = append(opts, otlpmetrichttp.WithInsecure())
	}
	if len(config.Headers) > 0 {
		opts = append(opts, otlpmetrichttp.WithHeaders(config.Headers))
	}
	return otlpmetrichttp.New(ctx, opts...)
}

// createConsoleTraceExporter creates a console trace exporter
func (t *Telemetry) createConsoleTraceExporter() (trace.SpanExporter, error) {
	return stdouttrace.New(stdouttrace.WithPrettyPrint())
}

// createConsoleMetricExporter creates a console metric exporter
func (t *Telemetry) createConsoleMetricExporter() (metric.Exporter, error) {
	return stdoutmetric.New(stdoutmetric.WithPrettyPrint())
}

// createSampler creates the appropriate sampler
func (t *Telemetry) createSampler() trace.Sampler {
	switch t.config.Tracing.Sampler.Type {
	case "always_on":
		return trace.AlwaysSample()
	case "always_off":
		return trace.NeverSample()
	case "traceidratio":
		return trace.TraceIDRatioBased(t.config.Tracing.Sampler.Fraction)
	case "parentbased":
		return trace.ParentBased(trace.TraceIDRatioBased(t.config.Tracing.Sampler.Fraction))
	default:
		return trace.AlwaysSample()
	}
}

// startInfraMetrics starts infrastructure metrics collection
func (t *Telemetry) startInfraMetrics() error {
	if err := host.Start(); err != nil {
		return fmt.Errorf("failed to start host metrics: %w", err)
	}

	if err := runtime.Start(runtime.WithMinimumReadMemStatsInterval(10 * time.Second)); err != nil {
		return fmt.Errorf("failed to start runtime metrics: %w", err)
	}

	return nil
}

// Shutdown gracefully shuts down the telemetry system
func (t *Telemetry) Shutdown(ctx context.Context) error {
	var errs []error
	for _, fn := range t.shutdownFuncs {
		if err := fn(ctx); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("shutdown errors: %v", errs)
	}
	return nil
}

// GetConfig returns the current configuration
func (t *Telemetry) GetConfig() Config {
	return t.config
}

// noopTraceExporter is a no-op trace exporter
type noopTraceExporter struct{}

func (e *noopTraceExporter) ExportSpans(ctx context.Context, spans []trace.ReadOnlySpan) error {
	return nil
}

func (e *noopTraceExporter) Shutdown(ctx context.Context) error {
	return nil
}

// noopMetricExporter is a no-op metric exporter
type noopMetricExporter struct{}

func (e *noopMetricExporter) Export(ctx context.Context, metrics *metricdata.ResourceMetrics) error {
	return nil
}

func (e *noopMetricExporter) ForceFlush(ctx context.Context) error {
	return nil
}

func (e *noopMetricExporter) Shutdown(ctx context.Context) error {
	return nil
}

func (e *noopMetricExporter) Aggregation(kind metric.InstrumentKind) metric.Aggregation {
	return metric.AggregationDefault{}
}

func (e *noopMetricExporter) Temporality(kind metric.InstrumentKind) metricdata.Temporality {
	return metricdata.CumulativeTemporality
}
