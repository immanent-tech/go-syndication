// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: 	MIT

// Package atom contains objects and methods defining the Atom syndication format.
package atom

import (
	"errors"
	"fmt"
	"mime"
	"slices"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/immanent-tech/go-syndication/sanitization"
	"github.com/immanent-tech/go-syndication/validation"
)

var ErrPersonConstruct = errors.New("person construct is invalid")

func init() {
	if err := validation.RegisterValidation("type_attr", validateTypeAttr); err != nil {
		panic(err)
	}
}

// String returns string-ified format of the PersonConstruct. This will be the format "name (email)". The email part is
// omitted if the PersonConstruct has no email.
func (p *PersonConstruct) String() string {
	var value strings.Builder
	value.WriteString(p.Name.Value)
	if p.Email != nil && p.Email.Value != "" {
		value.WriteString(" (" + p.Email.Value + ")")
	}
	if p.URI != nil && p.URI.Value != "" {
		value.WriteString(" " + p.URI.Value)
	}
	return value.String()
}

// Validate ensures that the PersonConstruct is valid. If not, it returns a non-nil error containing details of any
// failed validation.
func (p *PersonConstruct) Validate() error {
	// htmlEncodedErr := validation.Validate(p.Name.Value, "html_encoded")
	// var validateErrs validator.ValidationErrors
	// if errors.As(htmlEncodedErr, &validateErrs) {
	// 	slog.Info("invalid name")
	// 	return fmt.Errorf("%w: name cannot be HTML encoded", ErrPersonConstruct)
	// }
	// returns nil or ValidationErrors ( []FieldError
	if err := validation.ValidateStruct(p); err != nil {
		return fmt.Errorf("person construct is not valid: %w", err)
	}
	return nil
}

// String returns the string-ified format of the Category. It will return the first found of: any human-readable label,
// the element value or the term attribute value, in that order.
func (c *Category) String() string {
	// Use the label attribute if present.
	if c.Label != nil && c.Label.Value != "" {
		return sanitization.SanitizeString(c.Label.Value)
	}
	// Use any value if present.
	if c.UndefinedContent != nil {
		return sanitization.SanitizeString(*c.UndefinedContent)
	}
	// Use the term attribute.
	return sanitization.SanitizeString(c.Term.Value)
}

func (t *Title) String() string {
	return sanitization.SanitizeString(t.Value)
}

func (t *Subtitle) String() string {
	return sanitization.SanitizeString(t.Value)
}

func (s *Summary) String() string {
	return sanitization.SanitizeString(s.Value)
}

func (i *ID) String() string {
	return i.Value
}

func (l *Link) String() string {
	if l.Href == "" && l.UndefinedContent != nil {
		return *l.UndefinedContent
	}
	return l.Href
}

func validateTypeAttr(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	if slices.Contains([]string{"text", "html", "xhtml"}, value) {
		return true
	}
	_, _, err := mime.ParseMediaType(value)
	return err == nil
}
