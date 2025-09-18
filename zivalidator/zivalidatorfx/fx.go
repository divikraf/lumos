package zivalidatorfx

import (
	"github.com/divikraf/lumos/zivalidator"
	"go.uber.org/fx"
)

type validatorParams struct {
	fx.In

	Options []zivalidator.Option `group:"validator.option"`
}

// OptionFns can configure how validator behaves.
//
// DO NOT assume the sort order of the options. The array sort is not guaranteed
// as per this documentation:
// https://pkg.go.dev/go.uber.org/fx#hdr-Value_Groups.
//
// T is an injectable parameters/dependencies in Uber Fx framework. If you're
// unsure, use [go.uber.org/fx.Lifecycle] for T, or use the simpler [Options].
func OptionFns[T any](os ...func(T) zivalidator.Option) fx.Option {
	functors := []any{}
	for _, o := range os {
		functors = append(functors, fx.Annotate(o, fx.ResultTags(`group:"validator.option"`)))
	}
	return fx.Provide(functors...)
}

// Options is a syntactic sugar for [OptionFns].
//
// DO NOT assume the sort order of the options. The array sort is not guaranteed
// as per Fx documentation: https://pkg.go.dev/go.uber.org/fx#hdr-Value_Groups.
func Options(os ...zivalidator.Option) fx.Option {
	optionFns := []func(fx.Lifecycle) zivalidator.Option{}
	for _, o := range os {
		optionFns = append(optionFns, func(fx.Lifecycle) zivalidator.Option {
			return o
		})
	}
	return OptionFns(optionFns...)
}

type fxResult struct {
	fx.Out

	Validator *zivalidator.Validator
	Validate  zivalidator.Validate
}

// Provider provides *validator.Validator instance. To configure behavior, use
// [OptionFns].
var Provider = fx.Provide(
	func(params validatorParams) fxResult {
		v := zivalidator.New(params.Options...)
		return fxResult{
			Validator: v,
			Validate:  v,
		}
	},
)
