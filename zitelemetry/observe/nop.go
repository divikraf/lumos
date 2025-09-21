package observe

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/embedded"
)

// NoOpTelemetry is a no-op implementation of telemetry
type NoOpTelemetry struct{}

// NewNoOpTelemetry creates a new no-op telemetry instance
func NewNoOpTelemetry() *NoOpTelemetry {
	return &NoOpTelemetry{}
}

// Shutdown does nothing
func (n *NoOpTelemetry) Shutdown(ctx context.Context) error {
	return nil
}

// GetConfig returns an empty config
func (n *NoOpTelemetry) GetConfig() Config {
	return Config{}
}

// NoOpSpan is a no-op span implementation
type NoOpSpan struct {
	embedded.Span
}

// NewNoOpSpan creates a new no-op span
func NewNoOpSpan() *NoOpSpan {
	return &NoOpSpan{}
}

// End does nothing
func (s *NoOpSpan) End(...trace.SpanEndOption) {}

// AddEvent does nothing
func (s *NoOpSpan) AddEvent(name string, options ...trace.EventOption) {}

// IsRecording returns false
func (s *NoOpSpan) IsRecording() bool {
	return false
}

// RecordError does nothing
func (s *NoOpSpan) RecordError(err error, options ...trace.EventOption) {}

// SpanContext returns an empty span context
func (s *NoOpSpan) SpanContext() trace.SpanContext {
	return trace.SpanContext{}
}

// SetStatus does nothing
func (s *NoOpSpan) SetStatus(code codes.Code, description string) {}

// SetName does nothing
func (s *NoOpSpan) SetName(name string) {}

// SetAttributes does nothing
func (s *NoOpSpan) SetAttributes(kv ...attribute.KeyValue) {}

// AddLink does nothing
func (s *NoOpSpan) AddLink(link trace.Link) {}

// TracerProvider returns a no-op tracer provider
func (s *NoOpSpan) TracerProvider() trace.TracerProvider {
	return trace.NewNoopTracerProvider()
}

// Compile-time interface compliance checks
var (
	_ trace.Span = (*NoOpSpan)(nil)
	_ Span       = (*NoOpSpan)(nil)
)
