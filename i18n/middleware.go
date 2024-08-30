package i18n

import (
	"github.com/gin-gonic/gin"
	"golang.org/x/text/language"
)

// LanguageMiddleware is a Gin middleware to inject the language tag into the context.
func LanguageMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Default language
		lang := FallbackLanguage

		// Check for language in the "Accept-Language" header
		if langHeader := c.GetHeader("Accept-Language"); langHeader != "" {
			parsedPrefLags, _, err := language.ParseAcceptLanguage(langHeader)
			if err != nil || parsedPrefLags == nil {
				lang = parsedPrefLags[0]
			}
		}

		nctx := WithContext(c.Request.Context(), lang)

		c.Request = c.Request.WithContext(nctx)

		// Continue to the next middleware/handler
		c.Next()
	}
}
