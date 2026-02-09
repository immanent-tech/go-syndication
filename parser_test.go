// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: 	MIT

package feeds

import (
	"errors"

	"github.com/go-playground/validator/v10"
	"github.com/immanent-tech/go-syndication/validation"
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
	var validateErrs validation.StructError
	if errors.Is(err, &validateErrs) {
		for _, e := range validateErrs.Fields {
			failedValidations[e.StructField] = append(failedValidations[e.StructNamespace], e.Tag)
		}
	}
	return failedValidations, nil
}
