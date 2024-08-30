package i18n

import (
	"context"

	"golang.org/x/text/language"
)

// FallbackLanguage is a global value for a default language. You can safely
// override this value in the beginning of your app.
var FallbackLanguage = language.Indonesian

type ctxKey struct{}

// WithContext wraps ctx with given lang.
func WithContext(ctx context.Context, lang language.Tag) context.Context {
	return context.WithValue(ctx, ctxKey{}, lang)
}

// FromContext returns language tag from given ctx, returns FallbackLanguage
// when empty.
func FromContext(ctx context.Context) language.Tag {
	ctxVal := ctx.Value(ctxKey{})
	if ctxVal == nil {
		return FallbackLanguage
	}
	lTag, _ := ctxVal.(language.Tag)
	return lTag
}
