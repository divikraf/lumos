package hook

import (
	"github.com/newrelic/go-agent/v3/integrations/logcontext"
	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/rs/zerolog"
)

// NewRelicRecorderHook is a hook that calls to NewRelic log metric so that APM page
// displays more informational data about log calls.
func NewRelicRecorderHook(txn *newrelic.Transaction) zerolog.Hook {
	return zerolog.HookFunc(
		func(e *zerolog.Event, level zerolog.Level, message string) {
			lmd := txn.GetLinkingMetadata()
			e.Str(logcontext.KeyEntityName, lmd.EntityName)
			e.Str(logcontext.KeyEntityGUID, lmd.EntityGUID)
			e.Str(logcontext.KeyEntityType, lmd.EntityType)
			e.Str(logcontext.KeyHostname, lmd.Hostname)

			data := newrelic.LogData{
				Severity: level.String(),
				Message:  message,
			}

			txn.RecordLog(data)
		},
	)
}
