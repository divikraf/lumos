package ziredis

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/go-playground/validator/v10"
	nrredis "github.com/newrelic/go-agent/v3/integrations/nrredis-v9"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
)

// New returns connection creator.
func New(validator validator.Validate, logger *zerolog.Logger) *connector {
	return &connector{
		validator: validator,
		logger:    logger,
		conns:     &sync.Map{},
	}
}

type multiHostPort []HostPort

func (mhp multiHostPort) Strings() (res []string) {
	for _, hp := range mhp {
		res = append(res, hp.String())
	}
	return
}

type HostPort struct {
	Host string `validate:"required"`
	Port string `validate:"required"`
}

func (hp HostPort) String() string {
	return fmt.Sprintf("%s:%s", hp.Host, hp.Port)
}

type ConnectionConfig struct {
	UseFIFO     bool
	PoolSize    uint `validate:"required"`
	MinIdleConn uint
	MaxIdleConn uint
	ReadTimeout time.Duration
	PoolTimeout time.Duration
	MaxIdleTime time.Duration
	MaxLifeTime time.Duration `validate:"required"`
}

type connector struct {
	validator validator.Validate
	logger    *zerolog.Logger
	conns     *sync.Map
}

func (c *connector) PingAll(ctx context.Context) error {
	var returnErr error
	c.conns.Range(func(_, conn any) bool {
		if err := conn.(redis.UniversalClient).Ping(ctx).Err(); err != nil {
			c.logger.Error().Err(err).
				Msgf("failed to ping Redis database")
			returnErr = err
			return false
		}
		return true
	})
	return returnErr
}

func (c *connector) CloseAll() error {
	var returnErr error
	c.conns.Range(func(addr, conn any) bool {
		if err := conn.(redis.UniversalClient).Close(); err != nil {
			c.logger.Error().Err(err).
				Msgf("failed to close Redis database: %s", addr)
			returnErr = err
			return false
		}
		return true
	})
	return returnErr
}

type InputSingle struct {
	ClientName string
	HostPort   HostPort `validate:"required"`
	Username   string
	Password   string
	DBNumber   uint
	ConnConfig ConnectionConfig
}

func (c *connector) MustConnectSingle(ctx context.Context, input InputSingle) *redis.Client {
	cl, err := c.ConnectSingle(ctx, input)
	if err != nil {
		panic(err)
	}
	return cl
}

func (rc *connector) ConnectSingle(ctx context.Context, input InputSingle) (*redis.Client, error) {
	errValidate := rc.validator.StructCtx(ctx, input)
	if errValidate != nil {
		rc.logger.Error().Err(errValidate).Msg(errValidate.Error())
		return nil, errValidate
	}

	opt := &redis.Options{
		Addr:                  input.HostPort.String(),
		ClientName:            input.ClientName,
		Username:              input.Username,
		Password:              input.Password,
		DB:                    int(input.DBNumber),
		ReadTimeout:           input.ConnConfig.ReadTimeout,
		ContextTimeoutEnabled: true,
		PoolFIFO:              input.ConnConfig.UseFIFO,
		PoolSize:              int(input.ConnConfig.PoolSize),
		PoolTimeout:           input.ConnConfig.PoolTimeout,
		MinIdleConns:          int(input.ConnConfig.MinIdleConn),
		MaxIdleConns:          int(input.ConnConfig.MaxIdleConn),
		ConnMaxIdleTime:       input.ConnConfig.MaxIdleTime,
		ConnMaxLifetime:       input.ConnConfig.MaxLifeTime,
		DisableIndentity:      true,
	}
	cl := redis.NewClient(opt)

	cl.AddHook(nrredis.NewHook(opt))

	var stor redis.UniversalClient = cl
	rc.conns.Store(input.HostPort.String(), stor)
	return cl, nil
}

type InputCluster struct {
	ClientName string     `validate:"required"`
	HostPorts  []HostPort `validate:"required"`
	Username   string
	Password   string
	ConnConfig ConnectionConfig
}

func (c *connector) MustConnectCluster(ctx context.Context, input InputCluster) *redis.ClusterClient {
	cl, err := c.ConnectCluster(ctx, input)
	if err != nil {
		panic(err)
	}
	return cl
}

func (c *connector) ConnectCluster(ctx context.Context, input InputCluster) (*redis.ClusterClient, error) {
	errValidate := c.validator.StructCtx(ctx, input)
	if errValidate != nil {
		c.logger.Error().Err(errValidate).Msg(errValidate.Error())
		return nil, errValidate
	}

	opt := &redis.ClusterOptions{
		Addrs:                 multiHostPort(input.HostPorts).Strings(),
		ClientName:            input.ClientName,
		Username:              input.Username,
		Password:              input.Password,
		ReadTimeout:           input.ConnConfig.ReadTimeout,
		ContextTimeoutEnabled: true,
		PoolFIFO:              input.ConnConfig.UseFIFO,
		PoolSize:              int(input.ConnConfig.PoolSize),
		PoolTimeout:           input.ConnConfig.PoolTimeout,
		MinIdleConns:          int(input.ConnConfig.MinIdleConn),
		MaxIdleConns:          int(input.ConnConfig.MaxIdleConn),
		ConnMaxIdleTime:       input.ConnConfig.MaxIdleTime,
		ConnMaxLifetime:       input.ConnConfig.MaxLifeTime,
	}

	cl := redis.NewClusterClient(opt)

	cl.AddHook(nrredis.NewHook(nil))

	var stor redis.UniversalClient = cl
	c.conns.Store(strings.Join(multiHostPort(input.HostPorts).Strings(), ","), stor)
	return cl, nil
}
