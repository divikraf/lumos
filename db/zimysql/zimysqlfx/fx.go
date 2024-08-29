package zimysqlfx

import (
	"context"

	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog"
	"gitlab.com/divikraf/lumos/db/zimysql"
	"go.uber.org/fx"
)

type Connector interface {
	// Connect attempts to connect to a MySQL database, err is returned if
	// failed to do so.
	Connect(ctx context.Context, input zimysql.Input) (*sqlx.DB, error)
	// MustConnect attempts to connect to a MySQL database, then panics if
	// failed to do so.
	MustConnect(ctx context.Context, input zimysql.Input) *sqlx.DB
}

type connParams struct {
	fx.In

	LC        fx.Lifecycle
	Validator *validator.Validate
	Logger    *zerolog.Logger
}

var Provider = fx.Provide(
	fx.Annotate(func(params connParams) Connector {
		conn := zimysql.New(params.Validator, params.Logger)
		params.LC.Append(fx.StartHook(conn.PingAll))
		params.LC.Append(fx.StopHook(conn.CloseAll))
		return conn
	}, fx.As(new(Connector))),
)
