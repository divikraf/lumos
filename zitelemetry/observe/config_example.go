package observe

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

// ExampleConfigBasedNoOp shows how to use no-op implementations based on configuration
func ExampleConfigBasedNoOp() {
	// Example 1: Tracing ENABLED
	fmt.Println("=== Example 1: Tracing ENABLED ===")

	enabledConfig := Config{
		Service: ServiceConfig{
			Name: "example-service",
		},
		Environment: "development",
		Tracing: TracingConfig{
			Enabled: true, // Tracing is enabled
			Exporter: ExporterConfig{
				Type:     "console",
				Endpoint: "",
			},
		},
		Metrics: MetricsConfig{
			Enabled: true,
		},
	}

	// Create telemetry with tracing enabled
	telemetry, err := New(context.Background(), enabledConfig)
	if err != nil {
		log.Printf("Failed to create telemetry: %v", err)
	} else {
		defer telemetry.Shutdown(context.Background())
	}

	// Create service context with tracing enabled
	ctx := CreateServiceContext(context.Background(), enabledConfig, "example-service", telemetry)

	// This will use REAL tracing
	processWithTracing(ctx, "enabled-request")

	// Example 2: Tracing DISABLED
	fmt.Println("\n=== Example 2: Tracing DISABLED ===")

	disabledConfig := Config{
		Service: ServiceConfig{
			Name: "example-service",
		},
		Environment: "development",
		Tracing: TracingConfig{
			Enabled: false, // Tracing is disabled
		},
		Metrics: MetricsConfig{
			Enabled: true,
		},
	}

	// Create service context with tracing disabled
	ctx2 := CreateServiceContext(context.Background(), disabledConfig, "example-service", nil)

	// This will use NO-OP tracing (zero overhead)
	processWithTracing(ctx2, "disabled-request")

	// Example 3: Check configuration in code
	fmt.Println("\n=== Example 3: Check Configuration ===")

	checkTracingConfig(ctx, "enabled-context")
	checkTracingConfig(ctx2, "disabled-context")
}

// processWithTracing demonstrates how tracing works regardless of configuration
func processWithTracing(ctx context.Context, requestID string) {
	// Extract tracer from context - this will be real or no-op based on config
	tracer := EnhancedFromContext(ctx)

	// Start span - this will be real or no-op based on config
	spanCtx, span := tracer.Start(ctx, "process-request")
	defer span.End()

	// Set attributes - this will be real or no-op based on config
	span.SetAttributes(
		attribute.String("request.id", requestID),
		attribute.String("operation", "process"),
		attribute.String("timestamp", time.Now().Format(time.RFC3339)),
	)

	// Extract span from context for nested operations
	spanFromCtx := EnhancedSpanFromContext(spanCtx)
	spanFromCtx.SetAttributes(attribute.String("nested.operation", "processing"))

	// Simulate some work
	time.Sleep(1 * time.Millisecond)

	// Set status
	span.SetStatus(codes.Ok, "operation completed successfully")

	fmt.Printf("Processed request: %s\n", requestID)
}

// checkTracingConfig shows how to check if tracing is enabled
func checkTracingConfig(ctx context.Context, contextName string) {
	enhancedCtx := FromEnhancedContext(ctx)
	if enhancedCtx != nil {
		fmt.Printf("Context '%s': Tracing enabled = %v, Metrics enabled = %v\n",
			contextName,
			enhancedCtx.IsTracingEnabled(),
			enhancedCtx.IsMetricsEnabled())

		// Get the tracer to check its type
		tracer := enhancedCtx.CreateTracer("check")
		fmt.Printf("Context '%s': Tracer type = %T\n", contextName, tracer)
	} else {
		fmt.Printf("Context '%s': No enhanced context found\n", contextName)
	}
}

// ExampleServiceWithConfig shows how a service can use config-based telemetry
type ExampleService struct {
	config Config
}

func NewExampleService(config Config) *ExampleService {
	return &ExampleService{config: config}
}

func (s *ExampleService) HandleRequest(ctx context.Context, request Request) Response {
	// Create service context based on configuration
	serviceCtx := CreateServiceContext(ctx, s.config, "example-service", nil)

	// Extract tracer - will be real or no-op based on config
	tracer := EnhancedFromContext(serviceCtx)

	// Start span - will be real or no-op based on config
	spanCtx, span := tracer.Start(serviceCtx, "handle-request")
	defer span.End()

	// Set request attributes
	span.SetAttributes(
		attribute.String("request.method", request.Method),
		attribute.String("request.path", request.Path),
		attribute.Int("request.size", len(request.Body)),
	)

	// Process the request
	response := s.processRequest(spanCtx, request)

	// Set response attributes
	span.SetAttributes(
		attribute.String("response.status", response.Status),
		attribute.Int("response.size", len(response.Body)),
	)

	// Set span status based on response
	if response.Status == "error" {
		span.SetStatus(codes.Error, response.Error)
	} else {
		span.SetStatus(codes.Ok, "request processed successfully")
	}

	return response
}

func (s *ExampleService) processRequest(ctx context.Context, request Request) Response {
	// Extract span from context for nested operations
	span := EnhancedSpanFromContext(ctx)
	span.SetAttributes(attribute.String("processing.stage", "business-logic"))

	// Simulate processing
	time.Sleep(2 * time.Millisecond)

	// Your business logic here
	return Response{
		Status: "success",
		Body:   fmt.Sprintf("Processed: %s", request.Body),
		Error:  "",
	}
}

// ExampleConfigurationScenarios shows different configuration scenarios
func ExampleConfigurationScenarios() {
	scenarios := []struct {
		name   string
		config Config
	}{
		{
			name: "Development with tracing",
			config: Config{
				Service:     ServiceConfig{Name: "dev-service"},
				Environment: "development",
				Tracing:     TracingConfig{Enabled: true},
				Metrics:     MetricsConfig{Enabled: true},
			},
		},
		{
			name: "Production with tracing",
			config: Config{
				Service:     ServiceConfig{Name: "prod-service"},
				Environment: "production",
				Tracing:     TracingConfig{Enabled: true},
				Metrics:     MetricsConfig{Enabled: true},
			},
		},
		{
			name: "Testing without tracing",
			config: Config{
				Service:     ServiceConfig{Name: "test-service"},
				Environment: "test",
				Tracing:     TracingConfig{Enabled: false},
				Metrics:     MetricsConfig{Enabled: false},
			},
		},
		{
			name: "Performance testing (metrics only)",
			config: Config{
				Service:     ServiceConfig{Name: "perf-service"},
				Environment: "performance",
				Tracing:     TracingConfig{Enabled: false},
				Metrics:     MetricsConfig{Enabled: true},
			},
		},
	}

	for _, scenario := range scenarios {
		fmt.Printf("\n=== Scenario: %s ===\n", scenario.name)

		// Create service context
		ctx := CreateServiceContext(context.Background(), scenario.config, scenario.config.Service.Name, nil)

		// Check what type of tracer we get
		tracer := EnhancedFromContext(ctx)
		fmt.Printf("Tracer type: %T\n", tracer)

		// Create a span to see the behavior
		_, span := tracer.Start(ctx, "scenario-test")
		span.SetAttributes(attribute.String("scenario", scenario.name))
		span.End()

		fmt.Printf("Tracing enabled: %v\n", scenario.config.Tracing.Enabled)
		fmt.Printf("Metrics enabled: %v\n", scenario.config.Metrics.Enabled)
	}
}

// Request and Response types for examples
type Request struct {
	Method string
	Path   string
	Body   string
}

type Response struct {
	Status string
	Body   string
	Error  string
}

// ExampleFXIntegration shows how to integrate with FX
func ExampleFXIntegration() {
	// This would typically be in your main.go or service setup

	// Configuration from environment or config file
	config := Config{
		Service: ServiceConfig{
			Name: "my-service",
		},
		Environment: "development",
		Tracing: TracingConfig{
			Enabled: false, // Can be controlled by environment variable
		},
		Metrics: MetricsConfig{
			Enabled: true,
		},
	}

	// Create telemetry (only if needed)
	var telemetry *Telemetry
	if config.Tracing.Enabled || config.Metrics.Enabled {
		var err error
		telemetry, err = New(context.Background(), config)
		if err != nil {
			log.Printf("Failed to create telemetry: %v", err)
		} else {
			defer telemetry.Shutdown(context.Background())
		}
	}

	// Create service with configuration
	service := NewExampleService(config)

	// Create request context with appropriate telemetry
	ctx := CreateServiceContext(context.Background(), config, "my-service", telemetry)

	// Handle request - will use real or no-op tracing based on config
	request := Request{
		Method: "GET",
		Path:   "/api/users",
		Body:   "user data",
	}

	response := service.HandleRequest(ctx, request)
	fmt.Printf("Response: %+v\n", response)
}
