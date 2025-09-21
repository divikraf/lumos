package observe

import (
	"context"

	"go.opentelemetry.io/otel/trace"
)

// Tracer interface defines the contract for tracing operations
type Tracer interface {
	// Start starts a new span
	Start(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span)

	// SpanFromContext extracts the current span from context
	SpanFromContext(ctx context.Context) trace.Span
}

// TelemetryTracer wraps the OpenTelemetry tracer with our telemetry configuration
type TelemetryTracer struct {
	tracer trace.Tracer
}

// NewTelemetryTracer creates a new telemetry tracer
func NewTelemetryTracer(tracer trace.Tracer) *TelemetryTracer {
	return &TelemetryTracer{tracer: tracer}
}

// Start starts a new span
func (t *TelemetryTracer) Start(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return t.tracer.Start(ctx, name, opts...)
}

// SpanFromContext extracts the current span from context
func (t *TelemetryTracer) SpanFromContext(ctx context.Context) trace.Span {
	return trace.SpanFromContext(ctx)
}

// NoOpTracer is a no-op implementation of Tracer
type NoOpTracer struct{}

// NewNoOpTracer creates a new no-op tracer
func NewNoOpTracer() *NoOpTracer {
	return &NoOpTracer{}
}

// Start returns the context with a no-op span
func (n *NoOpTracer) Start(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	span := NewNoOpSpan()
	return trace.ContextWithSpan(ctx, span), span
}

// SpanFromContext returns a no-op span
func (n *NoOpTracer) SpanFromContext(ctx context.Context) trace.Span {
	span := trace.SpanFromContext(ctx)
	if span.SpanContext().IsValid() {
		return span
	}
	return NewNoOpSpan()
}
