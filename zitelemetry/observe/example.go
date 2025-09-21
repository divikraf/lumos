package observe

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// Example shows how to use the context-based telemetry implementation
func Example() {
	// Create a telemetry instance
	config := Config{
		Service: ServiceConfig{
			Name: "example-service",
		},
		Environment: "development",
		Tracing: TracingConfig{
			Enabled: true,
		},
		Metrics: MetricsConfig{
			Enabled: true,
		},
	}

	_, err := New(context.Background(), config)
	if err != nil {
		fmt.Printf("Failed to create telemetry: %v\n", err)
		return
	}

	// Create a tracer from the telemetry instance
	tracer := NewTelemetryTracer(
		trace.SpanFromContext(context.Background()).TracerProvider().Tracer("example"),
	)

	// Create context with the tracer
	ctx := WithContext(context.Background(), tracer)

	// Example 1: Extract tracer from context
	extractedTracer := FromContext(ctx)
	if extractedTracer != nil {
		fmt.Println("Tracer extracted from context successfully")
	}

	// Example 2: Create and use a span
	spanCtx, span := extractedTracer.Start(ctx, "example-operation")
	defer span.End()

	// Set attributes on the span
	span.SetAttributes(
		attribute.String("operation.name", "example"),
		attribute.Int("operation.id", 123),
	)

	// Example 3: Extract span from context
	spanFromCtx := SpanFromContext(spanCtx)
	if spanFromCtx != nil {
		fmt.Println("Span extracted from context successfully")
		spanFromCtx.SetAttributes(attribute.String("context.span", "extracted"))
	}

	// Example 4: Use no-op tracer when no tracer is in context
	emptyCtx := context.Background()
	noOpTracer := FromContext(emptyCtx)
	noOpSpan := noOpTracer.SpanFromContext(emptyCtx)
	
	// This will be a no-op span
	noOpSpan.SetAttributes(attribute.String("noop", "true"))
	noOpSpan.End()

	fmt.Println("Context-based telemetry example completed")
}

// ExampleWithNoTelemetry shows how the system gracefully handles missing telemetry
func ExampleWithNoTelemetry() {
	// Context without any telemetry
	ctx := context.Background()

	// This will return a no-op tracer
	tracer := FromContext(ctx)
	fmt.Printf("Tracer type: %T\n", tracer) // Should print *NoOpTracer

	// This will return a no-op span
	span := SpanFromContext(ctx)
	fmt.Printf("Span type: %T\n", span) // Should print *NoOpSpan

	// These operations will be no-ops but won't panic
	span.SetAttributes(attribute.String("test", "noop"))
	span.End()

	fmt.Println("No-telemetry example completed successfully")
}

// ExampleNestedSpans shows how to create nested spans
func ExampleNestedSpans() {
	// Create telemetry
	config := Config{
		Service: ServiceConfig{
			Name: "nested-example",
		},
		Environment: "development",
		Tracing: TracingConfig{
			Enabled: true,
		},
	}

	_, err := New(context.Background(), config)
	if err != nil {
		fmt.Printf("Failed to create telemetry: %v\n", err)
		return
	}

	tracer := NewTelemetryTracer(
		trace.SpanFromContext(context.Background()).TracerProvider().Tracer("nested"),
	)

	ctx := WithContext(context.Background(), tracer)

	// Parent span
	parentCtx, parentSpan := tracer.Start(ctx, "parent-operation")
	defer parentSpan.End()

	parentSpan.SetAttributes(attribute.String("span.type", "parent"))

	// Child span
	childCtx, childSpan := tracer.Start(parentCtx, "child-operation")
	defer childSpan.End()

	childSpan.SetAttributes(attribute.String("span.type", "child"))

	// Extract spans from context
	parentSpanFromCtx := SpanFromContext(parentCtx)
	childSpanFromCtx := SpanFromContext(childCtx)

	parentSpanFromCtx.SetAttributes(attribute.String("extracted", "parent"))
	childSpanFromCtx.SetAttributes(attribute.String("extracted", "child"))

	fmt.Println("Nested spans example completed")
}

// ExampleServiceIntegration shows how to integrate with a service
func ExampleServiceIntegration() {
	// This would typically be set up during application initialization
	tracer := NewTelemetryTracer(
		trace.SpanFromContext(context.Background()).TracerProvider().Tracer("service"),
	)

	// Create service context with tracer
	serviceCtx := WithContext(context.Background(), tracer)

	// Service method that uses telemetry
	result := processRequest(serviceCtx, "example-request")
	fmt.Printf("Process result: %s\n", result)
}

func processRequest(ctx context.Context, request string) string {
	// Extract tracer from context
	tracer := FromContext(ctx)

	// Start span for this operation
	spanCtx, span := tracer.Start(ctx, "process-request")
	defer span.End()

	span.SetAttributes(
		attribute.String("request.id", request),
		attribute.String("service.name", "example-service"),
	)

	// Simulate some work
	time.Sleep(10 * time.Millisecond)

	// Extract span from context to add more attributes
	spanFromCtx := SpanFromContext(spanCtx)
	spanFromCtx.SetAttributes(attribute.String("processing.status", "completed"))

	return fmt.Sprintf("processed-%s", request)
}
