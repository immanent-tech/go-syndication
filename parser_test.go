// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: 	MIT

package feeds

import (
	"errors"
	"os"
	"slices"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/immanent-tech/go-syndication/validation"
)

// getFailedValidations extracts the validation errors from the given error object and converts them into a map of the
// struct field name and validation tags that failed. Test helpers use this format to check that specific validation
// rules have passed or failed.
func getFailedValidations(err error) (map[string][]string, error) {
	failedValidations := make(map[string][]string)
	if invalidValidationError, ok := errors.AsType[*validator.InvalidValidationError](err); ok {
		return nil, invalidValidationError
	}
	if validateErrs, ok := errors.AsType[*validation.Errors](err); ok && validateErrs != nil {
		for e := range slices.Values(validateErrs.Fields) {
			failedValidations[e.StructNamespace] = append(failedValidations[e.StructNamespace], e.Tag)
		}
	}
	return failedValidations, nil
}

func trunc(s string) string {
	if len(s) > 160 {
		return s[:160] + "..."
	}
	return s
}

func readFile(t *testing.T, path string) []byte {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return b
}
