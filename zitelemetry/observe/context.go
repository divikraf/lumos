package observe

import (
	"context"
)

type ctxKey struct{}

// contextKey the context key for Tracer.
var contextKey ctxKey = ctxKey(struct{}{})

// WithContext wrap ctx with tracer.
func WithContext(ctx context.Context, tracer Tracer) context.Context {
	return context.WithValue(ctx, contextKey, tracer)
}

// FromContext extracts Tracer from context.
// Returns noop Tracer if no Tracer is associated with context.
func FromContext(ctx context.Context) Tracer {
	v := ctx.Value(contextKey)
	if v == nil {
		return NewNoOpTracer()
	}
	return v.(Tracer)
}

// SpanFromContext extracts current Span from context.
// Returns noop Span if no Span is associated with context.
func SpanFromContext(ctx context.Context) Span {
	inst := FromContext(ctx)
	span := inst.SpanFromContext(ctx)
	return NewTelemetrySpan(span)
}
