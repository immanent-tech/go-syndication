// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: 	MIT

// Package rss contains objects and methods defining the RSS syndication format.
//
//revive:disable:exported // function definitions can be ascertained from Channel.
package rss

import (
	"encoding/xml"
	"errors"
	"fmt"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/immanent-tech/go-syndication/atom"
	"github.com/immanent-tech/go-syndication/extensions/rss"
	"github.com/immanent-tech/go-syndication/types"
	"github.com/immanent-tech/go-syndication/validation"
)

var _ types.FeedSource = (*RSS)(nil)

// outputLayout produces one of the profile's three recommended universal
// forms: "Thu, 04 Oct 2007 23:59:45 +0000" (i.e. UTC, numeric zero offset).
const outputLayout = "Mon, 02 Jan 2006 15:04:05 -0700"

// namedZoneOffsets maps RFC 822 zone abbreviations to their UTC offset in
// seconds. Go's time.Parse does NOT reliably resolve these itself -- an
// unrecognized "MST"-style abbreviation is silently assigned a zero offset
// by the standard library, which would misparse e.g. "EST" as UTC. We look
// these up ourselves instead of trusting the stdlib's zone-name parsing.
var namedZoneOffsets = map[string]int{
	"UT": 0, "GMT": 0, "Z": 0,
	"EST": -5 * 3600, "EDT": -4 * 3600,
	"CST": -6 * 3600, "CDT": -5 * 3600,
	"MST": -7 * 3600, "MDT": -6 * 3600,
	"PST": -8 * 3600, "PDT": -7 * 3600,
}

// dateOnlyLayouts are candidate layouts, all ending in a literal "-0700"
// placeholder for a *numeric* offset. We normalize any named zone
// abbreviation in the input to a numeric offset before trying these, so a
// single set of layouts covers both cases.
var dateOnlyLayouts = []string{
	"Mon, 02 Jan 2006 15:04:05 -0700",
	"Mon, 02 Jan 06 15:04:05 -0700",
	"02 Jan 2006 15:04:05 -0700",
	"02 Jan 06 15:04:05 -0700",
	"Mon, 02 Jan 2006 15:04 -0700", // seconds sometimes omitted in the wild
	"Mon, 02 Jan 06 15:04 -0700",
	"02 Jan 2006 15:04 -0700",
	"02 Jan 06 15:04 -0700",
}

// String returns the value of the Category.
func (c Category) String() string {
	return c.Value
}

// NewRSS creates a new RSS version 2.0 object with the required title, description, and link values and any given
// options.
func NewRSS(title, description, link string, options ...RSSOption) *RSS {
	rss := &RSS{
		Version: N20,
		Channel: Channel{
			Title:         title,
			Description:   description,
			Link:          link,
			LastBuildDate: NewTimestamp(time.Now().UTC()),
			Generator:     new("go-syndication"),
			Docs:          new("https://www.rssboard.org/rss-specification"),
		},
	}

	for option := range slices.Values(options) {
		option(rss)
	}

	return rss
}

// RSSOption is a functional applied to an RSS object.
type RSSOption func(*RSS)

func WithAtomLink(link *atom.Link) RSSOption {
	return func(r *RSS) {
		r.Channel.AtomLink = link
	}
}

// WithCopyright option sets the RSS channel copyright.
func WithCopyright(copyright string) RSSOption {
	return func(r *RSS) {
		r.Channel.Copyright = new(copyright)
	}
}

// WithGenerator options sets the generator. This will default to "go-syndication".
func WithGenerator(generator string) RSSOption {
	return func(r *RSS) {
		r.Channel.Generator = new(generator)
	}
}

// WithManagingEditor option sets the RSS channel managingEditor.
func WithManagingEditor(editor string) RSSOption {
	return func(r *RSS) {
		r.Channel.ManagingEditor = new(editor)
	}
}

// WithWebmaster option sets the RSS channel webmaster.
func WithWebmaster(webmaster string) RSSOption {
	return func(r *RSS) {
		r.Channel.WebMaster = new(webmaster)
	}
}

// WithLastBuildDate option sets the last build date of the RSS object. This will default to time.Now().UTC().
func WithLastBuildDate(ts time.Time) RSSOption {
	return func(r *RSS) {
		r.Channel.LastBuildDate = NewTimestamp(ts)
	}
}

// WithPublishedDate option sets the published date of the RSS object.
func WithPublishedDate(ts time.Time) RSSOption {
	return func(r *RSS) {
		r.Channel.PubDate = NewTimestamp(ts)
	}
}

// WithChannelLanguage option sets the RSS channel language. Should be an ISO country code to be valid.
func WithChannelLanguage(lang string) RSSOption {
	return func(r *RSS) {
		r.Channel.Language = new(lang)
	}
}

// WithChannelImage option sets an image.
func WithChannelImage(image *Image) RSSOption {
	return func(r *RSS) {
		r.Channel.Image = image
	}
}

// WithUpdatePeriod option sets the update period of the feed.
func WithUpdatePeriod(up string) RSSOption {
	return func(r *RSS) {
		r.Channel.SYUdatePeriod = &rss.SYUpdatePeriod{
			Value: up,
		}
	}
}

// WithUpdateFrequency option sets the update frequency of the feed.
func WithUpdateFrequency(freq int) RSSOption {
	return func(r *RSS) {
		r.Channel.SYUpdateFrequency = &rss.SYUpdateFrequency{
			Value: freq,
		}
	}
}

func (r *RSS) GetTitle() string {
	return r.Channel.GetTitle()
}

func (r *RSS) GetDescription() string {
	return r.Channel.GetDescription()
}

func (r *RSS) GetSourceURL() string {
	return r.Channel.GetSourceURL()
}

func (r *RSS) SetSourceURL(url string) {
	r.Channel.SetSourceURL(url)
}

func (r *RSS) GetLink() string {
	return r.Channel.GetLink()
}

func (r *RSS) GetUpdatedDate() *time.Time {
	return r.Channel.GetUpdatedDate()
}

func (r *RSS) GetPublishedDate() *time.Time {
	return r.Channel.GetPublishedDate()
}

func (r *RSS) GetCategories() []string {
	return r.Channel.GetCategories()
}

func (r *RSS) GetAuthors() []string {
	return r.Channel.GetAuthors()
}

func (r *RSS) GetContributors() []string {
	return r.Channel.GetContributors()
}

func (r *RSS) GetRights() *string {
	return r.Channel.GetRights()
}

func (r *RSS) GetLanguage() *string {
	return r.Channel.GetLanguage()
}

func (r *RSS) GetImage() *types.ImageInfo {
	return r.Channel.GetImage()
}

func (r *RSS) SetImage(image *types.ImageInfo) {
	r.Channel.SetImage(image)
}

func (r *RSS) GetItems() []types.ItemSource {
	return r.Channel.GetItems()
}

func (r *RSS) GetUpdateInterval() time.Duration {
	return r.Channel.GetUpdateInterval()
}

// Validate applies custom validation to an feed.
func (r *RSS) Validate() error {
	if err := validation.ValidateStruct(r); err != nil {
		return fmt.Errorf("rss validation failed: %w", err)
	}
	return nil
}

// MarshalXML implements xml.Marshaler. It builds the xmlns:* attribute list
// from r.Namespaces at encode time -- this is the only way to get a
// *dynamic* set of attributes out of encoding/xml, since struct tags are
// necessarily static.
func (r RSS) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	version := r.Version
	if version == "" {
		version = N201
	}
	start.Name = xml.Name{Local: "rss"}
	start.Attr = []xml.Attr{{Name: xml.Name{Local: "version"}, Value: string(version)}}

	// De-duplicate by prefix and sort for deterministic, diffable output.
	seen := make(map[string]bool, len(r.Namespaces))
	namespaces := make([]types.Namespace, 0, len(r.Namespaces))
	for _, ns := range r.Namespaces {
		if ns.Prefix == "" || ns.URI == "" || seen[ns.Prefix] {
			continue
		}
		seen[ns.Prefix] = true
		namespaces = append(namespaces, ns)
	}
	sort.Slice(namespaces, func(i, j int) bool { return namespaces[i].Prefix < namespaces[j].Prefix })

	for _, ns := range namespaces {
		// Literal "xmlns:prefix" as the attribute's local name -- the same
		// workaround used elsewhere for namespaced element/attribute names,
		// since encoding/xml doesn't manage prefix allocation for us.
		start.Attr = append(start.Attr, xml.Attr{
			Name:  xml.Name{Local: "xmlns:" + ns.Prefix},
			Value: ns.URI,
		})
	}

	if err := e.EncodeToken(start); err != nil {
		return fmt.Errorf("encode rss: %w", err)
	}
	// e.Encode(r.Channel) would ignore the "channel" tag on the RSS.Channel
	// field (struct-field tags only apply when encoding/xml walks a
	// struct's fields itself; a custom MarshalXML steps outside that path).
	// EncodeElement with an explicit start tag fixes it.
	channelStart := xml.StartElement{Name: xml.Name{Local: "channel"}}
	if err := e.EncodeElement(r.Channel, channelStart); err != nil {
		return fmt.Errorf("encode channel: %w", err)
	}
	return e.EncodeToken(start.End())
}

// UnmarshalXML implements xml.Unmarshaler. It recovers whichever namespaces
// the source document actually declared -- known or not -- so a
// decode/re-encode round trip preserves them.
func (r *RSS) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	for attr := range slices.Values(start.Attr) {
		switch {
		case attr.Name.Local == "version" && attr.Name.Space == "":
			r.Version = RSSVersion(attr.Value)
		case attr.Name.Space == "xmlns":
			// Go's decoder represents "xmlns:foo" attributes this way.
			r.Namespaces = append(r.Namespaces, types.Namespace{Prefix: attr.Name.Local, URI: attr.Value})
		case strings.HasPrefix(attr.Name.Local, "xmlns:"):
			// Defensive fallback in case some parser instead hands us the
			// literal, unsplit form.
			r.Namespaces = append(r.Namespaces, types.Namespace{
				Prefix: strings.TrimPrefix(attr.Name.Local, "xmlns:"),
				URI:    attr.Value,
			})
		}
	}
	var wrapper struct {
		Channel Channel `xml:"channel"`
	}
	if err := d.DecodeElement(&wrapper, &start); err != nil {
		return fmt.Errorf("rss decode: %w", err)
	}
	r.Channel = wrapper.Channel
	return nil
}

// AutoDeclareNamespaces inspects the populated extension fields across the
// channel and its items and appends any missing namespace declarations for
// the extensions this package knows how to model (content, media, atom,
// dc, slash, syn). This directly targets the "unbound prefix" bug: as long
// as you call this before marshaling, you can't forget to declare a
// namespace for a typed extension field you populated.
//
// It does NOT cover ExtensionElement catch-all content or any custom
// namespace you use yourself -- for those, append to r.Namespaces (or use
// NS(...)) manually, since there's no reliable way to recover the intended
// prefix from a raw namespace URI alone.
func (r *RSS) AutoDeclareNamespaces() {
	need := map[string]bool{}
	if r.Channel.AtomLink != nil {
		need["atom"] = true
	}
	if r.Channel.SYUdatePeriod != nil || r.Channel.SYUpdateFrequency != nil {
		need["syn"] = true
	}
	for item := range slices.Values(r.Channel.Items) {
		if item.ContentEncoded != nil {
			need["content"] = true
		}
		if len(item.MediaThumbnails) > 0 {
			need["media"] = true
		}
		if item.Creator != nil {
			need["dc"] = true
		}
	}

	existing := make(map[string]bool, len(r.Namespaces))
	for _, ns := range r.Namespaces {
		existing[ns.Prefix] = true
	}
	for prefix := range need {
		if !existing[prefix] {
			r.Namespaces = append(r.Namespaces, types.NewNamespace(prefix))
		}
	}
}

// ParseRFC822 parses an RSS date-time value leniently: it accepts both
// numeric zone offsets (+0100, -0600) and the named zone abbreviations
// registered in namedZoneOffsets, with or without a weekday, with a 2- or
// 4-digit year, and with or without seconds.
func ParseRFC822(ts string) (time.Time, error) {
	ts = strings.TrimSpace(ts)

	fields := strings.Fields(ts)
	if len(fields) == 0 {
		return time.Time{}, errors.New("rss date-time: empty value")
	}
	lastIdx := len(fields) - 1
	zone := fields[lastIdx]

	// If the trailing token is a known named zone, rewrite it as a
	// numeric offset so a single family of layouts handles everything.
	if off, ok := namedZoneOffsets[strings.ToUpper(zone)]; ok {
		sign := "+"
		if off < 0 {
			sign = "-"
			off = -off
		}
		fields[lastIdx] = fmt.Sprintf("%s%02d%02d", sign, off/3600, (off%3600)/60)
		ts = strings.Join(fields, " ")
	}
	// Otherwise, if it's already a numeric offset (+0100 / -0600) or a
	// literal "Z", leave it as-is; the loop below will try it against
	// each layout and Go's -0700 verb correctly parses "+HHMM"/"-HHMM".

	var lastErr error
	for _, layout := range dateOnlyLayouts {
		if t, err := time.Parse(layout, ts); err == nil {
			return t, nil
		} else {
			lastErr = err
		}
	}
	return time.Time{}, fmt.Errorf("rss date-time: could not parse %q: %w", ts, lastErr)
}

// IsCanonical reports whether s is already in one of the profile's three
// recommended universal forms -- "... +0000", "... -0000", or "... GMT" --
// with a well-formed date-time prefix. Useful for producers who want to
// flag non-canonical input before emitting it verbatim.
func IsCanonical(ts string) bool {
	switch {
	case strings.HasSuffix(ts, " GMT"):
		_, err := time.Parse("Mon, 02 Jan 2006 15:04:05 GMT", ts)
		return err == nil
	case strings.HasSuffix(ts, " +0000"), strings.HasSuffix(ts, " -0000"):
		_, err := time.Parse(outputLayout, ts)
		return err == nil
	default:
		return false
	}
}

func NewTimestamp(value time.Time) *Timestamp {
	return &Timestamp{Value: value}
}

func (t Timestamp) String() string {
	return t.Value.Format(outputLayout)
}

// MarshalXML implements xml.Marshaler. Always normalizes to UTC and emits
// the "+0000" form, one of the profile's three recommended universal
// representations.
func (t Timestamp) MarshalXML(enc *xml.Encoder, start xml.StartElement) error {
	if t.Value.IsZero() {
		return fmt.Errorf("rss timestamp: zero time.Time value for <%s>", start.Name.Local)
	}
	if err := enc.EncodeToken(start); err != nil {
		return fmt.Errorf("rss timestamp: encode start element: %w", err)
	}
	formatted := t.Value.UTC().Format(outputLayout)
	if err := enc.EncodeToken(xml.CharData(formatted)); err != nil {
		return fmt.Errorf("rss timestamp: encode: %w", err)
	}

	if err := enc.EncodeToken(start.End()); err != nil {
		return fmt.Errorf("rss timestamp: encode end element: %w", err)
	}

	return nil
}

// UnmarshalXML implements xml.Unmarshaler, accepting any RFC 822-conformant
// value (per the profile's requirements) rather than only the canonical
// output forms.
func (t *Timestamp) UnmarshalXML(dec *xml.Decoder, start xml.StartElement) error {
	var valueStruct struct {
		Value string `xml:",chardata"`
	}
	if err := dec.DecodeElement(&valueStruct, &start); err != nil {
		return fmt.Errorf("rss timestamp: decode start element: %w", err)
	}
	parsed, err := ParseRFC822(valueStruct.Value)
	if err != nil {
		return fmt.Errorf("<%s>: %w", start.Name.Local, err)
	}
	t.Value = parsed
	return nil
}
