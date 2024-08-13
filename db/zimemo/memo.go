package zimemo

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"strings"

	lru "github.com/hashicorp/golang-lru/v2"
	"github.com/jmoiron/sqlx"
	"github.com/newrelic/go-agent/v3/newrelic"
	"gitlab.com/divikraf/lumos/zilog"
)

var (
	errCacheMiss = errors.New("cached is missing")
	metricName   = "zimemo.prepared_statement_created"
)

// ZiMemoization is the API for github.com/jmoiron/sqlx prepared statement memoization strategy.
type ZiMemoization interface {
	// Prepare checks cache for memoized statement by its plain query and returns it.
	// If no cache found then it will create new prepared statement, save into cache, and then returns it.
	Prepare(ctx context.Context, db *sqlx.DB, query string) (statement, error)
	// Prepare checks cache for memoized statement by its plain query and returns it.
	// If no cache found then it will create new prepared statement, save into cache, and then returns it.
	PrepareNamed(ctx context.Context, db *sqlx.DB, query string) (statement, error)
	// Purge closes all cached statements.
	Purge()
}

// New create a ZiMemoization instance.
func New(size int, app *newrelic.Application) ZiMemoization {
	l, _ := lru.NewWithEvict(size, stmtEvictionStrategy)
	return &ziMemoizationImpl{
		storage: l,
		nrApp:   app,
	}
}

func stmtEvictionStrategy(key string, value statement) {
	if value.Stmt != nil {
		if err := value.Stmt.Close(); err != nil {
			zilog.DefaultLogger.Warn().Err(err).Msg(err.Error())
		}
	}
	if value.NamedStmt != nil {
		if err := value.NamedStmt.Close(); err != nil {
			zilog.DefaultLogger.Warn().Err(err).Msg(err.Error())
		}
	}
}

// statement is a wrapper that holds the memoized statement.
type statement struct {
	key       string
	Stmt      *sqlx.Stmt
	NamedStmt *sqlx.NamedStmt // "named" variant of the stmt
	Query     string          // the underlying query of the statement
}

// sqlxMemoizationImpl is the implementation for SQLXMemoization.
type ziMemoizationImpl struct {
	storage *lru.Cache[string, statement]
	nrApp   *newrelic.Application
}

func hash(text string) string {
	algorithm := md5.New()
	algorithm.Write([]byte(text))
	return hex.EncodeToString(algorithm.Sum(nil))
}

func (impl *ziMemoizationImpl) getCachedStmt(query string) (statement, error) {
	key := hash(query)
	existing, found := impl.storage.Get(key)
	if !found {
		return statement{key: key}, errCacheMiss
	}

	// if query is equivalent then just yield the existing
	if strings.EqualFold(existing.Query, query) {
		return existing, nil
	}

	// close existing stmt because the query is not equal
	// so that we can safely create another stmt
	err := existing.NamedStmt.Close()
	if err != nil {
		return existing, err
	}
	return statement{key: key}, errCacheMiss
}

// Prepare implements ZiMemoization.
func (impl *ziMemoizationImpl) Prepare(ctx context.Context, db *sqlx.DB, query string) (statement, error) {
	cached, err := impl.getCachedStmt(query)
	if err == nil {
		return cached, nil
	}

	stmt, err := db.PreparexContext(ctx, query)
	if err != nil {
		return cached, err
	}

	cached = statement{
		key:       cached.key,
		Stmt:      stmt,
		NamedStmt: nil,
		Query:     query,
	}
	impl.storage.Add(cached.key, cached)
	impl.nrApp.RecordCustomMetric(metricName, 1)

	return cached, err
}

// PrepareNamed implements ZiMemoization.
func (impl *ziMemoizationImpl) PrepareNamed(ctx context.Context, db *sqlx.DB, query string) (statement, error) {
	cached, err := impl.getCachedStmt(query)
	if err == nil {
		return cached, nil
	}

	// create new named statement from query
	namedStmt, err := db.PrepareNamedContext(ctx, query)
	if err != nil {
		return cached, err
	}

	// memo it
	cached = statement{
		key:       cached.key,
		Stmt:      namedStmt.Stmt,
		NamedStmt: namedStmt,
		Query:     query,
	}
	impl.storage.Add(cached.key, cached)
	impl.nrApp.RecordCustomMetric(metricName, 1)

	return cached, err
}

// Purge implements ZiMemoization.
func (impl *ziMemoizationImpl) Purge() {
	impl.storage.Purge()
}
