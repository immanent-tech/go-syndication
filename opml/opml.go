// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: 	MIT

package opml

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/immanent-tech/go-syndication/extensions"
	"github.com/immanent-tech/go-syndication/validation"
	"golang.org/x/net/html/charset"
)

const (
	// MimeType indicates the canonical mimetype for an OPML file.
	MimeType = "text/x-opml+xml"
)

// NewOPMLFromBytes generates an OPML object from the given byte array.
func NewOPMLFromBytes(b []byte) (*OPML, error) {
	var root OPML

	reader := bytes.NewReader(b)
	decoder := xml.NewDecoder(reader)
	decoder.CharsetReader = charset.NewReaderLabel
	if err := decoder.Decode(&root); err != nil {
		return nil, fmt.Errorf("could not decode byte array to OPML: %w", err)
	}

	return &root, nil
}

// NewOPML creates a new OPML object.
func NewOPML(options ...Option) *OPML {
	opml := &OPML{
		Version: "2.0",
		Head: &Head{
			DateCreated:  NewRFC822Time(time.Now().UTC()),
			DateModified: NewRFC822Time(time.Now().UTC()),
		},
	}

	for option := range slices.Values(options) {
		option(opml)
	}

	return opml
}

// Option is a functional option to apply to an OPML object.
type Option func(*OPML)

// WithTitle option sets a title for the OPML object.
func WithTitle(title string) Option {
	return func(o *OPML) {
		o.Head.Title = &title
	}
}

// WithOutlines option appends the given outlines to the OPML object.
func WithOutlines(outlines ...Outline) Option {
	return func(o *OPML) {
		o.Body = append(o.Body, outlines...)
	}
}

func (o Outline) Attr(name string) (string, bool) {
	for _, a := range o.Attrs {
		if a.Name.Local == name && a.Name.Space == "" {
			return a.Value, true
		}
	}
	return "", false
}

func (o *Outline) SetAttr(name, value string) {
	for i, a := range o.Attrs {
		if a.Name.Local == name && a.Name.Space == "" {
			o.Attrs[i].Value = value
			return
		}
	}
	o.Attrs = append(o.Attrs, xml.Attr{Name: xml.Name{Local: name}, Value: value})
}

// Convenience accessors for the "well-known" attributes named by the spec's Subscription Lists and Inclusion sections.
// These are just typed sugar over Attr/SetAttr, not separate storage.

func (o Outline) XMLURL() (string, bool)          { return o.Attr("xmlUrl") }
func (o Outline) HTMLURL() (string, bool)         { return o.Attr("htmlUrl") }
func (o Outline) FeedDescription() (string, bool) { return o.Attr("description") }
func (o Outline) FeedLanguage() (string, bool)    { return o.Attr("language") }
func (o Outline) FeedTitle() (string, bool)       { return o.Attr("title") }
func (o Outline) FeedVersion() (string, bool)     { return o.Attr("version") } // "RSS1" | "RSS" | "scriptingNews"
func (o Outline) URL() (string, bool)             { return o.Attr("url") }     // type=link/include

// EffectiveType lower-cases Type, since "the value of type attributes are not case-sensitive.".
func (o Outline) EffectiveType() string {
	if o.Type != nil {
		return strings.ToLower(*o.Type)
	}
	return ""
}

func (o Outline) EffectiveIsComment() bool    { return o.IsComment != nil && *o.IsComment }
func (o Outline) EffectiveIsBreakpoint() bool { return o.IsBreakpoint != nil && *o.IsBreakpoint }

func (o Outline) Validate() error {
	if err := validation.ValidateStruct(&o); err != nil {
		return fmt.Errorf("validate outline: %w", err)
	}
	switch o.EffectiveType() {
	case "rss":
		if _, ok := o.XMLURL(); !ok {
			return errors.New("type=rss requires an xmlUrl attribute")
		}
	case "link", "include":
		if _, ok := o.URL(); !ok {
			return errors.New("type=link|include requires a url attribute")
		}
	}

	return nil
}

var opmlVersionPattern = regexp.MustCompile(`^[0-9]+\.[0-9]+$`)

func (o OPML) MarshalXML(enc *xml.Encoder, start xml.StartElement) error {
	start.Name = xml.Name{Local: "opml"}
	version := o.Version
	if version == "" {
		version = "2.0"
	}
	start.Attr = []xml.Attr{{Name: xml.Name{Local: "version"}, Value: version}}

	seen := map[string]bool{}
	namespaces := make([]extensions.Namespace, 0, len(o.Namespaces))
	for _, ns := range o.Namespaces {
		if ns.Prefix == "" || ns.URI == "" || seen[ns.Prefix] {
			continue
		}
		seen[ns.Prefix] = true
		namespaces = append(namespaces, ns)
	}
	sort.Slice(namespaces, func(i, j int) bool { return namespaces[i].Prefix < namespaces[j].Prefix })
	for _, ns := range namespaces {
		start.Attr = append(start.Attr, xml.Attr{Name: xml.Name{Local: "xmlns:" + ns.Prefix}, Value: ns.URI})
	}

	type opmlAlias OPML // sheds OPML's own MarshalXML, breaking recursion
	if err := enc.EncodeElement(opmlAlias(o), start); err != nil {
		return fmt.Errorf("marshal opml: %w", err)
	}
	return nil
}

func (o *OPML) UnmarshalXML(dec *xml.Decoder, start xml.StartElement) error {
	var version string
	var namespaces []extensions.Namespace
	for attr := range slices.Values(start.Attr) {
		switch {
		case attr.Name.Local == "version" && attr.Name.Space == "":
			version = attr.Value
		case attr.Name.Space == "xmlns":
			namespaces = append(namespaces, extensions.Namespace{Prefix: attr.Name.Local, URI: attr.Value})
		case strings.HasPrefix(attr.Name.Local, "xmlns:"):
			namespaces = append(
				namespaces,
				extensions.Namespace{Prefix: strings.TrimPrefix(attr.Name.Local, "xmlns:"), URI: attr.Value},
			)
		}
	}
	type opmlAlias OPML
	var alias opmlAlias
	if err := dec.DecodeElement(&alias, &start); err != nil {
		return fmt.Errorf("unmarshal opml: %w", err)
	}
	*o = OPML(alias)
	o.Version = version
	o.Namespaces = namespaces
	return nil
}

func (o OPML) Validate() error {
	if err := validation.ValidateStruct(o); err != nil {
		return fmt.Errorf("validate opml: %w", err)
	}
	if !opmlVersionPattern.MatchString(o.Version) {
		return fmt.Errorf("validate opml: version %q must be of the form x.y", o.Version)
	}
	return nil
}

// RFC 822 date-times (dateCreated/dateModified in <head>, an ELEMENT context). Per the spec's own note: "the year may
// be expressed with two characters or four characters (four preferred)."

const rfc822OutputLayout = "Mon, 02 Jan 2006 15:04:05 GMT"

var namedZoneOffsets = map[string]int{
	"UT": 0, "GMT": 0, "Z": 0,
	"EST": -5 * 3600, "EDT": -4 * 3600,
	"CST": -6 * 3600, "CDT": -5 * 3600,
	"MST": -7 * 3600, "MDT": -6 * 3600,
	"PST": -8 * 3600, "PDT": -7 * 3600,
}

var rfc822Layouts = []string{
	"Mon, 02 Jan 2006 15:04:05 -0700",
	"Mon, 02 Jan 06 15:04:05 -0700",
	"02 Jan 2006 15:04:05 -0700",
	"02 Jan 06 15:04:05 -0700",
	"Mon, 02 Jan 2006 15:04 -0700",
	"Mon, 02 Jan 06 15:04 -0700",
}

func ParseRFC822(timestamp string) (time.Time, error) {
	timestamp = strings.TrimSpace(timestamp)
	fields := strings.Fields(timestamp)
	if len(fields) == 0 {
		return time.Time{}, errors.New("opml: empty date-time value")
	}
	last := len(fields) - 1
	if off, ok := namedZoneOffsets[strings.ToUpper(fields[last])]; ok {
		sign := "+"
		if off < 0 {
			sign, off = "-", -off
		}
		fields[last] = fmt.Sprintf("%s%02d%02d", sign, off/3600, (off%3600)/60)
		timestamp = strings.Join(fields, " ")
	}
	var lastErr error
	for layout := range slices.Values(rfc822Layouts) {
		if t, err := time.Parse(layout, timestamp); err == nil {
			return t, nil
		} else {
			lastErr = err
		}
	}
	return time.Time{}, fmt.Errorf("opml: could not parse date-time %q: %w", timestamp, lastErr)
}

func NewRFC822Time(value time.Time) *RFC822Time {
	return &RFC822Time{Time: value}
}

func (t RFC822Time) String() string {
	return t.Time.Format(rfc822OutputLayout)
}

func (t RFC822Time) MarshalXML(enc *xml.Encoder, start xml.StartElement) error {
	if t.Time.IsZero() {
		return fmt.Errorf("marshal time: zero time.Time for <%s>", start.Name.Local)
	}
	if err := enc.EncodeToken(start); err != nil {
		return fmt.Errorf("marshal time: %w", err)
	}
	if err := enc.EncodeToken(xml.CharData(t.Time.UTC().Format(rfc822OutputLayout))); err != nil {
		return fmt.Errorf("marshal time: %w", err)
	}
	if err := enc.EncodeToken(start.End()); err != nil {
		return fmt.Errorf("marshal time: %w", err)
	}
	return nil
}

func (t *RFC822Time) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var v struct {
		Value string `xml:",chardata"`
	}
	if err := d.DecodeElement(&v, &start); err != nil {
		return fmt.Errorf("unmarshal time: %w", err)
	}
	parsed, err := ParseRFC822(v.Value)
	if err != nil {
		return fmt.Errorf("<%s>: %w", start.Name.Local, err)
	}
	t.Time = parsed
	return nil
}

func NewRFC822AttrTime(value time.Time) *RFC822AttrTime {
	return &RFC822AttrTime{Time: value}
}

func (t RFC822AttrTime) String() string {
	return t.Time.Format(rfc822OutputLayout)
}

func (t RFC822AttrTime) MarshalXMLAttr(name xml.Name) (xml.Attr, error) {
	if t.Time.IsZero() {
		return xml.Attr{}, nil // empty Name -> encoding/xml omits the attribute entirely
	}
	return xml.Attr{Name: name, Value: t.Time.UTC().Format(rfc822OutputLayout)}, nil
}

func (t *RFC822AttrTime) UnmarshalXMLAttr(attr xml.Attr) error {
	parsed, err := ParseRFC822(attr.Value)
	if err != nil {
		return fmt.Errorf("marshal time: attribute %s: %w", attr.Name.Local, err)
	}
	t.Time = parsed
	return nil
}

func (c Categories) MarshalXMLAttr(name xml.Name) (xml.Attr, error) {
	if len(c) == 0 {
		return xml.Attr{}, nil
	}
	return xml.Attr{Name: name, Value: strings.Join(c, ",")}, nil
}

func (c *Categories) UnmarshalXMLAttr(attr xml.Attr) error {
	*c = nil
	for part := range strings.SplitSeq(attr.Value, ",") {
		if s := strings.TrimSpace(part); s != "" {
			*c = append(*c, s)
		}
	}
	return nil
}

func (e ExpansionState) MarshalXML(enc *xml.Encoder, start xml.StartElement) error {
	if len(e) == 0 {
		return nil
	}
	parts := make([]string, len(e))
	for i, n := range e {
		parts[i] = strconv.Itoa(n)
	}
	if err := enc.EncodeElement(strings.Join(parts, ","), start); err != nil {
		return fmt.Errorf("marshal expansionState: %w", err)
	}
	return nil
}

func (e *ExpansionState) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var v struct {
		Value string `xml:",chardata"`
	}
	if err := d.DecodeElement(&v, &start); err != nil {
		return fmt.Errorf("unmarshal expansionState: %w", err)
	}
	*e = nil
	for part := range strings.SplitSeq(v.Value, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		n, err := strconv.Atoi(part)
		if err != nil {
			return fmt.Errorf("head/expansionState: invalid line number %q: %w", part, err)
		}
		*e = append(*e, n)
	}
	return nil
}
