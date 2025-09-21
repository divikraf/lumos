package hook

import (
	"context"

	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel/trace"
)

// NewOpenTelemetryHook creates a hook that adds OpenTelemetry trace context to logs
func NewOpenTelemetryHook() zerolog.Hook {
	return zerolog.HookFunc(
		func(e *zerolog.Event, level zerolog.Level, message string) {
			// Add OpenTelemetry trace context to logs
			span := trace.SpanFromContext(context.Background())
			if span.IsRecording() {
				spanCtx := span.SpanContext()
				if spanCtx.IsValid() {
					e.Str("trace_id", spanCtx.TraceID().String())
					e.Str("span_id", spanCtx.SpanID().String())
					e.Str("trace_flags", spanCtx.TraceFlags().String())
				}
			}
		},
	)
}

// NewRelicRecorderHook is deprecated - use NewOpenTelemetryHook instead
// Kept for backward compatibility during migration
func NewRelicRecorderHook(txn interface{}) zerolog.Hook {
	return NewOpenTelemetryHook()
}
