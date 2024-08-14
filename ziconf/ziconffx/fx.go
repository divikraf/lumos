package ziconffx

import (
	"gitlab.com/divikraf/lumos/ziconf"
	"go.uber.org/fx"
)

func WithConfig[T ziconf.Config]() fx.Option {
	return fx.Provide(func() T {
		return ziconf.ReadConfig[T]()
	})
}
