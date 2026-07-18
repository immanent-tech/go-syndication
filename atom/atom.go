// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: 	MIT

// Package atom contains objects and methods defining the Atom syndication format.
package atom

import (
	"encoding/xml"
	"fmt"
	"mime"
	"slices"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/immanent-tech/go-syndication/sanitization"
	"github.com/immanent-tech/go-syndication/validation"
)

const xhtmlNS = "http://www.w3.org/1999/xhtml"

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
		value.WriteString(" (")
		value.WriteString(p.Email.Value)
		value.WriteString(")")
	}
	if p.URI != nil && p.URI.Value != "" {
		value.WriteString(" ")
		value.WriteString(p.URI.Value)
	}
	return value.String()
}

// Validate ensures that the PersonConstruct is valid. If not, it returns a non-nil error containing details of any
// failed validation.
func (p *PersonConstruct) Validate() error {
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
	if c.UndefinedContent != nil && *c.UndefinedContent != "" {
		return sanitization.SanitizeString(*c.UndefinedContent)
	}
	// Use the term attribute.
	return sanitization.SanitizeString(c.Term.Value)
}

func (i *ID) String() string {
	return i.Value
}

func (l *Link) String() string {
	switch {
	case l.Href != "":
		return l.Href
	case l.UndefinedContent != nil && *l.UndefinedContent != "":
		return *l.UndefinedContent
	default:
		return ""
	}
}

func validateTypeAttr(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	if slices.Contains([]string{"text", "html", "xhtml"}, value) {
		return true
	}
	_, _, err := mime.ParseMediaType(value)
	return err == nil
}

func (t TextConstruct) String() string {
	if t.Type == nil || *t.Type != TypeXhtml {
		return t.Value
	}
	return *t.XHTML
}

// MarshalXML implements xml.Marshaler. The element name itself (title,
// summary, subtitle, rights, ...) comes from `start`, as set by the
// enclosing struct's field tag -- e.g. `Title TextConstruct \`xml:"title"\“.
func (t TextConstruct) MarshalXML(enc *xml.Encoder, start xml.StartElement) error {
	var typ Type
	if t.Type == nil {
		typ = TypeText
	} else {
		typ = *t.Type
	}
	if typ != TypeText && typ != TypeHtml && typ != TypeXhtml {
		return fmt.Errorf("atom text construct: invalid type %q (must be text, html, or xhtml)", typ)
	}

	start.Attr = nil // don't inherit anything unexpected from the caller
	start.Attr = append(start.Attr, xml.Attr{Name: xml.Name{Local: "type"}, Value: string(typ)})
	if t.Lang != nil {
		start.Attr = append(start.Attr, xml.Attr{Name: xml.Name{Space: "xml", Local: "lang"}, Value: *t.Lang})
	}
	if t.Base != nil {
		start.Attr = append(start.Attr, xml.Attr{Name: xml.Name{Space: "xml", Local: "base"}, Value: *t.Base})
	}

	if err := enc.EncodeToken(start); err != nil {
		return fmt.Errorf("text construct: encode start element: %w", err)
	}

	if typ == TypeXhtml {
		div := struct {
			XMLName xml.Name `xml:"div"`
			XMLNS   string   `xml:"xmlns,attr"`
			Inner   string   `xml:",innerxml"`
		}{
			XMLName: xml.Name{Local: "div"},
			XMLNS:   xhtmlNS,
			Inner:   *t.XHTML,
		}
		if err := enc.Encode(div); err != nil {
			return fmt.Errorf("text construct: encode %s: %w", typ, err)
		}
	} else {
		// EncodeToken auto-escapes special characters, which is exactly
		// what "html" type requires ("<br>" -> "&lt;br>") and is harmless
		// (a no-op beyond normal XML escaping) for "text" type.
		if err := enc.EncodeToken(xml.CharData(t.Value)); err != nil {
			return fmt.Errorf("text construct: encode %s: %w", typ, err)
		}
	}

	if err := enc.EncodeToken(start.End()); err != nil {
		return fmt.Errorf("text construct: encode end element: %w", err)
	}

	return nil
}

// UnmarshalXML implements xml.Unmarshaler.
func (t *TextConstruct) UnmarshalXML(dec *xml.Decoder, start xml.StartElement) error {
	typ := TypeText // spec: absent type attribute defaults to "text"
	for attr := range slices.Values(start.Attr) {
		switch {
		case attr.Name.Local == "type" && attr.Name.Space == "":
			typ = Type(attr.Value)
		case attr.Name.Local == "lang" && attr.Name.Space == "xml":
			t.Lang = new(attr.Value)
		case attr.Name.Local == "base" && attr.Name.Space == "xml":
			t.Base = new(attr.Value)
		}
	}
	t.Type = new(typ)

	if typ == "xhtml" {
		var wrapper struct {
			Div struct {
				Inner string `xml:",innerxml"`
			} `xml:"div"` // matches any namespace's local-name "div"
		}
		if err := dec.DecodeElement(&wrapper, &start); err != nil {
			return fmt.Errorf("text construct: marshal wrapper: %w", err)
		}
		t.XHTML = new(strings.TrimSpace(wrapper.Div.Inner))
		return nil
	}

	// "text" and "html" (and, leniently, anything else): plain character
	// data. The decoder already unescapes entities for us, so for "html"
	// content this correctly yields real markup back as a Go string.
	var valueStruct struct {
		Value string `xml:",chardata"`
	}
	if err := dec.DecodeElement(&valueStruct, &start); err != nil {
		return fmt.Errorf("text construct: marshal value: %w", err)
	}
	t.Value = valueStruct.Value
	return nil
}
