// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: 	MIT

// Package atom contains objects and methods defining the Atom syndication format.
package atom

import (
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"mime"
	"slices"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/immanent-tech/go-syndication/sanitization"
	"github.com/immanent-tech/go-syndication/validation"
)

const atomNS = "http://www.w3.org/1999/xhtml"

// dateLayout mirrors time.RFC3339Nano: "2006-01-02T15:04:05.999999999Z07:00". The trailing ".999999999" is Go's
// convention for "trim trailing zero fractional digits, omit entirely if zero". This naturally produces the spec's
// *optional* fractional-seconds behavior. The literal "T" and the "Z07:00" zone verb naturally produce uppercase "T"
// and "Z" on output, exactly matching the spec's requirement.
const dateLayout = time.RFC3339Nano

func init() {
	if err := validation.RegisterValidation("type_attr", validateTypeAttr); err != nil {
		panic(err)
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

// String returns string-ified format of the PersonConstruct. This will be the format "name (email)". The email part is
// omitted if the PersonConstruct has no email.
func (p PersonConstruct) String() string {
	var value strings.Builder
	value.WriteString(p.Name)
	if p.Email != nil && *p.Email != "" {
		value.WriteString(" (")
		value.WriteString(*p.Email)
		value.WriteString(")")
	}
	if p.URI != nil && *p.URI != "" {
		value.WriteString(" ")
		value.WriteString(*p.URI)
	}
	return value.String()
}

// String returns the string-ified format of the Category. It will return the first found of: any human-readable label,
// the element value or the term attribute value, in that order.
func (c Category) String() string {
	// Use the term attribute.
	if c.Term.Value != "" {
		return sanitization.SanitizeString(c.Term.Value)
	}
	// Use the label attribute if present.
	if c.Label != nil && c.Label.Value != "" {
		return sanitization.SanitizeString(c.Label.Value)
	}
	// Use any value if present.
	if len(c.Extensions) > 0 {
		content := make([]string, 0, len(c.Extensions))
		for value := range slices.Values(c.Extensions) {
			if value.Content != "" {
				content = append(content, value.Content)
			}
		}
		return sanitization.SanitizeString(strings.Join(content, " "))
	}
	return ""
}

// String formats the generator value as a string in the format VALUE[/VERSION] [(URI)].
func (g Generator) String() string {
	var gen strings.Builder
	gen.WriteString(g.Value)
	if g.Version != nil && *g.Version != "" {
		gen.WriteString("/")
		gen.WriteString(*g.Version)
	}
	if g.URI != nil && *g.URI != "" {
		gen.WriteString("(")
		gen.WriteString(*g.URI)
		gen.WriteString(")")
	}

	return gen.String()
}

func (i Icon) String() string {
	return i.Value
}

func (i ID) String() string {
	return i.Value
}

func (l Logo) String() string {
	return l.Value
}

func (l Link) String() string {
	switch {
	case l.Href != "":
		return l.Href
	case l.UndefinedContent != nil && *l.UndefinedContent != "":
		return sanitization.SanitizeString(*l.UndefinedContent)
	default:
		return ""
	}
}

func (l *Link) Validate() error {
	if l.Rel == LinkRelEnclosure && l.Length != nil {
		// SHOULD, not MUST -- not a hard error, but worth flagging.
		return fmt.Errorf("atom:link: rel=%q SHOULD include a length attribute", LinkRelEnclosure)
	}
	if err := validation.ValidateStruct(l); err != nil {
		return fmt.Errorf("validate atom:link: %w", err)
	}
	return nil
}

func (t TextConstruct) String() string {
	switch {
	case t.Type == nil || *t.Type != TypeXhtml:
		return strings.TrimSpace(t.Value)
	case t.XHTML != nil:
		return *t.XHTML
	default:
		return ""
	}
}

// MarshalXML implements xml.Marshaler. The element name itself (title, summary, subtitle, rights, ...) comes from
// `start`, as set by the enclosing struct's field tag -- e.g. `Title TextConstruct \`xml:"title"\“.
func (t TextConstruct) MarshalXML(enc *xml.Encoder, start xml.StartElement) error {
	var typ Type
	if t.Type == nil {
		typ = TypeText
	} else {
		typ = *t.Type
	}
	if typ != TypeText && typ != TypeHtml && typ != TypeXhtml {
		return fmt.Errorf("text construct: invalid type %q (must be text, html, or xhtml)", typ)
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
		return fmt.Errorf("text construct: marshal: %w", err)
	}

	if typ == TypeXhtml {
		div := struct {
			XMLName xml.Name `xml:"div"`
			XMLNS   string   `xml:"xmlns,attr"`
			Inner   string   `xml:",innerxml"`
		}{
			XMLName: xml.Name{Local: "div"},
			XMLNS:   atomNS,
			Inner:   *t.XHTML,
		}
		if err := enc.Encode(div); err != nil {
			return fmt.Errorf("text construct: marshal %s: %w", typ, err)
		}
	} else {
		// EncodeToken auto-escapes special characters, which is exactly what "html" type requires ("<br>" -> "&lt;br>")
		// and is harmless (a no-op beyond normal XML escaping) for "text" type.
		if err := enc.EncodeToken(xml.CharData(t.Value)); err != nil {
			return fmt.Errorf("text construct: marshal: %w", err)
		}
	}

	if err := enc.EncodeToken(start.End()); err != nil {
		return fmt.Errorf("text construct: marshal: %w", err)
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
			return fmt.Errorf("text construct: unmarshal: %w", err)
		}
		t.XHTML = new(strings.TrimSpace(wrapper.Div.Inner))
		return nil
	}

	// "text" and "html" (and, leniently, anything else): plain character data. The decoder already unescapes entities
	// for us, so for "html" content this correctly yields real markup back as a Go string.
	var valueStruct struct {
		Value string `xml:",chardata"`
	}
	if err := dec.DecodeElement(&valueStruct, &start); err != nil {
		return fmt.Errorf("text construct: unmarshal: %w", err)
	}
	t.Value = valueStruct.Value
	return nil
}

func (d DateConstruct) String() string {
	return d.Value.Format(time.RFC3339)
}

// MarshalXML implements xml.Marshaler.
func (d DateConstruct) MarshalXML(enc *xml.Encoder, start xml.StartElement) error {
	if d.Value.IsZero() {
		return fmt.Errorf("date construct: zero time.Time value for <%s>", start.Name.Local)
	}
	start.Attr = nil
	if d.Base != nil {
		start.Attr = append(start.Attr, xml.Attr{Name: xml.Name{Space: "xml", Local: "base"}, Value: *d.Base})
	}
	if d.Lang != nil {
		start.Attr = append(start.Attr, xml.Attr{Name: xml.Name{Space: "xml", Local: "lang"}, Value: *d.Lang})
	}
	if err := enc.EncodeToken(start); err != nil {
		return fmt.Errorf("date construct: marshal: %w", err)
	}
	if err := enc.EncodeToken(xml.CharData(d.Value.Format(dateLayout))); err != nil {
		return fmt.Errorf("date construct: marshal: %w", err)
	}

	if err := enc.EncodeToken(start.End()); err != nil {
		return fmt.Errorf("date construct: marshal: %w", err)
	}

	return nil
}

// UnmarshalXML implements xml.Unmarshaler.
//
// We parse leniently (Go's time.Parse against the RFC3339 layout already accepts an optional fractional-seconds
// component even though the layout itself doesn't spell one out (a documented quirk of time.Parse) and also happens to
// accept lowercase "t"/"z", which strictly isn't legal Atom). If you need to reject non-conformant producers rather
// than accept them liberally, call Validate on the raw text before parsing, or check d.Time.Format(dateLayout) against
// the original string after decoding.
func (d *DateConstruct) UnmarshalXML(dec *xml.Decoder, start xml.StartElement) error {
	for _, a := range start.Attr {
		switch {
		case a.Name.Local == "base" && a.Name.Space == "xml":
			d.Base = &a.Value
		case a.Name.Local == "lang" && a.Name.Space == "xml":
			d.Lang = &a.Value
		}
	}
	var valueStruct struct {
		Value string `xml:",chardata"`
	}
	if err := dec.DecodeElement(&valueStruct, &start); err != nil {
		return fmt.Errorf("date construct: unmarshal: %w", err)
	}
	t, err := time.Parse(time.RFC3339, valueStruct.Value)
	if err != nil {
		return fmt.Errorf("date construct: invalid date-time %q: %w", valueStruct.Value, err)
	}
	d.Value = t
	return nil
}

// Validate rejects date-time strings that parse fine under RFC 3339 in general but violate RFC 4287's stricter
// uppercase-T/Z requirement.
func (d *DateConstruct) Validate() error {
	raw := d.String()
	if _, err := time.Parse(time.RFC3339, raw); err != nil {
		return fmt.Errorf("date construct: invalid date-time %q: %w", raw, err)
	}
	// time.Parse accepts lowercase t/z against this layout too; the spec doesn't, so check the literal separator
	// characters ourselves.
	tIdx := 10 // "2006-01-02" is always 10 bytes before the separator
	if tIdx >= len(raw) || raw[tIdx] != 'T' {
		return fmt.Errorf("date construct: %q must use an uppercase %q separator", raw, "T")
	}
	if raw[len(raw)-1] == 'z' {
		return fmt.Errorf("date construct: %q must use an uppercase %q zone indicator", raw, "Z")
	}
	return nil
}

func isXMLMediaType(t Type) bool {
	return t != TypeText && t != TypeHtml && t != TypeXhtml &&
		!strings.HasPrefix(string(t), "text/") &&
		(strings.HasSuffix(string(t), "+xml") || strings.HasSuffix(string(t), "/xml"))
}

func (c Content) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	var typ Type
	if c.Type == nil {
		typ = TypeText
	} else {
		typ = *c.Type
	}
	start.Attr = nil
	start.Attr = append(start.Attr, xml.Attr{Name: xml.Name{Local: "type"}, Value: string(typ)})
	if c.Source != nil {
		start.Attr = append(start.Attr, xml.Attr{Name: xml.Name{Local: "src"}, Value: *c.Source})
	}
	if c.Base != nil {
		start.Attr = append(start.Attr, xml.Attr{Name: xml.Name{Space: "xml", Local: "base"}, Value: *c.Base})
	}
	if c.Lang != nil {
		start.Attr = append(start.Attr, xml.Attr{Name: xml.Name{Space: "xml", Local: "lang"}, Value: *c.Lang})
	}

	// The embedded-XML branch needs start+raw-content+end written in one shot via the ",innerxml" struct tag (which, in
	// addition to being how Unmarshal captures raw XML, is also how Marshal writes a string out UNESCAPED rather than
	// as normal character data, as opposed to text.CharData's automatic escaping). xml.Encoder has no public raw-write
	// method, so this indirection through a throwaway struct is the idiomatic way to get unescaped output.
	if c.Source == nil && isXMLMediaType(typ) {
		wrapper := struct {
			Inner string `xml:",innerxml"`
		}{Inner: *c.XML}
		return e.EncodeElement(wrapper, start)
	}

	if err := e.EncodeToken(start); err != nil {
		return err
	}
	switch {
	case c.Source != nil:
		// out-of-line: element must be empty regardless of type
	case typ == TypeXhtml:
		div := struct {
			XMLName xml.Name `xml:"div"`
			XMLNS   string   `xml:"xmlns,attr"`
			Inner   string   `xml:",innerxml"`
		}{XMLName: xml.Name{Local: "div"}, XMLNS: "http://www.w3.org/1999/xhtml", Inner: *c.XHTML}
		if err := e.Encode(div); err != nil {
			return err
		}
	case typ == TypeText || typ == TypeHtml || strings.HasPrefix(string(typ), "text/"):
		if err := e.EncodeToken(xml.CharData(*c.Text)); err != nil {
			return err
		}
	default:
		if err := e.EncodeToken(xml.CharData(base64.StdEncoding.EncodeToString(c.Base64))); err != nil {
			return err
		}
	}
	return e.EncodeToken(start.End())
}

func (c *Content) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	typ := TypeText
	for _, a := range start.Attr {
		switch {
		case a.Name.Local == "type" && a.Name.Space == "":
			typ = Type(a.Value)
		case a.Name.Local == "src" && a.Name.Space == "":
			c.Source = &a.Value
		case a.Name.Local == "base" && a.Name.Space == "xml":
			c.Base = &a.Value
		case a.Name.Local == "lang" && a.Name.Space == "xml":
			c.Lang = &a.Value
		}
	}
	c.Type = &typ

	if c.Source != nil {
		return d.Skip() // element is empty; nothing further to decode
	}

	switch {
	case typ == TypeXhtml:
		var wrapper struct {
			Div struct {
				Inner string `xml:",innerxml"`
			} `xml:"div"`
		}
		if err := d.DecodeElement(&wrapper, &start); err != nil {
			return err
		}
		c.XHTML = new(strings.TrimSpace(wrapper.Div.Inner))
		return nil
	case typ == TypeText || typ == TypeHtml || strings.HasPrefix(string(typ), "text/"):
		var v struct {
			Value string `xml:",chardata"`
		}
		if err := d.DecodeElement(&v, &start); err != nil {
			return err
		}
		c.Text = &v.Value
		return nil
	case isXMLMediaType(typ):
		var v struct {
			Inner string `xml:",innerxml"`
		}
		if err := d.DecodeElement(&v, &start); err != nil {
			return err
		}
		c.XML = new(strings.TrimSpace(v.Inner))
		return nil
	default:
		var v struct {
			Value string `xml:",chardata"`
		}
		if err := d.DecodeElement(&v, &start); err != nil {
			return err
		}
		decoded, err := base64.StdEncoding.DecodeString(strings.TrimSpace(v.Value))
		if err != nil {
			return fmt.Errorf("atom:content: invalid base64 for type %q: %w", typ, err)
		}
		c.Base64 = decoded
		return nil
	}
}

func (c Content) String() string {
	switch {
	case c.Type == nil && c.Text != nil:
		return *c.Text
	case *c.Type == TypeText || *c.Type == TypeHtml:
		return *c.Text
	case *c.Type == TypeXhtml:
		return *c.XHTML
	case isXMLMediaType(*c.Type):
		return *c.XML
	case c.Base64 != nil:
		return string(c.Base64)
	default:
		return ""
	}
}

// RequiresSummary implements the rule from §4.1.3.3 / §4.1.2: an entry containing this content MUST also contain
// atom:summary.
func (c Content) RequiresSummary() bool {
	if c.Source != nil {
		return true
	}
	typ := *c.Type
	if typ == "" || typ == TypeText || typ == TypeHtml || typ == TypeXhtml || strings.HasPrefix(string(typ), "text/") {
		return false
	}
	return !isXMLMediaType(typ) // i.e. it's the Base64 branch
}
