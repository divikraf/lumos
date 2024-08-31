package zilong

import (
	"context"
	"os"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/newrelic/go-agent/v3/newrelic"
	"gitlab.com/divikraf/lumos/db/zimemo/zimemofx"
	"gitlab.com/divikraf/lumos/db/zipg/zipgfx"
	"gitlab.com/divikraf/lumos/db/ziredis/ziredisfx"
	"gitlab.com/divikraf/lumos/ziconf"
	"gitlab.com/divikraf/lumos/ziconf/ziconffx"
	"gitlab.com/divikraf/lumos/zilog"
	"gitlab.com/divikraf/lumos/zilog/zilogfx"
	"gitlab.com/divikraf/lumos/zin/zinfx"
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

func newrelicFx(lc fx.Lifecycle, config ziconf.Config) *newrelic.Application {
	nrApp, err := newrelic.NewApplication(
		newrelic.ConfigAppName(config.GetService().Name),
		newrelic.ConfigLicense(config.GetNewRelic().LicenseKey),
		newrelic.ConfigInfoLogger(os.Stdout),
		newrelic.ConfigAppLogForwardingEnabled(true),
	)
	if err != nil {
		zilog.DefaultLogger.Fatal().Err(err).Msg("fail to init new relic application")
	}
	lc.Append(fx.StopHook(func() {
		nrApp.Shutdown(3 * time.Second)
	}))
	return nrApp
}

var NewRelicProvider = fx.Provide(newrelicFx)

func KitchenSink[T ziconf.Config]() []fx.Option {
	return []fx.Option{
		ContextProvider,
		ValidatorProvider,
		ziconffx.WithConfig[T](),
		zilogfx.FxLogger,
		zilogfx.ContextDecorator,
		zilogfx.Provider,
		NewRelicProvider,
		zipgfx.Provider,
		ziredisfx.Provider,
		zimemofx.Provider,
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
