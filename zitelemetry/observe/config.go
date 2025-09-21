package observe

import (
	"time"
)

// Config holds OpenTelemetry observability configuration
type Config struct {
	Service     ServiceConfig `json:"service" yaml:"service"`
	Environment string        `json:"environment" yaml:"environment"`
	Tracing     TracingConfig `json:"tracing" yaml:"tracing"`
	Metrics     MetricsConfig `json:"metrics" yaml:"metrics"`
}

type ServiceConfig struct {
	Name string `json:"name" yaml:"name"`
}

// TracingConfig holds tracing configuration
type TracingConfig struct {
	Enabled  bool           `json:"enabled" yaml:"enabled"`
	Exporter ExporterConfig `json:"exporter" yaml:"exporter"`
	Sampler  SamplerConfig  `json:"sampler" yaml:"sampler"`
	Batch    BatchConfig    `json:"batch" yaml:"batch"`
}

// MetricsConfig holds metrics configuration
type MetricsConfig struct {
	Enabled  bool           `json:"enabled" yaml:"enabled"`
	Exporter ExporterConfig `json:"exporter" yaml:"exporter"`
	Reader   ReaderConfig   `json:"reader" yaml:"reader"`
}

// ExporterConfig holds exporter configuration
type ExporterConfig struct {
	Type     string            `json:"type" yaml:"type"` // "otlp", "jaeger", "console", "none"
	Endpoint string            `json:"endpoint" yaml:"endpoint"`
	Protocol string            `json:"protocol" yaml:"protocol"` // "grpc", "http"
	Headers  map[string]string `json:"headers" yaml:"headers"`
	Insecure bool              `json:"insecure" yaml:"insecure"`
	Timeout  time.Duration     `json:"timeout" yaml:"timeout"`
}

// SamplerConfig holds sampling configuration
type SamplerConfig struct {
	Type     string  `json:"type" yaml:"type"`         // "always_on", "always_off", "traceidratio", "parentbased"
	Fraction float64 `json:"fraction" yaml:"fraction"` // for traceidratio sampler
}

// BatchConfig holds batch processing configuration
type BatchConfig struct {
	MaxExportBatchSize int           `json:"max_export_batch_size" yaml:"max_export_batch_size"`
	ExportTimeout      time.Duration `json:"export_timeout" yaml:"export_timeout"`
	MaxQueueSize       int           `json:"max_queue_size" yaml:"max_queue_size"`
}

// ReaderConfig holds metrics reader configuration
type ReaderConfig struct {
	Interval time.Duration `json:"interval" yaml:"interval"`
	Timeout  time.Duration `json:"timeout" yaml:"timeout"`
}
