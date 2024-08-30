package zivalidator

import (
	"context"

	"github.com/go-playground/locales/en"
	"github.com/go-playground/locales/id"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	en_translations "github.com/go-playground/validator/v10/translations/en"
	id_translations "github.com/go-playground/validator/v10/translations/id"
	"gitlab.com/divikraf/lumos/i18n"
	"golang.org/x/text/language"
)

type Validator struct {
	uni      *ut.UniversalTranslator
	validate *validator.Validate
}

var _ Validate = (*Validator)(nil)

// Option is a configuration function for Validator.
type Option func(ut *ut.UniversalTranslator, v *validator.Validate) error

// New creates validator object with Indonesian and English default translations
// loaded. This function is not safe to run concurrently. It is best that you
// only create 1 instance of Validator for your app and reuse it many times.
func New(opts ...Option) *Validator {
	localeEN := en.New()
	localeID := id.New()
	uni := ut.New(localeEN, localeID)

	translatorEN, _ := uni.GetTranslator(localeEN.Locale())
	translatorID, _ := uni.GetTranslator(localeID.Locale())

	validate := validator.New()

	// register default validation translations for Indonesian.
	if err := id_translations.RegisterDefaultTranslations(validate, translatorID); err != nil {
		panic(err)
	}

	// register default validation translations for English.
	if err := en_translations.RegisterDefaultTranslations(validate, translatorEN); err != nil {
		panic(err)
	}

	for _, o := range opts {
		errOpt := o(uni, validate)
		if errOpt != nil {
			panic(errOpt)
		}
	}

	return &Validator{
		uni:      uni,
		validate: validate,
	}
}

// ValidateStruct will do a struct validation given ctx and arbitrary struct.
// This function will automatically determine which language should the
// validation string should be outputted from given ctx. Language defaults to
// "id" when not found in the ctx.
func (v *Validator) ValidateStruct(ctx context.Context, s any) *ValidationResult {
	err := v.validate.StructCtx(ctx, s)
	if err == nil {
		return nil
	}

	langStr := "id"
	if i18n.FromContext(ctx) != language.Indonesian {
		langStr = "en"
	}

	out := &ValidationResult{}
	theTranslator, _ := v.uni.GetTranslator(langStr)
	out.FieldErrors, out.Message = NewFieldErrors(theTranslator, err)

	return out
}
