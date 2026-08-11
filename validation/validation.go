// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: 	AGPL-3.0-or-later

package validation

import (
	"errors"
	"fmt"
	"mime"
	"net/url"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
)

var ErrInvalidField = errors.New("invalid field")
var ErrInvalidStruct = errors.New("invalid struct")

var validate *validator.Validate

func init() {
	validate = validator.New(validator.WithRequiredStructEnabled())
	if err := validate.RegisterValidation("mimetype_string", validateMimetype); err != nil {
		panic(err)
	}
	if err := validate.RegisterValidation("rfc3066lang", validateRFC3066Lang); err != nil {
		panic(err)
	}
	if err := validate.RegisterValidation("absolute_uri", validateAbsoluteURI); err != nil {
		panic(err)
	}
	if err := validate.RegisterValidation("not_url_encoded", validateNotURLEncoded); err != nil {
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

// Errors contains validation errors on individual fields in a struct.
type Errors struct {
	Fields []FieldError
}

// Error satisfies the Error interface.
func (e *Errors) Error() string {
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
func ValidateStruct(s any) *Errors {
	if err := validate.Struct(s); err != nil {
		errs := &Errors{}
		if validateErrs, ok := errors.AsType[validator.ValidationErrors](err); ok {
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

func ValidateField(value any, rule string) error {
	if err := validate.Var(value, rule); err != nil {
		errs := &Errors{}
		if validateErrs, ok := errors.AsType[validator.ValidationErrors](err); ok {
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

// validateMimetype checks that the value is a valid mimetype. Note that this does not check whether the mimetype is a
// registered IANA mimetype, only that it is *structurally* a mimetype.
func validateMimetype(fl validator.FieldLevel) bool {
	_, _, err := mime.ParseMediaType(fl.Field().String())
	return err == nil
}

// langTagRE is a pragmatic check for an [RFC3066] language tag: primary subtag, optionally followed by "-" subtags.
var langTagRE = regexp.MustCompile(`^[A-Za-z]{1,8}(-[A-Za-z0-9]{1,8})*$`)

// validateRFC3066Lang checks that the value is a valid RFC3066 language tag.
func validateRFC3066Lang(fl validator.FieldLevel) bool {
	if lang := fl.Field().String(); lang != "" && !langTagRE.MatchString(lang) {
		return false
	}
	return true
}

// validateAbsoluteURI checks that the URI value is an absolute value (i.e., has scheme/host).
func validateAbsoluteURI(fl validator.FieldLevel) bool {
	switch value, err := url.Parse(fl.Field().String()); {
	case err != nil:
		return false
	case !value.IsAbs():
		return false
	default:
		return true
	}
}

// percentEncodedPattern checks for a percent-encoded (i.e., URL encoded) parameter.
var percentEncodedPattern = regexp.MustCompile(`%[0-9A-Fa-f]{2}`)

// validateNotURLEncoded checks that the value is not URL encoded.
func validateNotURLEncoded(fl validator.FieldLevel) bool {
	return !percentEncodedPattern.MatchString(fl.Field().String())
}
