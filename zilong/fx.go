package zilong

import (
	"context"

	"github.com/divikraf/lumos/db/zimysql/zimysqlfx"
	"github.com/divikraf/lumos/db/zipg/zipgfx"
	"github.com/divikraf/lumos/db/ziredis/ziredisfx"
	"github.com/divikraf/lumos/ziconf"
	"github.com/divikraf/lumos/ziconf/ziconffx"
	"github.com/divikraf/lumos/zilog/zilogfx"
	"github.com/divikraf/lumos/zin/zinfx"
	"github.com/divikraf/lumos/zitelemetry/observe/observefx"
	"github.com/divikraf/lumos/zitelemetry/revelio/reveliofx"
	"github.com/divikraf/lumos/zivalidator/zivalidatorfx"
	"github.com/go-playground/validator/v10"
	"go.uber.org/fx"
)

func contextFx(lc fx.Lifecycle) context.Context {
	ct, cancel := context.WithCancel(context.Background())
	lc.Append(fx.StopHook(cancel))
	return ct
}

// ContextProvider provides a cancelable context.Context instance. It creates a
// new context.Context with [context.WithCancel] and attaches the cancel function
// to the [go.uber.org/fx.Lifecycle]. When the fx app is stopped, this cancel
// function will be called, canceling the context.
var ContextProvider = fx.Provide(contextFx)

func validatorFx() *validator.Validate {
	validator := validator.New()
	return validator
}

var ValidatorProvider = fx.Provide(validatorFx)

func KitchenSink[T ziconf.Config]() []fx.Option {
	return []fx.Option{
		ContextProvider,
		ValidatorProvider,
		ziconffx.WithConfig[T](),
		observefx.Module,
		reveliofx.DefaultScopeProvider,
		reveliofx.MeterProviderProvider,
		zilogfx.FxLogger,
		zilogfx.ContextDecorator,
		zilogfx.Provider,
		zipgfx.Provider,
		zimysqlfx.Provider,
		ziredisfx.Provider,
		zivalidatorfx.Provider,
		zinfx.Provider,
	}
}

func New[T ziconf.Config](subModules ...fx.Option) []fx.Option {
	return append(KitchenSink[T](), subModules...)
}

func App[T ziconf.Config](modules ...fx.Option) *fx.App {
	return fx.New(
		New[T](modules...)...,
	)
}
