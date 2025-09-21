package reveliofx

import (
	"github.com/divikraf/lumos/zitelemetry/revelio"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
	"go.uber.org/fx"
)

// ScopeParams holds dependencies for scope provider
type ScopeParams struct {
	fx.In
}

// ScopeResult holds the scope result
type ScopeResult struct {
	fx.Out

	Scope revelio.Scope
}

// DefaultScopeProvider provides the default scope
var DefaultScopeProvider = fx.Provide(
	func(params ScopeParams) ScopeResult {
		// Get the global default scope
		scope := revelio.GetDefault()

		return ScopeResult{
			Scope: scope,
		}
	},
)

// MeterProviderParams holds dependencies for meter provider
type MeterProviderParams struct {
	fx.In
}

// MeterProviderResult holds the meter provider result
type MeterProviderResult struct {
	fx.Out

	MeterProvider metric.MeterProvider
}

// MeterProviderProvider provides the OpenTelemetry meter provider
var MeterProviderProvider = fx.Provide(
	func(params MeterProviderParams) MeterProviderResult {
		// Get the global OpenTelemetry meter provider
		meterProvider := otel.GetMeterProvider()

		return MeterProviderResult{
			MeterProvider: meterProvider,
		}
	},
)

// WithCustomScope provides a custom scope with a specific meter name
func WithCustomScope(scopeName string) fx.Option {
	return fx.Provide(
		func(params ScopeParams) ScopeResult {
			scope, err := revelio.New(scopeName)
			if err != nil {
				panic(err)
			}

			// Set as default scope for helper functions
			revelio.SetDefault(scope)

			return ScopeResult{
				Scope: scope,
			}
		},
	)
}

// WithCustomMeter provides a scope with a custom meter
func WithCustomMeter(customMeter metric.Meter) fx.Option {
	return fx.Provide(
		func(params ScopeParams) ScopeResult {
			scope := revelio.NewFromMeter(customMeter)

			// Set as default scope for helper functions
			revelio.SetDefault(scope)

			return ScopeResult{
				Scope: scope,
			}
		},
	)
}

// WithCustomMeterProvider provides a custom meter provider
func WithCustomMeterProvider(customMeterProvider metric.MeterProvider) fx.Option {
	return fx.Provide(
		func(params MeterProviderParams) MeterProviderResult {
			return MeterProviderResult{
				MeterProvider: customMeterProvider,
			}
		},
	)
}

// WithNamedMeter provides a scope with a named meter from the meter provider
func WithNamedMeter(meterName string, opts ...metric.MeterOption) fx.Option {
	return fx.Provide(
		func(meterProvider metric.MeterProvider) ScopeResult {
			meter := meterProvider.Meter(meterName, opts...)
			scope := revelio.NewFromMeter(meter)

			// Set as default scope for helper functions
			revelio.SetDefault(scope)

			return ScopeResult{
				Scope: scope,
			}
		},
	)
}

// AllProviders returns all default providers
var AllProviders = fx.Options(
	DefaultScopeProvider,
	MeterProviderProvider,
)
