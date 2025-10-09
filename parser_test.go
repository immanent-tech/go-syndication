// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: 	AGPL-3.0-or-later

package feeds

import (
	"errors"

	"github.com/go-playground/validator/v10"
)

// getFailedValidations extracts the validation errors from the given error object and converts them into a map of the
// struct field name and validation tags that failed. Test helpers use this format to check that specific validation
// rules have passed or failed.
func getFailedValidations(err error) (map[string][]string, error) {
	failedValidations := make(map[string][]string)
	var invalidValidationError *validator.InvalidValidationError
	if errors.As(err, &invalidValidationError) {
		return nil, invalidValidationError
	}
	var validateErrs validator.ValidationErrors
	if errors.As(err, &validateErrs) {
		for _, e := range validateErrs {
			failedValidations[e.StructNamespace()] = append(failedValidations[e.StructNamespace()], e.Tag())
		}
	}
	return failedValidations, nil
}
