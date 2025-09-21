package observe

import (
	"context"
)

type enhancedCtxKey struct{}

// EnhancedContextKey the context key for enhanced telemetry context.
var EnhancedContextKey enhancedCtxKey = enhancedCtxKey(struct{}{})

// EnhancedContext wraps context with telemetry configuration and factory
type EnhancedContext struct {
	tracerFactory *TracerFactory
	config        Config
}

// WithEnhancedContext wraps ctx with telemetry configuration and factory
func WithEnhancedContext(ctx context.Context, config Config, telemetry *Telemetry) context.Context {
	factory := NewTracerFactory(config, telemetry)
	enhancedCtx := &EnhancedContext{
		tracerFactory: factory,
		config:        config,
	}
	return context.WithValue(ctx, EnhancedContextKey, enhancedCtx)
}

// FromEnhancedContext extracts EnhancedContext from context
func FromEnhancedContext(ctx context.Context) *EnhancedContext {
	v := ctx.Value(EnhancedContextKey)
	if v == nil {
		return nil
	}
	return v.(*EnhancedContext)
}

// CreateTracer creates a tracer based on the configuration in context
func (ec *EnhancedContext) CreateTracer(name string) Tracer {
	return ec.tracerFactory.CreateTracer(name)
}

// CreateContext creates a new context with tracer based on configuration
func (ec *EnhancedContext) CreateContext(ctx context.Context, tracerName string) context.Context {
	return ec.tracerFactory.CreateContext(ctx, tracerName)
}

// IsTracingEnabled returns whether tracing is enabled
func (ec *EnhancedContext) IsTracingEnabled() bool {
	return ec.config.Tracing.Enabled
}

// IsMetricsEnabled returns whether metrics are enabled
func (ec *EnhancedContext) IsMetricsEnabled() bool {
	return ec.config.Metrics.Enabled
}

// GetConfig returns the configuration
func (ec *EnhancedContext) GetConfig() Config {
	return ec.config
}

// EnhancedFromContext extracts Tracer from context with config awareness
func EnhancedFromContext(ctx context.Context) Tracer {
	enhancedCtx := FromEnhancedContext(ctx)
	if enhancedCtx == nil {
		// Fall back to regular context extraction
		return FromContext(ctx)
	}

	// If tracing is disabled, return no-op tracer
	if !enhancedCtx.IsTracingEnabled() {
		return NewNoOpTracer()
	}

	// Create tracer based on configuration
	return enhancedCtx.CreateTracer("default")
}

// EnhancedSpanFromContext extracts Span from context with config awareness
func EnhancedSpanFromContext(ctx context.Context) Span {
	enhancedCtx := FromEnhancedContext(ctx)
	if enhancedCtx == nil {
		// Fall back to regular context extraction
		return SpanFromContext(ctx)
	}

	// If tracing is disabled, return no-op span
	if !enhancedCtx.IsTracingEnabled() {
		return NewNoOpSpan()
	}

	// Extract span using the tracer from enhanced context
	tracer := enhancedCtx.CreateTracer("default")
	span := tracer.SpanFromContext(ctx)
	return NewTelemetrySpan(span)
}

// CreateServiceContext creates a context for a service with appropriate telemetry
func CreateServiceContext(ctx context.Context, config Config, serviceName string, telemetry *Telemetry) context.Context {
	// Create enhanced context
	enhancedCtx := WithEnhancedContext(ctx, config, telemetry)

	// If tracing is enabled, create a service-specific tracer
	if config.Tracing.Enabled {
		tracer := NewTracerFactory(config, telemetry).CreateTracer(serviceName)
		return WithContext(enhancedCtx, tracer)
	}

	// If tracing is disabled, use no-op tracer
	tracer := NewNoOpTracer()
	return WithContext(enhancedCtx, tracer)
}
