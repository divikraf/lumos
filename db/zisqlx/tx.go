package zisqlx

import (
	"context"
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

// TxWrapper wraps a sqlx.Tx to provide metrics and tracing capabilities
type TxWrapper struct {
	tx                *sqlx.Tx
	durationHistogram metric.Int64Histogram
	errorCounter      metric.Int64Counter
}

// newTx creates a new transaction wrapper
func newTx(tx *sqlx.Tx, durationHistogram metric.Int64Histogram, errorCounter metric.Int64Counter) *TxWrapper {
	return &TxWrapper{
		tx:                tx,
		durationHistogram: durationHistogram,
		errorCounter:      errorCounter,
	}
}

// Compile-time interface compliance checks
var (
	_ BasicQueryer  = (*TxWrapper)(nil)
	_ BasicExecuter = (*TxWrapper)(nil)
	_ TxInterface   = (*TxWrapper)(nil)
)

// GetContext executes a query that returns a single row, with metrics and tracing
func (t *TxWrapper) GetContext(ctx context.Context, operationName string, dest interface{}, query string, args ...any) error {
	start := time.Now()

	span := t.startSpan(ctx, operationName, "get", query)
	defer span.End()

	var err error
	err = t.tx.GetContext(ctx, dest, query, args...)

	duration := time.Since(start)
	t.recordMetrics(ctx, operationName, duration, err)
	t.logQuery(ctx, operationName, query, args, duration, err)

	return err
}

// SelectContext executes a query that returns multiple rows, with metrics and tracing
func (t *TxWrapper) SelectContext(ctx context.Context, operationName string, dest interface{}, query string, args ...any) error {
	start := time.Now()

	span := t.startSpan(ctx, operationName, "select", query)
	defer span.End()

	var err error
	err = t.tx.SelectContext(ctx, dest, query, args...)

	duration := time.Since(start)
	t.recordMetrics(ctx, operationName, duration, err)
	t.logQuery(ctx, operationName, query, args, duration, err)

	return err
}

// ExecContext executes a query without returning any rows, with metrics and tracing
func (t *TxWrapper) ExecContext(ctx context.Context, operationName string, query string, args ...any) (sql.Result, error) {
	start := time.Now()

	span := t.startSpan(ctx, operationName, "exec", query)
	defer span.End()

	var result sql.Result
	var err error

	result, err = t.tx.ExecContext(ctx, query, args...)

	duration := time.Since(start)
	t.recordMetrics(ctx, operationName, duration, err)
	t.logQuery(ctx, operationName, query, args, duration, err)

	return result, err
}

// Commit commits the transaction with metrics and tracing
func (t *TxWrapper) Commit() error {
	start := time.Now()

	span := t.startSpan(context.Background(), "commit", "tx_commit", "")
	defer span.End()

	err := t.tx.Commit()
	duration := time.Since(start)

	t.recordMetrics(context.Background(), "commit", duration, err)
	t.logOperation(context.Background(), "commit", "tx_commit", duration, err)

	return err
}

// Rollback rolls back the transaction with metrics and tracing
func (t *TxWrapper) Rollback() error {
	start := time.Now()

	span := t.startSpan(context.Background(), "rollback", "tx_rollback", "")
	defer span.End()

	err := t.tx.Rollback()
	duration := time.Since(start)

	t.recordMetrics(context.Background(), "rollback", duration, err)
	t.logOperation(context.Background(), "rollback", "tx_rollback", duration, err)

	return err
}

// GetTx returns the underlying sqlx.Tx for advanced usage
func (t *TxWrapper) GetTx() *sqlx.Tx {
	return t.tx
}

// Helper methods

func (t *TxWrapper) startSpan(ctx context.Context, operationName, operation, query string) trace.Span {
	tracer := trace.SpanFromContext(ctx).TracerProvider()
	if tracer == nil {
		return trace.SpanFromContext(ctx)
	}

	// Get service name from context logger if available
	serviceName := "unknown"
	// Try to get service name from span attributes if available
	if span := trace.SpanFromContext(ctx); span != nil {
		// For now, use a default service name
		// In a real implementation, you might want to pass this as a parameter
		serviceName = "database-service"
	}

	spanName := serviceName + "." + operationName + "." + operation
	ctx, span := tracer.Tracer("zisqlx").Start(ctx, spanName)

	span.SetAttributes(
		attribute.String("db.operation", operation),
		attribute.String("db.operation_name", operationName),
		attribute.Bool("db.transaction", true),
	)

	if query != "" {
		span.SetAttributes(attribute.String("db.statement", query))
	}

	return span
}

func (t *TxWrapper) recordMetrics(ctx context.Context, operationName string, duration time.Duration, err error) {
	if t.durationHistogram == nil || t.errorCounter == nil {
		return
	}

	attrs := []attribute.KeyValue{
		attribute.String("operation_name", operationName),
		attribute.Bool("transaction", true),
	}

	if err != nil {
		attrs = append(attrs, attribute.String("error", err.Error()))
		t.errorCounter.Add(ctx, 1, metric.WithAttributes(attrs...))
	}

	t.durationHistogram.Record(ctx, duration.Milliseconds(), metric.WithAttributes(attrs...))
}

func (t *TxWrapper) logQuery(ctx context.Context, operationName, query string, args []any, duration time.Duration, err error) {
	// Logging can be added here if needed
	// For now, we'll rely on metrics and tracing
}

func (t *TxWrapper) logOperation(ctx context.Context, operationName, operation string, duration time.Duration, err error) {
	// Logging can be added here if needed
	// For now, we'll rely on metrics and tracing
}
