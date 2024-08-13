package zilog

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/diode"
	"github.com/rs/zerolog/pkgerrors"
)

var (
	DefaultDiode  diode.Writer
	DefaultLogger zLog
)

const (
	LevelDebug = "debug"
	LevelInfo  = "info"
	LevelError = "error"
	LevelWarn  = "warn"
)

func init() {
	zerolog.TimestampFieldName = "timestamp"
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelInfo,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			switch a.Key {
			case slog.TimeKey:
				return slog.Int64(zerolog.TimestampFieldName, a.Value.Time().UnixMilli())
			case slog.LevelKey:
				return slog.String(slog.LevelKey, strings.ToLower(a.Value.String()))
			}
			return a
		},
	})))
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnixMs
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
	DefaultDiode = diode.NewWriter(NewLevelWriter(os.Stdout), 1000, 1*time.Millisecond, func(missed int) {
		slog.Error(fmt.Sprintf("zLog: Dropped %d logs!!!\n", missed))
	})
	DefaultLogger = New(DefaultDiode, WithLoggerCallerSkipFrameCount(zerolog.CallerSkipFrameCount+2))
	zerolog.DefaultContextLogger = &DefaultLogger.Logger
	zerolog.ErrorHandler = func(err error) {
		slog.Error(err.Error())
	}
}

func NewLevelWriter(w io.Writer) *levelWriter {
	return &levelWriter{w}
}

type levelWriter struct {
	w io.Writer
}

func (lw *levelWriter) Write(p []byte) (n int, err error) {
	return lw.w.Write(p)
}

func (lw *levelWriter) WriteLevel(level zerolog.Level, p []byte) (n int, err error) {
	if level >= zerolog.GlobalLevel() {
		return lw.Write(p)
	}
	return 0, nil
}

// FromContext returns zerolog's Logger associated with the ctx.
// If no logger is associated or if the logger is disabled,
// then a DefaultLogger is returned.
func FromContext(ctx context.Context) *zerolog.Logger {
	return zerolog.Ctx(ctx)
}

// NewContext wraps a given ctx with logger and optional log hooks appended.
func NewContext(ctx context.Context, hooks ...zerolog.Hook) (context.Context, *zerolog.Logger) {
	logger := zerolog.Ctx(ctx)
	for _, h := range hooks {
		l := logger.Hook(h).With().Logger()
		logger = &l
	}
	return logger.WithContext(ctx), logger
}

// zLog wraps zerolog.Logger so we can implement go-kit's log.Logger into it.
type zLog struct {
	// Logger is the underlying zerolog.Logger instance
	zerolog.Logger
	Config LoggerConfig
}

// LoggerConfig configurable values for zlog.New
type LoggerConfig struct {
	CallerSkipFrameCount int
}

// LoggerOption config functional option for zlog.New
type LoggerOption func(cfg *LoggerConfig)

// WithLoggerCallerSkipFrameCount set CallerSkipFrameCount.
func WithLoggerCallerSkipFrameCount(skipCount int) LoggerOption {
	return func(cfg *LoggerConfig) {
		cfg.CallerSkipFrameCount = skipCount
	}
}

// New creates new zLog instance with opinionated defaults.
func New(output io.Writer, opts ...LoggerOption) zLog {
	config := LoggerConfig{
		CallerSkipFrameCount: zerolog.CallerSkipFrameCount + 1,
	}

	for _, o := range opts {
		o(&config)
	}

	logger := zerolog.
		New(output).
		With().
		Timestamp().
		CallerWithSkipFrameCount(config.CallerSkipFrameCount).
		Logger()

	return zLog{
		Logger: logger,
		Config: config,
	}
}
