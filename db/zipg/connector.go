package zipg

import (
	"context"
	"fmt"
	"net/url"
	"sync"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
	_ "github.com/newrelic/go-agent/v3/integrations/nrpgx"
	"github.com/rs/zerolog"
)

// New returns connection creator.
func New(validator validator.Validate, logger *zerolog.Logger) *pgConnector {
	return &pgConnector{
		validator: validator,
		logger:    logger,
		conns:     &sync.Map{},
	}
}

type HostPort struct {
	Host string
	Post string `validate:"required"`
}

func (hp HostPort) String() string {
	return fmt.Sprintf("%s:%s", hp.Host, hp.Post)
}

type ConnectionConfig struct {
	MaxOpen         uint          `validate:"required"`
	MaxIdle         uint          `validate:"required"`
	ConnMaxIdleTime time.Duration `validate:"required"`
	ConnMaxLifeTime time.Duration
}

type Input struct {
	HostPort     HostPort `validate:"required"`
	Username     string   `validate:"required"`
	Password     string   `validate:"required"`
	DatabaseName string   `validate:"required"`
	ConnConfig   ConnectionConfig
	QueryParams  url.Values
}

type pgConnector struct {
	validator validator.Validate
	logger    *zerolog.Logger
	conns     *sync.Map
}

func (pgc *pgConnector) PingAll(ctx context.Context) error {
	var returnErr error
	pgc.conns.Range(func(_, conn any) bool {
		if err := conn.(*sqlx.DB).PingContext(ctx); err != nil {
			pgc.logger.Error().Err(err).
				Msg("failed to ping PostgreSQL database")
			returnErr = err
			return false
		}
		return true
	})
	return returnErr
}

func (pgc *pgConnector) CloseAll() error {
	var returnErr error
	pgc.conns.Range(func(addr, conn any) bool {
		if err := conn.(*sqlx.DB).Close(); err != nil {
			pgc.logger.Error().Err(err).
				Msgf("failed to close PostgreSQL database: %s", addr)
			returnErr = err
			return false
		}
		return true
	})
	return returnErr
}

func (pgc *pgConnector) MustConnect(ctx context.Context, input Input) *sqlx.DB {
	db, err := pgc.Connect(ctx, input)
	if err != nil {
		panic(err)
	}
	return db
}

func (pgc *pgConnector) Connect(ctx context.Context, input Input) (*sqlx.DB, error) {
	errValidate := pgc.validator.StructCtx(ctx, input)
	if errValidate != nil {
		pgc.logger.Error().Err(errValidate).Msg(errValidate.Error())
		return nil, errValidate
	}

	dsn := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(input.Username, input.Password),
		Host:     input.HostPort.String(),
		Path:     input.DatabaseName,
		RawQuery: input.QueryParams.Encode(),
	}

	logger := pgc.logger.With().
		Str("hostport", input.HostPort.String()).
		Str("dbname", input.DatabaseName).
		Interface("queryparams", dsn.Query()).
		Logger()

	sqldb, err := sqlx.Open("nrpgx", dsn.String())
	if err != nil {
		logger.Error().Err(err).Msg(err.Error())
		return nil, err
	}

	sqldb.DB.SetMaxOpenConns(int(input.ConnConfig.MaxOpen))
	sqldb.DB.SetMaxIdleConns(int(input.ConnConfig.MaxIdle))
	sqldb.DB.SetConnMaxLifetime(input.ConnConfig.ConnMaxLifeTime)
	sqldb.DB.SetConnMaxIdleTime(input.ConnConfig.ConnMaxIdleTime)

	pgc.conns.Store(input.HostPort.String(), sqldb)
	return sqldb, nil
}
