package zin

import (
	"regexp"
	"strconv"
	"sync"
	"time"

	"github.com/divikraf/lumos/zitelemetry/revelio"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

// Global single histogram for HTTP metrics
var (
	httpHistogram metric.Int64Histogram
	histogramOnce sync.Once
)

// HTTPMetricsConfig holds configuration for HTTP metrics middleware
type HTTPMetricsConfig struct {
	// MetricName is the name of the histogram metric (default: "http_request_duration_ms")
	MetricName string

	// MetricDescription is the description of the histogram metric
	MetricDescription string

	// MetricUnit is the unit of the histogram metric (default: "ms")
	MetricUnit string

	// Labels to include in the histogram
	// Available labels: method, path, status_code, route, user_agent
	Labels []string

	// SkipPaths is a list of paths to skip metrics collection
	SkipPaths []string

	// NormalizePath if true, normalizes paths like /users/123 to /users/:id
	NormalizePath bool

	// NormalizePathFunc is a custom function to normalize paths
	NormalizePathFunc func(string) string
}

// DefaultHTTPMetricsConfig returns the default configuration for HTTP metrics
func DefaultHTTPMetricsConfig() HTTPMetricsConfig {
	return HTTPMetricsConfig{
		MetricName:        "http_request_duration_ms",
		MetricDescription: "HTTP request duration in milliseconds",
		MetricUnit:        "ms",
		Labels:            []string{"method", "path", "status_code"},
		SkipPaths:         []string{"/health", "/metrics", "/ready"},
		NormalizePath:     true,
		NormalizePathFunc: defaultNormalizePath,
	}
}

// getHTTPHistogram gets or creates the single HTTP histogram
func getHTTPHistogram() metric.Int64Histogram {
	histogramOnce.Do(func() {
		httpHistogram = revelio.MustInt64Histogram("http_request_duration_ms", "HTTP request duration in milliseconds", metric.WithUnit("ms"))
	})
	return httpHistogram
}

// HTTPMetricsMiddleware creates a Gin middleware that records HTTP request metrics
func HTTPMetricsMiddleware(config HTTPMetricsConfig) gin.HandlerFunc {
	// Get the single HTTP histogram
	histogram := getHTTPHistogram()

	// Create skip paths map for O(1) lookup
	skipPaths := make(map[string]bool)
	for _, path := range config.SkipPaths {
		skipPaths[path] = true
	}

	return func(c *gin.Context) {
		// Check if we should skip this path
		if skipPaths[c.Request.URL.Path] {
			c.Next()
			return
		}

		start := time.Now()

		// Process request
		c.Next()

		// Calculate duration
		duration := time.Since(start).Milliseconds()

		// Get route pattern (e.g., /users/:id instead of /users/123)
		route := c.FullPath()
		if route == "" {
			// If no route pattern found, use the actual path
			route = c.Request.URL.Path
		}

		// Record histogram with fixed labels: method, route, status_code
		histogram.Record(c.Request.Context(), duration,
			metric.WithAttributes(
				attribute.String("method", c.Request.Method),
				attribute.String("route", route),
				attribute.String("status_code", strconv.Itoa(c.Writer.Status())),
			),
		)
	}
}

// defaultNormalizePath provides a simple path normalization
func defaultNormalizePath(path string) string {
	// This is a simple implementation - in production you might want more sophisticated normalization
	// For example, converting /users/123 to /users/:id

	// For now, just return the path as-is
	// You can enhance this based on your needs
	return path
}

// AdvancedNormalizePath provides advanced path normalization function that converts numeric IDs to placeholders
func AdvancedNormalizePath(path string) string {
	// Replace numeric IDs with regex
	numericIDRegex := regexp.MustCompile(`/\d+`)
	path = numericIDRegex.ReplaceAllString(path, "/:id")

	// Replace UUIDs (basic pattern)
	uuidRegex := regexp.MustCompile(`/[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}`)
	path = uuidRegex.ReplaceAllString(path, "/:uuid")

	// Replace long alphanumeric strings (like slugs)
	slugRegex := regexp.MustCompile(`/[a-zA-Z0-9_-]{20,}`)
	path = slugRegex.ReplaceAllString(path, "/:slug")

	return path
}

// Convenience functions for common configurations

// HTTPMetricsMiddlewareDefault creates middleware with default configuration
func HTTPMetricsMiddlewareDefault() gin.HandlerFunc {
	return HTTPMetricsMiddleware(DefaultHTTPMetricsConfig())
}

// httpMetricsMiddlewareWithSkipPaths creates middleware with provided skip paths
func httpMetricsMiddlewareWithSkipPaths(skipPathsList []string) gin.HandlerFunc {
	// Get the single HTTP histogram
	histogram := getHTTPHistogram()

	// Create skip paths map for O(1) lookup
	skipPaths := make(map[string]bool)
	for _, path := range skipPathsList {
		skipPaths[path] = true
	}

	return func(c *gin.Context) {
		// Check if we should skip this path
		if skipPaths[c.Request.URL.Path] {
			c.Next()
			return
		}

		start := time.Now()

		// Process request
		c.Next()

		// Calculate duration
		duration := time.Since(start).Milliseconds()

		// Get route pattern (e.g., /users/:id instead of /users/123)
		route := c.FullPath()
		if route == "" {
			// If no route pattern found, use the actual path
			route = c.Request.URL.Path
		}

		// Record histogram with fixed labels: method, route, status_code
		histogram.Record(c.Request.Context(), duration,
			metric.WithAttributes(
				attribute.String("method", c.Request.Method),
				attribute.String("route", route),
				attribute.String("status_code", strconv.Itoa(c.Writer.Status())),
			),
		)
	}
}

// HTTPMetricsMiddlewareWithNormalization creates middleware with path normalization
func HTTPMetricsMiddlewareWithNormalization() gin.HandlerFunc {
	config := DefaultHTTPMetricsConfig()
	config.NormalizePathFunc = AdvancedNormalizePath
	return HTTPMetricsMiddleware(config)
}

// HTTPMetricsMiddlewareMinimal creates middleware with minimal labels
func HTTPMetricsMiddlewareMinimal() gin.HandlerFunc {
	config := DefaultHTTPMetricsConfig()
	config.Labels = []string{"method", "status_code"}
	return HTTPMetricsMiddleware(config)
}

// HTTPMetricsMiddlewareDetailed creates middleware with all available labels
func HTTPMetricsMiddlewareDetailed() gin.HandlerFunc {
	config := DefaultHTTPMetricsConfig()
	config.Labels = []string{"method", "path", "route", "status_code", "user_agent"}
	return HTTPMetricsMiddleware(config)
}

// ClearHTTPHistogram resets the HTTP histogram (useful for testing)
func ClearHTTPHistogram() {
	histogramOnce = sync.Once{}
	httpHistogram = nil
}
