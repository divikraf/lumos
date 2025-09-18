package zipgfx

import (
	"context"

	"github.com/divikraf/lumos/db/zipg"
	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog"
	"go.uber.org/fx"
)

type Connector interface {
	// Connect attempts to connect to a PostgreSQL database, err is returned if
	// failed to do so.
	Connect(ctx context.Context, input zipg.Input) (*sqlx.DB, error)
	// MustConnect attempts to connect to a PostgreSQL database, then panics if
	// failed to do so.
	MustConnect(ctx context.Context, input zipg.Input) *sqlx.DB
}

type connParams struct {
	fx.In

	LC        fx.Lifecycle
	Validator *validator.Validate
	Logger    *zerolog.Logger
}

var Provider = fx.Provide(
	fx.Annotate(func(params connParams) Connector {
		conn := zipg.New(params.Validator, params.Logger)
		params.LC.Append(fx.StartHook(conn.PingAll))
		params.LC.Append(fx.StopHook(conn.CloseAll))
		return conn
	}, fx.As(new(Connector))),
)
