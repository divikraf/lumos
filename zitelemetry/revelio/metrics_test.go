package revelio

import (
	"context"
	"testing"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

func TestHelperFunctions(t *testing.T) {
	// Test counter creation
	t.Run("Int64Counter", func(t *testing.T) {
		counter, err := Int64Counter("test_counter", "A test counter metric", metric.WithUnit("count"))
		if err != nil {
			t.Fatalf("Failed to create counter: %v", err)
		}

		if counter == nil {
			t.Fatal("Counter should not be nil")
		}

		// Test recording
		ctx := context.Background()
		counter.Add(ctx, 1, metric.WithAttributes(attribute.String("method", "GET")))
	})

	// Test histogram creation
	t.Run("Int64Histogram", func(t *testing.T) {
		histogram, err := Int64Histogram("test_histogram", "A test histogram metric", metric.WithUnit("ms"))
		if err != nil {
			t.Fatalf("Failed to create histogram: %v", err)
		}

		if histogram == nil {
			t.Fatal("Histogram should not be nil")
		}

		// Test recording
		ctx := context.Background()
		histogram.Record(ctx, 100, metric.WithAttributes(attribute.String("operation", "query")))
	})

	// Test float counter creation
	t.Run("Float64Counter", func(t *testing.T) {
		counter, err := Float64Counter("test_float_counter", "A test float counter metric", metric.WithUnit("bytes"))
		if err != nil {
			t.Fatalf("Failed to create float counter: %v", err)
		}

		if counter == nil {
			t.Fatal("Float counter should not be nil")
		}

		// Test recording
		ctx := context.Background()
		counter.Add(ctx, 1.5, metric.WithAttributes(attribute.String("type", "api")))
	})

	// Test gauge creation
	t.Run("Int64Gauge", func(t *testing.T) {
		gauge, err := Int64Gauge("test_gauge", "A test gauge metric", metric.WithUnit("items"))
		if err != nil {
			t.Fatalf("Failed to create gauge: %v", err)
		}

		if gauge == nil {
			t.Fatal("Gauge should not be nil")
		}

		// Test recording
		ctx := context.Background()
		gauge.Record(ctx, 1024, metric.WithAttributes(attribute.String("resource", "memory")))
	})
}

func TestMustFunctions(t *testing.T) {
	// Test MustInt64Counter
	t.Run("MustInt64Counter", func(t *testing.T) {
		counter := MustInt64Counter("must_counter", "Must counter test")
		if counter == nil {
			t.Fatal("Must counter should not be nil")
		}

		// Test recording
		ctx := context.Background()
		counter.Add(ctx, 1, metric.WithAttributes(attribute.String("test", "must")))
	})

	// Test MustInt64Histogram
	t.Run("MustInt64Histogram", func(t *testing.T) {
		histogram := MustInt64Histogram("must_histogram", "Must histogram test")
		if histogram == nil {
			t.Fatal("Must histogram should not be nil")
		}

		// Test recording
		ctx := context.Background()
		histogram.Record(ctx, 50, metric.WithAttributes(attribute.String("test", "must")))
	})
}

func TestDurationInstrument(t *testing.T) {
	t.Run("Duration", func(t *testing.T) {
		duration, err := Duration("test_duration", "Test duration metric")
		if err != nil {
			t.Fatalf("Failed to create duration: %v", err)
		}

		if duration == nil {
			t.Fatal("Duration should not be nil")
		}

		// Test recording
		ctx := context.Background()
		duration.RecordFloat64(ctx, 150.5, attribute.String("operation", "test"))
	})

	t.Run("MustDuration", func(t *testing.T) {
		duration := MustDuration("must_duration", "Must duration test")
		if duration == nil {
			t.Fatal("Must duration should not be nil")
		}

		// Test recording
		ctx := context.Background()
		duration.RecordFloat64(ctx, 200.0, attribute.String("test", "must"))
	})
}

func TestScopeManagement(t *testing.T) {
	t.Run("GetDefault", func(t *testing.T) {
		defaultScope := GetDefault()
		if defaultScope == nil {
			t.Fatal("Default scope should not be nil")
		}
	})

	t.Run("NewFromMeter", func(t *testing.T) {
		meter := otel.GetMeterProvider().Meter("test")
		scope := NewFromMeter(meter)
		if scope == nil {
			t.Fatal("Scope from meter should not be nil")
		}

		// Test that it can create metrics
		counter := MustInt64Counter("scope_test_counter", "Test counter from custom scope")
		if counter == nil {
			t.Fatal("Counter from custom scope should not be nil")
		}
	})
}
