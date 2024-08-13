package zimemofx

import (
	"github.com/newrelic/go-agent/v3/newrelic"
	"gitlab.com/divikraf/lumos/db/zimemo"
	"go.uber.org/fx"
)

type memoSize uint

func SizeFn[T any](size func(T) uint) fx.Option {
	return fx.Provide(func(a T) memoSize {
		return memoSize(size(a))
	})
}

func Size(size uint) fx.Option {
	return SizeFn(func(fx.Lifecycle) uint {
		return size
	})
}

type memoParams struct {
	fx.In

	LC    fx.Lifecycle
	NrApp *newrelic.Application
	Size  memoSize `optional:"true"`
}

var Provider = fx.Provide(
	func(param memoParams) zimemo.ZiMemoization {
		if param.Size == 0 {
			param.Size = 256
		}
		sqlxMemo := zimemo.New(int(param.Size), param.NrApp)
		param.LC.Append(fx.StopHook(sqlxMemo.Purge))
		return sqlxMemo
	},
)
