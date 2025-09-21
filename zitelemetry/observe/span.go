package observe

import (
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// Span interface defines the contract for span operations
type Span interface {
	// End ends the span
	End(...trace.SpanEndOption)

	// AddEvent adds an event to the span
	AddEvent(name string, options ...trace.EventOption)

	// IsRecording returns whether the span is recording
	IsRecording() bool

	// RecordError records an error in the span
	RecordError(err error, options ...trace.EventOption)

	// SpanContext returns the span context
	SpanContext() trace.SpanContext

	// SetStatus sets the status of the span
	SetStatus(code codes.Code, description string)

	// SetName sets the name of the span
	SetName(name string)

	// SetAttributes sets attributes on the span
	SetAttributes(kv ...attribute.KeyValue)

	// AddLink adds a link to the span
	AddLink(link trace.Link)
}

// TelemetrySpan wraps OpenTelemetry span
type TelemetrySpan struct {
	span trace.Span
}

// NewTelemetrySpan creates a new telemetry span
func NewTelemetrySpan(span trace.Span) *TelemetrySpan {
	return &TelemetrySpan{span: span}
}

// End ends the span
func (s *TelemetrySpan) End(options ...trace.SpanEndOption) {
	s.span.End(options...)
}

// AddEvent adds an event to the span
func (s *TelemetrySpan) AddEvent(name string, options ...trace.EventOption) {
	s.span.AddEvent(name, options...)
}

// IsRecording returns whether the span is recording
func (s *TelemetrySpan) IsRecording() bool {
	return s.span.IsRecording()
}

// RecordError records an error in the span
func (s *TelemetrySpan) RecordError(err error, options ...trace.EventOption) {
	s.span.RecordError(err, options...)
}

// SpanContext returns the span context
func (s *TelemetrySpan) SpanContext() trace.SpanContext {
	return s.span.SpanContext()
}

// SetStatus sets the status of the span
func (s *TelemetrySpan) SetStatus(code codes.Code, description string) {
	s.span.SetStatus(code, description)
}

// SetName sets the name of the span
func (s *TelemetrySpan) SetName(name string) {
	s.span.SetName(name)
}

// SetAttributes sets attributes on the span
func (s *TelemetrySpan) SetAttributes(kv ...attribute.KeyValue) {
	s.span.SetAttributes(kv...)
}

// AddLink adds a link to the span
func (s *TelemetrySpan) AddLink(link trace.Link) {
	s.span.AddLink(link)
}

// Compile-time interface compliance check
var _ Span = (*TelemetrySpan)(nil)
