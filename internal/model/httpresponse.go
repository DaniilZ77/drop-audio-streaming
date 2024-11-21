package model

import "github.com/MAXXXIMUS-tropical-milkshake/drop-audio-streaming/internal/model/validator"

type (
	errorResponse struct {
		Message string        `json:"message"`
		Details []interface{} `json:"details"`
	}

	validationErrors struct {
		Violations []violation `json:"violations"`
	}

	violation struct {
		Subject     string `json:"subject"`
		Description string `json:"description"`
	}
)

func ToErrorResponse(err error, details []interface{}) errorResponse {
	return errorResponse{
		Message: err.Error(),
		Details: details,
	}
}

func ToValidationErrors(v *validator.Validator) validationErrors {
	var errs validationErrors
	for k, v := range v.Errors {
		errs.Violations = append(errs.Violations, violation{
			Subject:     k,
			Description: v,
		})
	}
	return errs
}
