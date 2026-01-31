// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: 	AGPL-3.0-or-later

package validation

import (
	"mime"

	"github.com/go-playground/validator/v10"
)

var Validate *validator.Validate

func init() {
	Validate = validator.New()
	if err := Validate.RegisterValidation("mimetype", validateMimetype); err != nil {
		panic(err)
	}
}

// validateMimetype checks that the field is a valid mimetype.
func validateMimetype(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	_, _, err := mime.ParseMediaType(value)
	return err == nil
}
