package api

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type Validator struct{}

func (v *Validator) Validate(i any) error {
	if c, ok := i.(validation.Validatable); ok {
		return c.Validate()
	}
	return nil
}
