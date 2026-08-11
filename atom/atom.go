// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: 	MIT

// Package atom contains objects and methods defining the Atom syndication format.
package atom

import (
	"encoding/base64"
	"encoding/xml"
	"errors"
	"fmt"
	"html"
	"slices"
	"strings"
	"time"

	"github.com/immanent-tech/go-syndication/sanitization"
	"github.com/immanent-tech/go-syndication/validation"
)

var (
	// MimeTypes contains canonical/standard mimetypes for Atom.
	MimeTypes = []string{"application/atom+xml"}
)

const atomNS = "http://www.w3.org/1999/xhtml"
const xmlNS = "http://www.w3.org/XML/1998/namespace"

// dateLayout mirrors time.RFC3339Nano: "2006-01-02T15:04:05.999999999Z07:00". The trailing ".999999999" is Go's
// convention for "trim trailing zero fractional digits, omit entirely if zero". This naturally produces the spec's
// *optional* fractional-seconds behavior. The literal "T" and the "Z07:00" zone verb naturally produce uppercase "T"
// and "Z" on output, exactly matching the spec's requirement.
const dateLayout = time.RFC3339Nano

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

func (c Category) Validate() error {
	if err := validation.ValidateStruct(c); err != nil {
		return fmt.Errorf("category: %w", err)
	}
	return nil
}

// String formats the generator value as a string in the format VALUE[/VERSION] [(URI)].
func (g Generator) String() string {
	var gen strings.Builder
	gen.WriteString(strings.TrimSpace(g.Value))
	if g.Version != nil && *g.Version != "" {
		gen.WriteString("/")
		gen.WriteString(*g.Version)
	}
	if g.URI != nil && *g.URI != "" {
		gen.WriteString(" ")
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
	if l.Rel != nil {
		if *l.Rel == "" {
			return errors.New("atom:link: rel must not be empty")
		}
		if *l.Rel == LinkRelEnclosure && l.Length != nil {
			// SHOULD, not MUST -- not a hard error, but worth flagging.
			return fmt.Errorf("atom:link: rel=%q SHOULD include a length attribute", LinkRelEnclosure)
		}
	}
	if l.Title != nil {
		if *l.Title == "" {
			return errors.New("atom:link: title cannot be blank")
		}
	}
	if l.Type != nil {
		if *l.Type == "" {
			return errors.New("atom:link: type cannot be blank")
		}
	}
	if err := validation.ValidateStruct(l); err != nil {
		return fmt.Errorf("validate atom:link: %w", err)
	}
	return nil
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

func (p PersonConstruct) Validate() error {
	if err := validation.ValidateStruct(p); err != nil {
		return fmt.Errorf("validate person construct: %w", err)
	}
	if err := validation.ValidateField(p.Name, "html"); err == nil {
		return errors.New("validate person construct: name cannot contain HTML")
	}
	return nil
}

func (t TextConstruct) String() string {
	switch {
	case t.Type == nil || (*t.Type != TextConstructTypeXhtml && !contentIsXMLMediaType(ContentType(*t.Type))):
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
	var typ TextConstructType
	if t.Type == nil {
		typ = TextConstructTypeText
	} else {
		typ = *t.Type
	}
	if typ != TextConstructTypeText && typ != TextConstructTypeHtml && typ != TextConstructTypeXhtml {
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

	if typ == TextConstructTypeXhtml {
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
	typ := TextConstructTypeText // spec: absent type attribute defaults to "text"
	for attr := range slices.Values(start.Attr) {
		switch {
		case attr.Name.Local == "type" && attr.Name.Space == "":
			typ = TextConstructType(attr.Value)
		case attr.Name.Local == "lang" && attr.Name.Space == xmlNS:
			t.Lang = new(attr.Value)
		case attr.Name.Local == "base" && attr.Name.Space == xmlNS:
			t.Base = new(attr.Value)
		}
	}
	t.Type = new(typ)

	if typ == TextConstructTypeXhtml {
		var wrapper struct {
			Div struct {
				Inner string `xml:",innerxml"`
			} `xml:"div"` // matches any namespace's local-name "div"
			Inner string `xml:",innerxml"` // fallback capture of everything, div or not.
		}

		if err := dec.DecodeElement(&wrapper, &start); err != nil {
			return fmt.Errorf("text construct: unmarshal: %w", err)
		}
		t.XHTML = new(strings.TrimSpace(wrapper.Div.Inner))
		if *t.XHTML == "" {
			// Sloppy producer: type="xhtml" without the spec-required wrapping <div>. Rather than silently ending up
			// with an empty XHTML field, fall back to whatever markup is directly present.
			t.XHTML = new(strings.TrimSpace(wrapper.Inner))
		}
		return nil
	}
	// Leniency for non-conformant producers that put a MIME type here that really belongs on atom:content (e.g.
	// "application/xhtml+xml" instead of the spec's own "xhtml"). Unlike the xhtml case above, there's no required
	// wrapper element to unwrap; whatever child markup is present (e.g. a bare <code>...</code>) is captured as-is.
	if contentIsXMLMediaType(ContentType(typ)) {
		var wrapper struct {
			Inner string `xml:",innerxml"`
		}
		if err := dec.DecodeElement(&wrapper, &start); err != nil {
			return fmt.Errorf("text construct: unmarshal: %w", err)
		}
		t.XHTML = new(strings.TrimSpace(wrapper.Inner))
		return nil
	}

	// "text" and "html" (and, leniently, anything else): plain character data. The decoder already unescapes entities
	// for us, so for "html" content this correctly yields real markup back as a Go string.
	var valueStruct struct {
		Value string `xml:",chardata"`
		Inner string `xml:",innerxml"`
	}
	if err := dec.DecodeElement(&valueStruct, &start); err != nil {
		return fmt.Errorf("text construct: unmarshal: %w", err)
	}
	if strings.TrimSpace(valueStruct.Value) != "" {
		t.Value = valueStruct.Value
	} else {
		t.Value = strings.TrimSpace(valueStruct.Inner)
	}
	return nil
}

func (t TextConstruct) Validate() error {
	if t.Type != nil {
		if *t.Type == "" {
			return errors.New("text construct: type cannot be empty")
		}
		switch {
		case *t.Type == TextConstructTypeXhtml:
			// Must not be inline when type is XHTML.
			if strings.HasPrefix(*t.XHTML, "<![CDATA[") {
				return fmt.Errorf("text construct: cannot contain inline content for type %s", TextConstructTypeXhtml)
			}
			if html.UnescapeString(*t.XHTML) != *t.XHTML {
				return fmt.Errorf("text construct: cannot be escaped for type %s", TextConstructTypeXhtml)
			}
		case *t.Type == TextConstructTypeHtml || strings.HasSuffix(string(*t.Type), "html"):
			// Must be URL encoded when type is HTML.
			if err := validation.ValidateField(t.Value, "not_url_encoded"); err == nil {
				return fmt.Errorf("text construct: must be url encoded for type %s", TextConstructTypeHtml)
			}
			// Must be valid HTML.
			if err := validation.ValidateField(t.Value, "html"); err != nil {
				return fmt.Errorf("text construct: not html for type %s", TextConstructTypeText)
			}
		case *t.Type == TextConstructTypeText:
			fallthrough
		default:
			// For text type and by default, the value should not contain html.
			if err := validation.ValidateField(t.Value, "html"); err == nil {
				return fmt.Errorf("text construct: must not contain html for type %s", TextConstructTypeText)
			}
			if html.UnescapeString(t.Value) != t.Value {
				return errors.New("text construct: plain text must not contain escaped characters")
			}
		}
	}
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
		case a.Name.Local == "base" && a.Name.Space == xmlNS:
			d.Base = &a.Value
		case a.Name.Local == "lang" && a.Name.Space == xmlNS:
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
func (d DateConstruct) Validate() error {
	raw := d.String()
	if _, err := time.Parse(time.RFC3339, raw); err != nil {
		return fmt.Errorf("date construct: invalid date-time %q: %w", raw, err)
	}
	// time.Parse accepts lowercase t/z against this layout too; the spec doesn't, so check the literal separator
	// characters ourselves. "2006-01-02" is always 10 bytes before the separator
	if tIdx := 10; tIdx >= len(raw) || raw[tIdx] != 'T' {
		return fmt.Errorf("date construct: %q must use an uppercase %q separator", raw, "T")
	}
	if raw[len(raw)-1] == 'z' {
		return fmt.Errorf("date construct: %q must use an uppercase %q zone indicator", raw, "Z")
	}
	return nil
}

func contentIsXMLMediaType(t ContentType) bool {
	return t != ContentTypeText && t != ContentTypeHtml && t != ContentTypeXhtml &&
		!strings.HasPrefix(string(t), "text/") &&
		(strings.HasSuffix(string(t), "+xml") || strings.HasSuffix(string(t), "/xml"))
}

func (c Content) MarshalXML(enc *xml.Encoder, start xml.StartElement) error {
	var typ ContentType
	if c.Type == nil {
		typ = ContentTypeText
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
	if c.Source == nil && contentIsXMLMediaType(typ) {
		wrapper := struct {
			Inner string `xml:",innerxml"`
		}{Inner: *c.XML}
		if err := enc.EncodeElement(wrapper, start); err != nil {
			return fmt.Errorf("marshal content: %w", err)
		}
		return nil
	}

	if err := enc.EncodeToken(start); err != nil {
		return fmt.Errorf("marshal content: %w", err)
	}
	switch {
	case c.Source != nil:
		// out-of-line: element must be empty regardless of type
	case typ == ContentTypeXhtml:
		div := struct {
			XMLName xml.Name `xml:"div"`
			XMLNS   string   `xml:"xmlns,attr"`
			Inner   string   `xml:",innerxml"`
		}{XMLName: xml.Name{Local: "div"}, XMLNS: "http://www.w3.org/1999/xhtml", Inner: *c.XHTML}
		if err := enc.Encode(div); err != nil {
			return fmt.Errorf("marshal content: %w", err)
		}
	case typ == ContentTypeText || typ == ContentTypeHtml || strings.HasPrefix(string(typ), "text/"):
		if err := enc.EncodeToken(xml.CharData(*c.Text)); err != nil {
			return fmt.Errorf("marshal content: %w", err)
		}
	default:
		if err := enc.EncodeToken(xml.CharData(base64.StdEncoding.EncodeToString(c.Base64))); err != nil {
			return fmt.Errorf("marshal content: %w", err)
		}
	}
	if err := enc.EncodeToken(start.End()); err != nil {
		return fmt.Errorf("marshal content: %w", err)
	}
	return nil
}

func (c *Content) UnmarshalXML(dec *xml.Decoder, start xml.StartElement) error {
	typ := ContentTypeText
	for attr := range slices.Values(start.Attr) {
		switch {
		case attr.Name.Local == "type" && attr.Name.Space == "":
			typ = ContentType(attr.Value)
		case attr.Name.Local == "src" && attr.Name.Space == "":
			c.Source = &attr.Value
		case attr.Name.Local == "base" && attr.Name.Space == xmlNS:
			c.Base = &attr.Value
		case attr.Name.Local == "lang" && attr.Name.Space == xmlNS:
			c.Lang = &attr.Value
		}
	}
	c.Type = &typ

	if c.Source != nil {
		if err := dec.Skip(); err != nil { // element is empty; nothing further to decode
			return fmt.Errorf("unmarshal content: %w", err)
		}
		return nil
	}

	switch {
	case typ == ContentTypeXhtml:
		var wrapper struct {
			Div struct {
				Inner string `xml:",innerxml"`
			} `xml:"div"`
			Inner string `xml:",innerxml"` // fallback capture of everything, div or not.
		}
		if err := dec.DecodeElement(&wrapper, &start); err != nil {
			return fmt.Errorf("unmarshal content: %w", err)
		}
		c.XHTML = new(strings.TrimSpace(wrapper.Div.Inner))
		if *c.XHTML == "" {
			// Sloppy producer: type="xhtml" without the spec-required wrapping <div>. Rather than silently ending up
			// with an empty XHTML field, fall back to whatever markup is directly present.
			c.XHTML = new(strings.TrimSpace(wrapper.Inner))
		}
		return nil
	case typ == ContentTypeText || typ == ContentTypeHtml || strings.HasPrefix(string(typ), "text/"):
		// Spec-conformant content here is either entity-escaped text or a CDATA section -- both come through fine via
		// ",chardata" alone (Go's tokenizer treats CDATA and escaped text identically, so no special CDATA handling is
		// needed). What ISN'T handled by ",chardata" alone: a sloppy producer that puts real, unescaped HTML markup
		// here as actual child elements instead of escaping it. ",chardata" silently ignores element children -- it
		// wouldn't error, it would just leave c.Text as blank/whitespace, discarding real content with no signal
		// anything went wrong. Capturing innerxml alongside chardata and falling back to it when chardata is empty
		// closes that gap: still prefer the spec-conformant text when present, but don't silently drop content that's
		// merely mis-encoded rather than absent.
		var v struct {
			Value string `xml:",chardata"`
			Inner string `xml:",innerxml"`
		}
		if err := dec.DecodeElement(&v, &start); err != nil {
			return fmt.Errorf("unmarshal content: %w", err)
		}
		if strings.TrimSpace(v.Value) != "" {
			c.Text = &v.Value
		} else {
			c.Text = new(strings.TrimSpace(v.Inner))
		}
		return nil
	case contentIsXMLMediaType(typ):
		var v struct {
			Inner string `xml:",innerxml"`
		}
		if err := dec.DecodeElement(&v, &start); err != nil {
			return fmt.Errorf("unmarshal content: %w", err)
		}
		c.XML = new(strings.TrimSpace(v.Inner))
		return nil
	default:
		var v struct {
			Value string `xml:",chardata"`
		}
		if err := dec.DecodeElement(&v, &start); err != nil {
			return fmt.Errorf("unmarshal content: %w", err)
		}
		decoded, err := base64.StdEncoding.DecodeString(strings.TrimSpace(v.Value))
		if err != nil {
			return fmt.Errorf("unmarshal content: invalid base64 for type %q: %w", typ, err)
		}
		c.Base64 = decoded
		return nil
	}
}

func (c Content) String() string {
	switch {
	case c.Type == nil && c.Text != nil:
		return *c.Text
	case *c.Type == ContentTypeText || *c.Type == ContentTypeHtml || strings.HasPrefix(string(*c.Type), "text/"):
		return *c.Text
	case *c.Type == ContentTypeXhtml:
		return *c.XHTML
	case contentIsXMLMediaType(*c.Type):
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
	if typ == "" || typ == ContentTypeText || typ == ContentTypeHtml || typ == ContentTypeXhtml ||
		strings.HasPrefix(string(typ), "text/") {
		return false
	}
	return !contentIsXMLMediaType(typ) // i.e. it's the Base64 branch
}

func (c Content) Validate() error {
	if err := validation.ValidateStruct(c); err != nil {
		return fmt.Errorf("validate content: %w", err)
	}
	if c.Type != nil {
		if *c.Type == "" {
			return errors.New("validate content: type cannot be empty")
		}
		switch {
		case *c.Type == ContentTypeHtml || strings.HasPrefix(string(*c.Type), "html"):
			// If it indicates it contains encoded content, validate that.
			if err := validation.ValidateField(*c.Text, "url_encoded"); err != nil {
				return fmt.Errorf("validate content: %w", err)
			}
			// Validate it is valid HTML.
			if err := validation.ValidateField(*c.Text, "html"); err != nil {
				return fmt.Errorf("validate content: %w", err)
			}
		case *c.Type == ContentTypeText || strings.Contains(string(*c.Type), "plain"):
			// Validate text does not contain escaped content.
			if html.UnescapeString(*c.Text) != *c.Text {
				return errors.New("validate content: plain text contains escape characters")
			}
		default:
			// Validate type is valid mimetype.
			if err := validation.ValidateField(*c.Type, "mimetype_string"); err != nil {
				return fmt.Errorf("validate content: %w", err)
			}
		}
	}
	if len(c.Base64) > 0 {
		// Validate the content is actually base64 encoded.
		if err := validation.ValidateField(c.Base64, "base64"); err != nil {
			return fmt.Errorf("validate content: %w", err)
		}
	}
	if !c.RequiresSummary() {
		return errors.New("validate content: requires summary")
	}
	return nil
}
