package observe

import (
	"context"

	"go.opentelemetry.io/otel/trace"
)

// TracerFactory creates tracers based on configuration
type TracerFactory struct {
	config    Config
	telemetry *Telemetry
}

// NewTracerFactory creates a new tracer factory
func NewTracerFactory(config Config, telemetry *Telemetry) *TracerFactory {
	return &TracerFactory{
		config:    config,
		telemetry: telemetry,
	}
}

// CreateTracer creates a tracer based on the configuration
func (f *TracerFactory) CreateTracer(name string) Tracer {
	// If tracing is disabled, return no-op tracer
	if !f.config.Tracing.Enabled {
		return NewNoOpTracer()
	}

	// If telemetry is not available, return no-op tracer
	if f.telemetry == nil {
		return NewNoOpTracer()
	}

	// Create real tracer using OpenTelemetry
	return NewTelemetryTracer(
		trace.SpanFromContext(context.Background()).TracerProvider().Tracer(name),
	)
}

// CreateContext creates a context with the appropriate tracer based on config
func (f *TracerFactory) CreateContext(ctx context.Context, tracerName string) context.Context {
	tracer := f.CreateTracer(tracerName)
	return WithContext(ctx, tracer)
}

// IsTracingEnabled returns whether tracing is enabled in the configuration
func (f *TracerFactory) IsTracingEnabled() bool {
	return f.config.Tracing.Enabled
}

// IsMetricsEnabled returns whether metrics are enabled in the configuration
func (f *TracerFactory) IsMetricsEnabled() bool {
	return f.config.Metrics.Enabled
}

// CreateTracerFromConfig is a convenience function to create a tracer directly from config
func CreateTracerFromConfig(config Config, name string) Tracer {
	if !config.Tracing.Enabled {
		return NewNoOpTracer()
	}

	// This would need a telemetry instance to create a real tracer
	// For now, return no-op if tracing is disabled
	return NewNoOpTracer()
}

// CreateContextFromConfig is a convenience function to create context directly from config
func CreateContextFromConfig(ctx context.Context, config Config, tracerName string) context.Context {
	tracer := CreateTracerFromConfig(config, tracerName)
	return WithContext(ctx, tracer)
}
