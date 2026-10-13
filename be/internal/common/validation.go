package common

import (
	"github.com/go-playground/validator/v10"
)

var requestValidator = validator.New()

func Validate(value any) error {
	return requestValidator.Struct(value)
}
