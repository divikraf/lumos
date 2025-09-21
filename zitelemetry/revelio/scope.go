package revelio

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

// Scope represents a named meter that can create metrics
type Scope interface {
	// GetMeter returns the underlying meter
	GetMeter() metric.Meter

	// Duration creates a duration recorder (Float64Histogram with ms unit)
	Duration(name string, description string, options ...DurationOption) (DurationRecorder, error)

	// Standard OpenTelemetry metric creation methods
	Int64Counter(name string, description string, options ...metric.Int64CounterOption) (metric.Int64Counter, error)
	Int64UpDownCounter(name string, description string, options ...metric.Int64UpDownCounterOption) (metric.Int64UpDownCounter, error)
	Int64Histogram(name string, description string, options ...metric.Int64HistogramOption) (metric.Int64Histogram, error)
	Int64Gauge(name string, description string, options ...metric.Int64GaugeOption) (metric.Int64Gauge, error)
	Int64ObservableCounter(name string, description string, options ...metric.Int64ObservableCounterOption) (metric.Int64ObservableCounter, error)
	Int64ObservableUpDownCounter(name string, description string, options ...metric.Int64ObservableUpDownCounterOption) (metric.Int64ObservableUpDownCounter, error)
	Int64ObservableGauge(name string, description string, options ...metric.Int64ObservableGaugeOption) (metric.Int64ObservableGauge, error)

	Float64Counter(name string, description string, options ...metric.Float64CounterOption) (metric.Float64Counter, error)
	Float64UpDownCounter(name string, description string, options ...metric.Float64UpDownCounterOption) (metric.Float64UpDownCounter, error)
	Float64Histogram(name string, description string, options ...metric.Float64HistogramOption) (metric.Float64Histogram, error)
	Float64Gauge(name string, description string, options ...metric.Float64GaugeOption) (metric.Float64Gauge, error)
	Float64ObservableCounter(name string, description string, options ...metric.Float64ObservableCounterOption) (metric.Float64ObservableCounter, error)
	Float64ObservableUpDownCounter(name string, description string, options ...metric.Float64ObservableUpDownCounterOption) (metric.Float64ObservableUpDownCounter, error)
	Float64ObservableGauge(name string, description string, options ...metric.Float64ObservableGaugeOption) (metric.Float64ObservableGauge, error)

	// Callback registration
	RegisterCallback(f metric.Callback, instruments ...metric.Observable) (metric.Registration, error)
}

// scope is the implementation of Scope interface
type scope struct {
	meter metric.Meter
}

// GetMeter returns the underlying meter
func (s *scope) GetMeter() metric.Meter {
	return s.meter
}

// Duration creates a duration recorder (Float64Histogram with ms unit)
func (s *scope) Duration(name string, description string, options ...DurationOption) (DurationRecorder, error) {
	opts := []metric.Float64HistogramOption{
		metric.WithDescription(description),
		metric.WithUnit("ms"),
	}

	// Apply custom options
	for _, opt := range options {
		opts = append(opts, opt.toFloat64HistogramOption())
	}

	histogram, err := s.meter.Float64Histogram(name, opts...)
	if err != nil {
		return nil, err
	}

	return &durationRecorder{
		histogram: histogram,
	}, nil
}

// Standard metric creation methods delegate to the underlying meter
func (s *scope) Int64Counter(name string, description string, options ...metric.Int64CounterOption) (metric.Int64Counter, error) {
	opts := append([]metric.Int64CounterOption{metric.WithDescription(description)}, options...)
	return s.meter.Int64Counter(name, opts...)
}

func (s *scope) Int64UpDownCounter(name string, description string, options ...metric.Int64UpDownCounterOption) (metric.Int64UpDownCounter, error) {
	opts := append([]metric.Int64UpDownCounterOption{metric.WithDescription(description)}, options...)
	return s.meter.Int64UpDownCounter(name, opts...)
}

func (s *scope) Int64Histogram(name string, description string, options ...metric.Int64HistogramOption) (metric.Int64Histogram, error) {
	opts := append([]metric.Int64HistogramOption{metric.WithDescription(description)}, options...)
	return s.meter.Int64Histogram(name, opts...)
}

func (s *scope) Int64Gauge(name string, description string, options ...metric.Int64GaugeOption) (metric.Int64Gauge, error) {
	opts := append([]metric.Int64GaugeOption{metric.WithDescription(description)}, options...)
	return s.meter.Int64Gauge(name, opts...)
}

func (s *scope) Int64ObservableCounter(name string, description string, options ...metric.Int64ObservableCounterOption) (metric.Int64ObservableCounter, error) {
	opts := append([]metric.Int64ObservableCounterOption{metric.WithDescription(description)}, options...)
	return s.meter.Int64ObservableCounter(name, opts...)
}

func (s *scope) Int64ObservableUpDownCounter(name string, description string, options ...metric.Int64ObservableUpDownCounterOption) (metric.Int64ObservableUpDownCounter, error) {
	opts := append([]metric.Int64ObservableUpDownCounterOption{metric.WithDescription(description)}, options...)
	return s.meter.Int64ObservableUpDownCounter(name, opts...)
}

func (s *scope) Int64ObservableGauge(name string, description string, options ...metric.Int64ObservableGaugeOption) (metric.Int64ObservableGauge, error) {
	opts := append([]metric.Int64ObservableGaugeOption{metric.WithDescription(description)}, options...)
	return s.meter.Int64ObservableGauge(name, opts...)
}

func (s *scope) Float64Counter(name string, description string, options ...metric.Float64CounterOption) (metric.Float64Counter, error) {
	opts := append([]metric.Float64CounterOption{metric.WithDescription(description)}, options...)
	return s.meter.Float64Counter(name, opts...)
}

func (s *scope) Float64UpDownCounter(name string, description string, options ...metric.Float64UpDownCounterOption) (metric.Float64UpDownCounter, error) {
	opts := append([]metric.Float64UpDownCounterOption{metric.WithDescription(description)}, options...)
	return s.meter.Float64UpDownCounter(name, opts...)
}

func (s *scope) Float64Histogram(name string, description string, options ...metric.Float64HistogramOption) (metric.Float64Histogram, error) {
	opts := append([]metric.Float64HistogramOption{metric.WithDescription(description)}, options...)
	return s.meter.Float64Histogram(name, opts...)
}

func (s *scope) Float64Gauge(name string, description string, options ...metric.Float64GaugeOption) (metric.Float64Gauge, error) {
	opts := append([]metric.Float64GaugeOption{metric.WithDescription(description)}, options...)
	return s.meter.Float64Gauge(name, opts...)
}

func (s *scope) Float64ObservableCounter(name string, description string, options ...metric.Float64ObservableCounterOption) (metric.Float64ObservableCounter, error) {
	opts := append([]metric.Float64ObservableCounterOption{metric.WithDescription(description)}, options...)
	return s.meter.Float64ObservableCounter(name, opts...)
}

func (s *scope) Float64ObservableUpDownCounter(name string, description string, options ...metric.Float64ObservableUpDownCounterOption) (metric.Float64ObservableUpDownCounter, error) {
	opts := append([]metric.Float64ObservableUpDownCounterOption{metric.WithDescription(description)}, options...)
	return s.meter.Float64ObservableUpDownCounter(name, opts...)
}

func (s *scope) Float64ObservableGauge(name string, description string, options ...metric.Float64ObservableGaugeOption) (metric.Float64ObservableGauge, error) {
	opts := append([]metric.Float64ObservableGaugeOption{metric.WithDescription(description)}, options...)
	return s.meter.Float64ObservableGauge(name, opts...)
}

func (s *scope) RegisterCallback(f metric.Callback, instruments ...metric.Observable) (metric.Registration, error) {
	return s.meter.RegisterCallback(f, instruments...)
}

// DurationRecorder is a specialized recorder for duration measurements
type DurationRecorder interface {
	// Record records a duration measurement
	Record(ctx context.Context, duration time.Duration, attrs ...attribute.KeyValue)
	// RecordFloat64 records a duration measurement as float64 milliseconds
	RecordFloat64(ctx context.Context, durationMs float64, attrs ...attribute.KeyValue)
}

// durationRecorder is the implementation of DurationRecorder
type durationRecorder struct {
	histogram metric.Float64Histogram
}

// Record records a duration measurement
func (dr *durationRecorder) Record(ctx context.Context, duration time.Duration, attrs ...attribute.KeyValue) {
	dr.histogram.Record(ctx, float64(duration.Milliseconds()), metric.WithAttributes(attrs...))
}

// RecordFloat64 records a duration measurement as float64 milliseconds
func (dr *durationRecorder) RecordFloat64(ctx context.Context, durationMs float64, attrs ...attribute.KeyValue) {
	dr.histogram.Record(ctx, durationMs, metric.WithAttributes(attrs...))
}

// DurationOption is an option for configuring Duration instruments
type DurationOption interface {
	toFloat64HistogramOption() metric.Float64HistogramOption
}

// WithUnit sets the unit for a Duration instrument
func WithUnit(unit string) DurationOption {
	return unitOption{unit: unit}
}

type unitOption struct {
	unit string
}

func (u unitOption) toFloat64HistogramOption() metric.Float64HistogramOption {
	return metric.WithUnit(u.unit)
}

// WithExplicitBucketBoundaries sets explicit bucket boundaries for a Duration instrument
func WithExplicitBucketBoundaries(boundaries ...float64) DurationOption {
	return bucketBoundariesOption{boundaries: boundaries}
}

type bucketBoundariesOption struct {
	boundaries []float64
}

func (b bucketBoundariesOption) toFloat64HistogramOption() metric.Float64HistogramOption {
	return metric.WithExplicitBucketBoundaries(b.boundaries...)
}
