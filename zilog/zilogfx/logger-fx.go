package zilogfx

import (
	"context"
	"log/slog"
	"strings"
	"time"

	"go.uber.org/fx/fxevent"
)

// SlogLogger is an Fx event logger that logs events to log/slog.
type SlogLogger struct {
	Logger *slog.Logger
}

var _ fxevent.Logger = (*SlogLogger)(nil)

func newLogRecord(level slog.Level, msg string, fields []any) slog.Record {
	// The `pc` var was intended to be zero-ed, due to the fact that `source`
	// information cannot be determined in Uber FX callstack. Fret not, there
	// are still some fields that clients can refer when debugging that prefixed
	// with `fx.` namespace.
	r := slog.NewRecord(time.Now(), level, msg, 0)
	r.Add(fields...)
	return r
}

func (l *SlogLogger) writeInfo(msg string, fields ...any) {
	record := newLogRecord(slog.LevelInfo, msg, fields)
	_ = l.Logger.Handler().Handle(context.TODO(), record)
}

func (l *SlogLogger) writeError(msg string, fields ...any) {
	record := newLogRecord(slog.LevelError, msg, fields)
	_ = l.Logger.Handler().Handle(context.TODO(), record)
}

const (
	failedApplyOption = "error encountered while applying options"
)

func (l *SlogLogger) LogEvent(event fxevent.Event) { //nolint:funlen,gocognit // this is expected to have many switch cases.
	switch e := event.(type) {
	case *fxevent.OnStartExecuting:
		l.writeInfo("OnStart hook executing",
			slog.Group("fx",
				slog.String("callee", e.FunctionName),
				slog.String("caller", e.CallerName),
			))
	case *fxevent.OnStartExecuted:
		if e.Err == nil {
			l.writeInfo("OnStart hook executed",
				slog.Group("fx",
					slog.String("callee", e.FunctionName),
					slog.String("caller", e.CallerName),
					slog.Duration("runtime", e.Runtime),
				),
			)
			break
		}
		l.writeError("OnStart hook failed",
			slog.String("err", e.Err.Error()),
			slog.Group("fx",
				slog.String("callee", e.FunctionName),
				slog.String("caller", e.CallerName),
			),
		)
	case *fxevent.OnStopExecuting:
		l.writeInfo("OnStop hook executing",
			slog.Group("fx",
				slog.String("callee", e.FunctionName),
				slog.String("caller", e.CallerName),
			),
		)
	case *fxevent.OnStopExecuted:
		if e.Err == nil {
			l.writeInfo("OnStop hook executed",
				slog.Group("fx",
					slog.String("callee", e.FunctionName),
					slog.String("caller", e.CallerName),
					slog.String("runtime", e.Runtime.String()),
				),
			)
			break
		}
		l.writeError("OnStop hook failed",
			slog.String("err", e.Err.Error()),
			slog.Group("fx",
				slog.String("callee", e.FunctionName),
				slog.String("caller", e.CallerName),
			),
		)
	case *fxevent.Supplied:
		if e.Err == nil {
			l.writeInfo("supplied",
				slog.Group("fx",
					slog.String("module", e.ModuleName),
					slog.String("typename", e.TypeName),
					slog.Any("stacktrace", e.StackTrace),
				),
			)
			break
		}
		l.writeError(failedApplyOption,
			slog.String("err", e.Err.Error()),
			slog.Group("fx",
				slog.String("module", e.ModuleName),
				slog.String("typename", e.TypeName),
				slog.Any("stacktrace", e.StackTrace),
			),
		)
	case *fxevent.Provided:
		for _, rtype := range e.OutputTypeNames {
			l.writeInfo("provided",
				slog.Group("fx",
					slog.String("module", e.ModuleName),
					slog.Bool("private", e.Private),
					slog.String("typename", rtype),
					slog.String("constructor", e.ConstructorName),
					slog.Any("stacktrace", e.StackTrace),
				),
			)
		}
		if e.Err == nil {
			break
		}
		l.writeError(failedApplyOption,
			slog.String("err", e.Err.Error()),
			slog.Group("fx",
				slog.String("module", e.ModuleName),
				slog.Bool("private", e.Private),
				slog.Any("stacktrace", e.StackTrace),
			),
		)
	case *fxevent.Replaced:
		for _, rtype := range e.OutputTypeNames {
			l.writeInfo("replaced",
				slog.Group("fx",
					slog.String("module", e.ModuleName),
					slog.String("typename", rtype),
					slog.Any("stacktrace", e.StackTrace),
				),
			)
		}
		if e.Err == nil {
			break
		}
		l.writeError("error encountered while replacing",
			slog.String("err", e.Err.Error()),
			slog.Group("fx",
				slog.String("module", e.ModuleName),
				slog.Any("stacktrace", e.StackTrace),
			),
		)
	case *fxevent.Decorated:
		for _, rtype := range e.OutputTypeNames {
			l.writeInfo("decorated",
				slog.Group("fx",
					slog.String("module", e.ModuleName),
					slog.String("typename", rtype),
					slog.String("decorator", e.DecoratorName),
					slog.Any("stacktrace", e.StackTrace),
				),
			)
		}
		if e.Err == nil {
			break
		}
		l.writeError(failedApplyOption,
			slog.String("err", e.Err.Error()),
			slog.Group("fx",
				slog.String("module", e.ModuleName),
				slog.Any("stacktrace", e.StackTrace),
			),
		)
	case *fxevent.Run:
		if e.Err == nil {
			l.writeInfo("run",
				slog.Group("fx",
					slog.String("module", e.ModuleName),
					slog.String("name", e.Name),
					slog.String("kind", e.Kind),
				),
			)
			break
		}
		l.writeInfo("error returned",
			slog.String("err", e.Err.Error()),
			slog.Group("fx",
				slog.String("module", e.ModuleName),
				slog.String("name", e.Name),
				slog.String("kind", e.Kind),
			),
		)
	case *fxevent.Invoking:
		// Do not log stack as it will make logs hard to read.
		l.writeInfo("invoking",
			slog.Group("fx",
				slog.String("module", e.ModuleName),
				slog.String("function", e.FunctionName),
			),
		)
	case *fxevent.Invoked:
		if e.Err == nil {
			break
		}
		l.writeError("invoke failed",
			slog.String("err", e.Err.Error()),
			slog.Group("fx",
				slog.String("module", e.ModuleName),
				slog.String("stack", e.Trace),
				slog.String("function", e.FunctionName),
			),
		)
	case *fxevent.Stopping:
		l.writeInfo("received signal", slog.String("signal", strings.ToUpper(e.Signal.String())))
	case *fxevent.Stopped:
		if e.Err != nil {
			l.writeError("stop failed", slog.String("err", e.Err.Error()))
		}
	case *fxevent.RollingBack:
		l.writeError("start failed, rolling back", slog.String("err", e.StartErr.Error()))
	case *fxevent.RolledBack:
		if e.Err != nil {
			l.writeError("rollback failed", slog.String("err", e.Err.Error()))
		}
	case *fxevent.Started:
		if e.Err == nil {
			l.writeInfo("started")
			break
		}
		l.writeError("start failed", slog.String("err", e.Err.Error()))
	case *fxevent.LoggerInitialized:
		if e.Err != nil {
			l.writeError("custom logger initialization failed", slog.String("err", e.Err.Error()))
			break
		}
		l.writeInfo("initialized custom fxevent.Logger", slog.Group("fx", slog.String("function", e.ConstructorName)))
	}
}
