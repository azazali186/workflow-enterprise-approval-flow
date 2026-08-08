package validation

import (
	"fmt"

	"github.com/cloudwego/hertz/pkg/app/server/binding"
	"github.com/cloudwego/hertz/pkg/protocol"
	"github.com/go-playground/validator/v10"
)

// bindingValidator is a go-playground/validator instance configured to read
// the Gin-style `binding` struct tags used across the API request schemas.
var bindingValidator = func() *validator.Validate {
	v := validator.New()
	v.SetTagName("binding")
	return v
}()

// NewBindingValidatorFunc returns a Hertz ValidatorFunc that validates request
// structs against their `binding` tags.
//
// Hertz v0.10 replaced its default validation engine (which honoured the
// `binding` tag) with a `vd`-tag expression engine, silently disabling every
// validation rule in this codebase. Registering this function via
// server.WithCustomValidatorFunc restores the intended behaviour.
func NewBindingValidatorFunc() binding.ValidatorFunc {
	return func(_ *protocol.Request, obj any) error {
		if err := bindingValidator.Struct(obj); err != nil {
			return fmt.Errorf("request validation failed: %w", err)
		}
		return nil
	}
}
