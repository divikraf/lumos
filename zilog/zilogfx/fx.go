package zilogfx

import (
	"context"
	"log/slog"
	"os"

	"github.com/divikraf/lumos/zilog"
	"github.com/rs/zerolog"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
)

var Provider = fx.Provide(
	slog.Default,
	func() *zerolog.Logger {
		return &zilog.DefaultLogger.Logger
	},
)

type useConsoleLogger bool

type fxLogParams struct {
	fx.In

	DisableSlog useConsoleLogger `optional:"true"`
	L           *slog.Logger
}

// UseConsoleLogger sets Uber Fx framework logger to a simple console logger
// instead of the default JSON logger. This option might be useful when developing
// locally.
var UseConsoleLogger = fx.Provide(
	func() useConsoleLogger {
		return useConsoleLogger(true)
	},
)

// FxLogger is a Logger that may be used for fx.App
var FxLogger = fx.WithLogger(func(params fxLogParams) fxevent.Logger {
	if !params.DisableSlog {
		return &SlogLogger{
			Logger: params.L,
		}
	}
	return &fxevent.ConsoleLogger{
		W: os.Stdout,
	}
})

// ContextDecorator decorates a context.Context with a Logger from the provided
// Logger. This allows the Logger to be propagated to all dependencies.
var ContextDecorator = fx.Decorate(
	func(ctx context.Context, logger *zerolog.Logger) context.Context {
		return logger.WithContext(ctx)
	},
)
