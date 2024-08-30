package zivalidator

import (
	"context"

	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
)

type ValidationResult struct {
	Message     string `json:"message"`
	FieldErrors []FieldError
}

type FieldError struct {
	Key string `json:"field"`
	Msg string `json:"message"`
}

type FieldErrors []FieldError

func NewFieldErrors(translator ut.Translator, validationErrors error) (FieldErrors, string) {
	out := FieldErrors{}
	if validationErrors == nil {
		return out, ""
	}

	invalidValidationErr, isInvalidValidationErr := validationErrors.(*validator.InvalidValidationError)
	if isInvalidValidationErr {
		return append(out, FieldError{
			Key: "struct",
			Msg: invalidValidationErr.Error(),
		}), invalidValidationErr.Error()
	}

	errs, isErrs := validationErrors.(validator.ValidationErrors)
	if !isErrs {
		return out, ""
	}

	for _, e := range errs {
		out = append(out, FieldError{
			Key: e.Field(),
			Msg: e.Translate(translator),
		})
	}

	return out, "processable entity"
}

type Validate interface {
	ValidateStruct(ctx context.Context, s any) *ValidationResult
}
