package observefx

import (
	"context"

	"github.com/divikraf/lumos/ziconf"
	"github.com/divikraf/lumos/zitelemetry/observe"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/fx"
)

// Module provides OpenTelemetry observability components
var Module = fx.Module("observe",
	fx.Provide(
		provideTelemetry,
		provideTracer,
	),
	fx.Invoke(registerShutdown),
)

// provideTelemetry creates a Telemetry instance
func provideTelemetry(lc fx.Lifecycle, config ziconf.Config) *observe.Telemetry {
	ctx := context.Background()

	tel, err := observe.New(ctx, config.GetTelemetry())
	if err != nil {
		panic(err)
	}

	lc.Append(fx.Hook{
		OnStop: tel.Shutdown,
	})

	return tel
}

// provideTracer provides the global tracer
func provideTracer() trace.Tracer {
	return otel.Tracer("lumos")
}

// registerShutdown ensures proper shutdown
func registerShutdown(tel *observe.Telemetry, lc fx.Lifecycle) {
	lc.Append(fx.Hook{
		OnStop: tel.Shutdown,
	})
}
