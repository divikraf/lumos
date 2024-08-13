package ziredisfx

import (
	"context"

	"github.com/go-playground/validator/v10"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	"gitlab.com/divikraf/lumos/db/ziredis"
	"go.uber.org/fx"
)

type Connector interface {
	SingleConnector
	ClusterConnector
}

type SingleConnector interface {
	// ConnectSingle attempts to connect to a single Redis database, err is
	// returned if failed to do so.
	ConnectSingle(ctx context.Context, input ziredis.InputSingle) (*redis.Client, error)
	// MustConnectSingle attempts to connect to a single Redis database, then
	// panics if failed to do so.
	MustConnectSingle(ctx context.Context, input ziredis.InputSingle) *redis.Client
}

type ClusterConnector interface {
	// ConnectCluster attempts to connect to Redis cluster database, err is
	// returned if failed to do so.
	ConnectCluster(ctx context.Context, input ziredis.InputCluster) (*redis.ClusterClient, error)
	// MustConnectCluster attempts to connect to Redis cluster database, then
	// panics if failed to do so.
	MustConnectCluster(ctx context.Context, input ziredis.InputCluster) *redis.ClusterClient
}

type connParams struct {
	fx.In

	LC        fx.Lifecycle
	Validator validator.Validate
	Logger    *zerolog.Logger
}

type fxResult struct {
	fx.Out

	All     Connector
	Single  SingleConnector
	Cluster ClusterConnector
}

var Provider = fx.Provide(
	func(params connParams) fxResult {
		conn := ziredis.New(params.Validator, params.Logger)
		params.LC.Append(fx.StartHook(conn.PingAll))
		params.LC.Append(fx.StopHook(conn.CloseAll))
		return fxResult{
			All:     conn,
			Single:  conn,
			Cluster: conn,
		}
	},
)
