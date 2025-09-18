package ziconffx

import (
	"github.com/divikraf/lumos/ziconf"
	"go.uber.org/fx"
)

func WithConfig[T ziconf.Config]() fx.Option {
	return fx.Provide(
		func() *T {
			return ziconf.ReadConfig[T]()
		},
		func(x *T) ziconf.Config {
			return *x
		},
	)
}
