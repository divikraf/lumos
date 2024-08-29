package zimysql

import (
	"context"
	"fmt"
	"net/url"
	"sync"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
	_ "github.com/newrelic/go-agent/v3/integrations/nrmysql"
	"github.com/rs/zerolog"
)

// New returns connection creator.
func New(validator *validator.Validate, logger *zerolog.Logger) *mysqlConnector {
	return &mysqlConnector{
		validator: validator,
		logger:    logger,
		conns:     &sync.Map{},
	}
}

type HostPort struct {
	Host string
	Port string `validate:"required"`
}

func (hp HostPort) String() string {
	return fmt.Sprintf("%s:%s", hp.Host, hp.Port)
}

type ConnectionConfig struct {
	MaxOpen         uint          `validate:"required"`
	MaxIdle         uint          `validate:"required"`
	ConnMaxIdleTime time.Duration `validate:"required"`
	ConnMaxLifetime time.Duration
}

type Input struct {
	HostPort     HostPort         `validate:"required"`
	Username     string           `validate:"required"`
	Password     string           `validate:"required"`
	DatabaseName string           `validate:"required"`
	ConnConfig   ConnectionConfig `validate:"required"`
	QueryParams  url.Values
}

type mysqlConnector struct {
	validator *validator.Validate
	logger    *zerolog.Logger
	conns     *sync.Map
}

func (myc *mysqlConnector) PingAll(ctx context.Context) error {
	var returnErr error
	myc.conns.Range(func(_, conn any) bool {
		if err := conn.(*sqlx.DB).PingContext(ctx); err != nil {
			myc.logger.Error().Err(err).
				Msg("failed to ping MySQL database")
			returnErr = err
			return false
		}
		return true
	})
	return returnErr
}

func (myc *mysqlConnector) CloseAll() error {
	var returnErr error
	myc.conns.Range(func(addr, conn any) bool {
		if err := conn.(*sqlx.DB).Close(); err != nil {
			myc.logger.Error().Err(err).
				Msgf("failed to close MySQL database: %s", addr)
			returnErr = err
			return false
		}
		return true
	})
	return returnErr
}

func (myc *mysqlConnector) MustConnect(ctx context.Context, input Input) *sqlx.DB {
	db, err := myc.Connect(ctx, input)
	if err != nil {
		panic(err)
	}
	return db
}

func (myc *mysqlConnector) Connect(ctx context.Context, input Input) (*sqlx.DB, error) {
	errValidate := myc.validator.StructCtx(ctx, input)
	if errValidate != nil {
		myc.logger.Error().Err(errValidate).Msg(errValidate.Error())
		return nil, errValidate
	}

	queryParams := url.Values{}
	for key, values := range map[string][]string(input.QueryParams) {
		queryParams.Del(key) // reset
		for _, value := range values {
			queryParams.Add(key, value)
		}
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?%s", input.Username, input.Password, input.HostPort.String(), input.DatabaseName, queryParams.Encode())

	logger := myc.logger.With().
		Str("hostport", input.HostPort.String()).
		Str("dbname", input.DatabaseName).
		Interface("queryparams", queryParams).
		Logger()

	sqldb, err := sqlx.Open("nrmysql", dsn)
	if err != nil {
		logger.Error().Err(err).Msg(err.Error())
		return nil, err
	}

	sqldb.DB.SetMaxOpenConns(int(input.ConnConfig.MaxOpen))
	sqldb.DB.SetMaxIdleConns(int(input.ConnConfig.MaxIdle))
	sqldb.DB.SetConnMaxLifetime(input.ConnConfig.ConnMaxLifetime)
	sqldb.DB.SetConnMaxIdleTime(input.ConnConfig.ConnMaxIdleTime)

	myc.conns.Store(input.HostPort.String(), sqldb)
	return sqldb, nil
}
