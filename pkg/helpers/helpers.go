package helpers

import (
	"context"
	"fmt"
	"time"

	"github.com/dombrga/go-crud-person/pkg/models"
	"github.com/go-playground/validator/v10"
)

func ValidateBody(person models.PersonRequest) error {
	var validate = validator.New()
	var structErr = validate.Struct(person)
	return structErr
}

func TranslateErrors(errors validator.ValidationErrors) []string {
	var validationErrors []string
	for _, err := range errors {
		if err.Tag() == "required" {
			// "<field> is a required."
			var stringErr = fmt.Sprintf("%s is required", err.Field())
			validationErrors = append(validationErrors, stringErr)
		}
	}
	return validationErrors
}

func CreateContext() (context.Context, context.CancelFunc) {
	var ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	return ctx, cancel
}
