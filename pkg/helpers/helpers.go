package helpers

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

func TranslateErrors(errors validator.ValidationErrors) []string {
	var validationErrors []string
	for _, err := range errors {
		if err.Tag() == "required" {
			var stringErr = fmt.Sprintf("%s is required", err.Field())
			validationErrors = append(validationErrors, stringErr)
		}
	}
	return validationErrors
}
