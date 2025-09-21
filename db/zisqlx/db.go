package zisqlx

import (
	"context"
	"database/sql"
	"time"

	"github.com/divikraf/lumos/zitelemetry/observe"
	"github.com/divikraf/lumos/zitelemetry/revelio"
	"github.com/jmoiron/sqlx"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

// DB wraps a sqlx.DB to provide metrics and tracing capabilities
type DB struct {
	db                *sqlx.DB
	durationHistogram metric.Int64Histogram
	errorCounter      metric.Int64Counter
}

// New creates a new SQLx wrapper
func New(db *sqlx.DB) *DB {
	durationHistogram := revelio.MustInt64Histogram(
		"database_operation_duration_ms",
		"Duration of database operations in milliseconds",
		metric.WithUnit("ms"),
	)
	errorCounter := revelio.MustInt64Counter(
		"database_operation_errors_total",
		"Number of database operation errors",
	)
	return &DB{
		db:                db,
		durationHistogram: durationHistogram,
		errorCounter:      errorCounter,
	}
}

// Compile-time interface compliance checks
var (
	_ BasicQueryer  = (*DB)(nil)
	_ BasicExecuter = (*DB)(nil)
	_ TxBeginner    = (*DB)(nil)
)

// GetContext executes a query that returns a single row, with metrics and tracing
func (w *DB) GetContext(ctx context.Context, operationName string, dest interface{}, query string, args ...any) error {
	start := time.Now()

	span := w.startSpan(ctx, operationName, "get", query)
	defer span.End()

	var err error
	err = w.db.GetContext(ctx, dest, query, args...)

	duration := time.Since(start)
	w.recordMetrics(ctx, operationName, duration, err)

	if err != nil {
		w.errorCounter.Add(ctx, 1, metric.WithAttributes(attribute.String("error", err.Error())))
	}

	return err
}

// SelectContext executes a query that returns multiple rows, with metrics and tracing
func (w *DB) SelectContext(ctx context.Context, operationName string, dest interface{}, query string, args ...any) error {
	start := time.Now()

	span := w.startSpan(ctx, operationName, "select", query)
	defer span.End()

	var err error
	err = w.db.SelectContext(ctx, dest, query, args...)

	duration := time.Since(start)
	w.recordMetrics(ctx, operationName, duration, err)

	if err != nil {
		w.errorCounter.Add(ctx, 1, metric.WithAttributes(attribute.String("error", err.Error())))
	}

	return err
}

// ExecContext executes a query without returning any rows, with metrics and tracing
func (w *DB) ExecContext(ctx context.Context, operationName string, query string, args ...any) (sql.Result, error) {
	start := time.Now()

	span := w.startSpan(ctx, operationName, "exec", query)
	defer span.End()

	var result sql.Result
	var err error

	result, err = w.db.ExecContext(ctx, query, args...)

	duration := time.Since(start)
	w.recordMetrics(ctx, operationName, duration, err)

	if err != nil {
		w.errorCounter.Add(ctx, 1, metric.WithAttributes(attribute.String("error", err.Error())))
	}

	return result, err
}

// BeginTx starts a new transaction with metrics and tracing
func (w *DB) BeginTx(ctx context.Context, operationName string, opts *sql.TxOptions) (TxInterface, error) {
	start := time.Now()

	span := w.startSpan(ctx, operationName, "begin_tx", "")
	defer span.End()

	tx, err := w.db.BeginTxx(ctx, opts)
	duration := time.Since(start)

	w.recordMetrics(ctx, operationName, duration, err)

	if err != nil {
		w.errorCounter.Add(ctx, 1, metric.WithAttributes(attribute.String("error", err.Error())))
		return nil, err
	}

	return newTx(tx, w.durationHistogram, w.errorCounter), nil
}

// Helper methods

func (w *DB) startSpan(ctx context.Context, operationName, operation, query string) trace.Span {
	ctx, span := observe.FromContext(ctx).Start(ctx, operationName+"."+operation)
	span.SetAttributes(
		attribute.String("db.operation", operation),
		attribute.String("db.operation_name", operationName),
	)

	if query != "" {
		span.SetAttributes(attribute.String("db.statement", query))
	}

	return span
}

func (w *DB) recordMetrics(ctx context.Context, operationName string, duration time.Duration, err error) {
	if w.durationHistogram == nil || w.errorCounter == nil {
		return
	}

	attrs := []attribute.KeyValue{
		attribute.String("operation_name", operationName),
	}

	if err != nil {
		attrs = append(attrs, attribute.String("error", err.Error()))
		w.errorCounter.Add(ctx, 1, metric.WithAttributes(attrs...))
	}

	w.durationHistogram.Record(ctx, duration.Milliseconds(), metric.WithAttributes(attrs...))
}

// GetDB returns the underlying sqlx.DB for advanced usage
func (w *DB) GetDB() *sqlx.DB {
	return w.db
}
