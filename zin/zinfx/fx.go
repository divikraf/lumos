package zinfx

import (
	"github.com/divikraf/lumos/zin"
	"go.uber.org/fx"
)

var Provider = fx.Provide(zin.RegiterRouter)

var Invoker = fx.Invoke(zin.StartHttpServer)
