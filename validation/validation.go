// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: 	AGPL-3.0-or-later

package validation

import (
	"errors"
	"fmt"
	"mime"
	"strings"

	"github.com/go-playground/validator/v10"
)

var ErrInvalidField = errors.New("invalid field")
var ErrInvalidStruct = errors.New("invalid struct")

var validate *validator.Validate

func init() {
	validate = validator.New()
	if err := validate.RegisterValidation("mimetype", validateMimetype); err != nil {
		panic(err)
	}
}

// FieldError is a particular validation error on a particular field.
type FieldError struct {
	Namespace       string `json:"namespace"` // can differ when a custom TagNameFunc is registered or
	Field           string `json:"field"`     // by passing alt name to ReportError like below
	StructNamespace string `json:"structNamespace"`
	StructField     string `json:"structField"`
	Tag             string `json:"tag"`
	ActualTag       string `json:"actualTag"`
	Kind            string `json:"kind"`
	Type            string `json:"type"`
	Value           string `json:"value"`
	Param           string `json:"param"`
	Message         string `json:"message"`
}

// Error satisfies the Error interface.
func (e *FieldError) Error() string {
	return fmt.Sprintf(
		"%s %s (value(%q)) failed validation for %s: %s",
		ErrInvalidField.Error(),
		e.Field,
		e.Value,
		e.Tag,
		e.Message,
	)
}

// StructError contains validation errors on individual fields in a struct.
type StructError struct {
	Fields []FieldError
}

// Error satisfies the Error interface.
func (e *StructError) Error() string {
	var errStr strings.Builder
	errStr.WriteString("contains field errors")
	if len(e.Fields) > 0 {
		errStr.WriteRune('\n')
	}
	for idx, t := range e.Fields {
		errStr.WriteString(t.Error())
		if idx < (len(e.Fields) - 1) {
			errStr.WriteRune('\n')
		}
	}
	return errStr.String()
}

// ValidateStruct performs validation on the given struct. If validation fails, a non-nil error is returned that
// contains the details of individual field validation issues.
func ValidateStruct(s any) *StructError {
	if err := validate.Struct(s); err != nil {
		errs := &StructError{}
		var validateErrs validator.ValidationErrors
		if errors.As(err, &validateErrs) {
			errs.Fields = make([]FieldError, 0, len(validateErrs))
			for _, err := range validateErrs {
				errs.Fields = append(errs.Fields, FieldError{
					Namespace:       err.Namespace(),
					Field:           err.Field(),
					StructNamespace: err.StructNamespace(),
					StructField:     err.StructField(),
					Tag:             err.Tag(),
					ActualTag:       err.ActualTag(),
					Kind:            fmt.Sprintf("%v", err.Kind()),
					Type:            fmt.Sprintf("%v", err.Type()),
					Value:           fmt.Sprintf("%v", err.Value()),
					Param:           err.Param(),
					Message:         err.Error(),
				})
			}
			return errs
		}
	}
	return nil
}

// RegisterValidation will register a new validation tag, using the given function, on the global validator.
func RegisterValidation(tag string, f validator.Func) error {
	if err := validate.RegisterValidation(tag, f); err != nil {
		return fmt.Errorf("unable to register custom validator: %w", err)
	}
	return nil
}

// validateMimetype checks that the field is a valid mimetype.
func validateMimetype(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	_, _, err := mime.ParseMediaType(value)
	return err == nil
}
