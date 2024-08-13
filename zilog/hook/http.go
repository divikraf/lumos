package hook

import (
	"github.com/rs/zerolog"
)

// NewHTTPPath appends http.path into log.
func NewHTTPPath(p string) zerolog.Hook {
	return zerolog.HookFunc(func(e *zerolog.Event, level zerolog.Level, message string) {
		e.Str("http.path", p)
	})
}
