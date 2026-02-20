package validator

import (
	"github.com/go-playground/validator/v10"
)

// Validate is the shared validator instance.
var Validate *validator.Validate

func init() {
	Validate = validator.New()
}

// FormatValidationErrors extracts human-readable messages from validation errors.
func FormatValidationErrors(err error) []string {
	var errors []string
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, e := range validationErrors {
			switch e.Tag() {
			case "required":
				errors = append(errors, e.Field()+" is required")
			case "email":
				errors = append(errors, e.Field()+" must be a valid email address")
			case "min":
				errors = append(errors, e.Field()+" must be at least "+e.Param()+" characters")
			case "max":
				errors = append(errors, e.Field()+" must be at most "+e.Param()+" characters")
			case "gte":
				errors = append(errors, e.Field()+" must be greater than or equal to "+e.Param())
			case "lte":
				errors = append(errors, e.Field()+" must be less than or equal to "+e.Param())
			default:
				errors = append(errors, e.Field()+" failed validation: "+e.Tag())
			}
		}
	}
	return errors
}
