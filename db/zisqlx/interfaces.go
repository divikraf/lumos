package zisqlx

import (
	"context"
	"database/sql"
)

type BasicQueryerExecuter interface {
	BasicQueryer
	BasicExecuter
	TxBeginner
}

type BasicQueryer interface {

	// GetContext executes a query that returns a single row and scans it into dest
	GetContext(ctx context.Context, operationName string, dest interface{}, query string, args ...any) error

	// SelectContext executes a query that returns multiple rows and scans them into dest
	SelectContext(ctx context.Context, operationName string, dest interface{}, query string, args ...any) error
}

type BasicExecuter interface {
	// ExecContext executes a query without returning any rows
	ExecContext(ctx context.Context, operationName string, query string, args ...any) (sql.Result, error)
}

type TxBeginner interface {
	// BeginTx starts a new transaction
	BeginTx(ctx context.Context, operationName string, opts *sql.TxOptions) (TxInterface, error)
}

type TxInterface interface {
	BasicQueryer
	BasicExecuter

	Commit() error

	Rollback() error
}
