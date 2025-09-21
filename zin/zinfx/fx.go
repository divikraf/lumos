package zinfx

import (
	"github.com/divikraf/lumos/zin"
	"go.uber.org/fx"
)

var Provider = fx.Provide(zin.RegiterRouter)

var Invoker = fx.Invoke(zin.StartHttpServer)

// SkipPathProvider provides skip paths for HTTP metrics
type SkipPathProvider struct {
	fx.Out
	SkipPaths []string `group:"http-metrics-skip-paths"`
}

// AddSkipPaths adds skip paths for HTTP metrics
func AddSkipPaths(paths ...string) fx.Option {
	return fx.Provide(func() SkipPathProvider {
		return SkipPathProvider{
			SkipPaths: paths,
		}
	})
}
