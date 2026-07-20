// Copyright 2026 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: 	AGPL-3.0-or-later

package rdf

import (
	"encoding/xml"
	"errors"
	"fmt"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/immanent-tech/go-syndication/extensions"
	"github.com/immanent-tech/go-syndication/types"
	"golang.org/x/net/html/charset"
)

const (
	rss1NS = "http://purl.org/rss/1.0/"
)

var _ types.FeedSource = (*RDF)(nil)

func (r *RDF) GetAuthors() []string {
	return r.Channel.GetAuthors()
}

func (r *RDF) GetContributors() []string {
	return r.Channel.GetContributors()
}

func (r *RDF) GetCategories() []string {
	return r.Channel.GetCategories()
}

func (r *RDF) GetDescription() string {
	return r.Channel.GetDescription()
}

func (r *RDF) GetTitle() string {
	return r.Channel.GetTitle()
}

func (r *RDF) GetLanguage() *string {
	return r.Channel.GetLanguage()
}

func (r *RDF) GetLink() string {
	return r.Channel.GetLink()
}

func (r *RDF) GetPublishedDate() *time.Time {
	return r.Channel.GetPublishedDate()
}

func (r *RDF) GetUpdatedDate() *time.Time {
	return nil
}

func (r *RDF) GetRights() *string {
	return r.Channel.GetRights()
}

func (r *RDF) GetImage() *types.ImageInfo {
	if r.Image != nil {
		return &types.ImageInfo{
			Title: r.Image.Title,
			URL:   r.Image.URL,
		}
	}
	return nil
}

func (r *RDF) SetImage(img *types.ImageInfo) {
	r.Image.Title = img.GetTitle()
	r.Image.URL = img.GetURL()
}

func (r *RDF) GetSourceURL() string {
	return r.Channel.GetSourceURL()
}

func (r *RDF) SetSourceURL(value string) {
	r.Channel.SetSourceURL(value)
}

func (r *RDF) GetItems() []types.ItemSource {
	items := make([]types.ItemSource, 0, len(r.Items))
	for item := range slices.Values(r.Items) {
		items = append(items, &item)
	}
	return items
}

func (r *RDF) GetUpdateInterval() time.Duration {
	if interval := r.Channel.GetUpdateInterval(); interval > 0 {
		return interval
	}
	if items := r.GetItems(); len(items) > 2 {
		var intervals []time.Duration
		for idx := range items {
			if idx < len(items)-1 {
				if items[idx].GetPublishedDate() != nil &&
					items[idx+1].GetPublishedDate() != nil {
					intervals = append(
						intervals,
						items[idx].GetPublishedDate().Sub(*items[idx+1].GetPublishedDate()).Abs(),
					)
				}
			}
		}
		if len(intervals) > 0 {
			return types.GetMedianInterval(intervals)
		}
	}
	return types.DefaultFeedUpdateInterval
}

// MarshalXML implements xml.Marshaler, building the default-namespace and
// xmlns:* attribute list dynamically at encode time.
func (r RDF) MarshalXML(enc *xml.Encoder, start xml.StartElement) error {
	defaultNS := r.DefaultNamespace
	if defaultNS == "" {
		defaultNS = rss1NS
	}

	start.Name = xml.Name{Local: "rdf:RDF"}
	start.Attr = []xml.Attr{{Name: xml.Name{Local: "xmlns"}, Value: defaultNS}}

	// De-duplicate by prefix, force "rdf" to be present, sort for
	// deterministic output.
	seen := map[string]bool{}
	namespaces := make([]extensions.Namespace, 0, len(r.Namespaces)+1)
	haveRDF := false
	for namespace := range slices.Values(r.Namespaces) {
		if namespace.Prefix == "" || namespace.URI == "" || seen[namespace.Prefix] {
			continue
		}
		seen[namespace.Prefix] = true
		if namespace.Prefix == "rdf" {
			haveRDF = true
		}
		namespaces = append(namespaces, namespace)
	}
	if !haveRDF {
		namespaces = append(namespaces, extensions.NewNamespace("rdf"))
	}
	sort.Slice(namespaces, func(i, j int) bool {
		// Keep "rdf" first purely for readability; the rest alphabetical.
		if namespaces[i].Prefix == "rdf" {
			return true
		}
		if namespaces[j].Prefix == "rdf" {
			return false
		}
		return namespaces[i].Prefix < namespaces[j].Prefix
	})

	for _, ns := range namespaces {
		start.Attr = append(start.Attr, xml.Attr{
			Name:  xml.Name{Local: "xmlns:" + ns.Prefix},
			Value: ns.URI,
		})
	}

	if err := enc.EncodeToken(start); err != nil {
		return fmt.Errorf("marshal rdf: %w", err)
	}
	if err := enc.Encode(r.Channel); err != nil { // Channel has its own XMLName; correct element name either way
		return fmt.Errorf("marshal rdf: %w", err)
	}
	if r.Image != nil {
		if err := enc.Encode(r.Image); err != nil {
			return fmt.Errorf("marshal rdf: %w", err)
		}
	}
	for _, item := range r.Items {
		if err := enc.Encode(item); err != nil {
			return fmt.Errorf("marshal rdf: %w", err)
		}
	}
	if r.TextInput != nil {
		if err := enc.Encode(r.TextInput); err != nil {
			return fmt.Errorf("marshal rdf: %w", err)
		}
	}
	if err := enc.EncodeToken(start.End()); err != nil {
		return fmt.Errorf("marshal rdf: %w", err)
	}
	return nil
}

// UnmarshalXML implements xml.Unmarshaler, recovering whichever namespaces the source document actually declared, so a
// decode/re-encode round trip preserves them.
func (r *RDF) UnmarshalXML(dec *xml.Decoder, start xml.StartElement) error {
	if dec.CharsetReader == nil {
		dec.CharsetReader = charset.NewReaderLabel
	}
	for attr := range slices.Values(start.Attr) {
		switch {
		case attr.Name.Local == "xmlns" && attr.Name.Space == "":
			r.DefaultNamespace = attr.Value
		case attr.Name.Space == "xmlns":
			r.Namespaces = append(r.Namespaces, extensions.NewNamespace(attr.Name.Local, attr.Value))
		case strings.HasPrefix(attr.Name.Local, "xmlns:"):
			r.Namespaces = append(r.Namespaces, extensions.NewNamespace(
				strings.TrimPrefix(attr.Name.Local, "xmlns:"),
				attr.Value,
			))
		}
	}

	var wrapper struct {
		Channel   Channel    `xml:"http://purl.org/rss/1.0/ channel"`
		Image     *Image     `xml:"http://purl.org/rss/1.0/ image"`
		Items     []Item     `xml:"http://purl.org/rss/1.0/ item"`
		TextInput *TextInput `xml:"http://purl.org/rss/1.0/ textinput"`
	}
	if err := dec.DecodeElement(&wrapper, &start); err != nil {
		return fmt.Errorf("unmarshal rdf: %w", err)
	}
	r.Channel = wrapper.Channel
	r.Image = wrapper.Image
	r.Items = wrapper.Items
	r.TextInput = wrapper.TextInput
	return nil
}

// AutoDeclareNamespaces inspects the populated Dublin Core / Syndication
// fields across the channel and its items and appends any missing
// namespace declaration. Call this before marshaling.
func (r *RDF) AutoDeclareNamespaces() {
	need := map[string]bool{}

	if r.Channel.Creator != nil || r.Channel.Date != nil ||
		r.Channel.Publisher != nil || r.Channel.Rights != nil {
		need["dc"] = true
	}
	if r.Channel.SYUdatePeriod != nil ||
		(r.Channel.SYUpdateFrequency != nil && r.Channel.SYUpdateFrequency.Value != 0) ||
		r.Channel.SYUpdateBase != nil {
		need["sy"] = true
	}
	for _, it := range r.Items {
		if it.Creator != nil || it.Date != nil || it.Subject != nil {
			need["dc"] = true
		}
	}

	existing := make(map[string]bool, len(r.Namespaces))
	for _, ns := range r.Namespaces {
		existing[ns.Prefix] = true
	}
	for prefix := range need {
		if !existing[prefix] {
			r.Namespaces = append(r.Namespaces, extensions.NewNamespace(prefix))
		}
	}
}

// Link derives all the channel's cross-reference fields (its <items> sequence, and its <image>/<textinput> resource
// pointers) from the top-level Image/Items/TextInput actually present, so callers only have to populate the top-level
// elements once rather than keep two representations manually in sync. Call this before marshaling.
func (r *RDF) Link() {
	refs := make(ItemRefs, 0, len(r.Items))
	for _, it := range r.Items {
		refs = append(refs, it.About)
	}
	r.Channel.Items = refs

	if r.Image != nil {
		r.Channel.Image = &ResourceRef{Resource: r.Image.About}
	}
	if r.TextInput != nil {
		r.Channel.TextInput = &ResourceRef{Resource: r.TextInput.About}
	}
}

// Validate checks the RDF-specific consistency rules the struct shape alone can't enforce: rdf:about uniqueness across
// the whole document, that channel's resource pointers match an actual top-level element, and that the channel's
// <items> sequence exactly matches the set of top-level <item> elements (per the spec's warning that mismatched items
// "are likely to be discarded by RDF parsers").
func (r RDF) Validate() error {
	if r.Channel.About == "" {
		return errors.New("rdf: channel rdf:about is required")
	}
	if len(r.Items) == 0 {
		return errors.New("rdf: at least one item is required")
	}

	seen := map[string]string{}
	check := func(about, label string) error {
		if about == "" {
			return fmt.Errorf("rdf: %s is missing rdf:about", label)
		}
		if other, ok := seen[about]; ok {
			return fmt.Errorf("rdf: rdf:about %q used by both %s and %s (must be unique)", about, other, label)
		}
		seen[about] = label
		return nil
	}

	if err := check(r.Channel.About, "channel"); err != nil {
		return err
	}
	for i, it := range r.Items {
		if err := check(it.About, fmt.Sprintf("item[%d]", i)); err != nil {
			return err
		}
	}

	if r.Image != nil {
		if err := check(r.Image.About, "image"); err != nil {
			return err
		}
		if r.Channel.Image == nil || r.Channel.Image.Resource != r.Image.About {
			return fmt.Errorf(
				"rdf: channel's <image rdf:resource> must match the top-level <image rdf:about> (%q)",
				r.Image.About,
			)
		}
	} else if r.Channel.Image != nil {
		return errors.New("rdf: channel references an <image>, but no top-level <image> element is present")
	}

	if r.TextInput != nil {
		if err := check(r.TextInput.About, "textinput"); err != nil {
			return err
		}
		if r.Channel.TextInput == nil || r.Channel.TextInput.Resource != r.TextInput.About {
			return fmt.Errorf(
				"rdf: channel's <textinput rdf:resource> must match the top-level <textinput rdf:about> (%q)",
				r.TextInput.About,
			)
		}
	} else if r.Channel.TextInput != nil {
		return errors.New("rdf: channel references a <textinput>, but no top-level <textinput> element is present")
	}

	itemSet := make(map[string]bool, len(r.Items))
	for _, it := range r.Items {
		itemSet[it.About] = true
	}
	for _, ref := range r.Channel.Items {
		if !itemSet[ref] {
			return fmt.Errorf(
				"rdf: channel's <items> sequence references %q, which has no matching top-level <item rdf:about>",
				ref,
			)
		}
	}
	for _, it := range r.Items {
		if found := slices.Contains(r.Channel.Items, it.About); !found {
			return fmt.Errorf(
				"rdf: item %q is not listed in the channel's <items> sequence and will likely be discarded by RDF parsers",
				it.About,
			)
		}
	}
	return nil
}

// MarshalXML implements xml.Marshaler for marshaling ItemRefs into the correct structure.
func (r ItemRefs) MarshalXML(enc *xml.Encoder, start xml.StartElement) error {
	if len(r) == 0 {
		return nil
	}
	if err := enc.EncodeToken(start); err != nil {
		return fmt.Errorf("marshal itemrefs: %w", err)
	}
	seq := xml.StartElement{Name: xml.Name{Local: "rdf:Seq"}}
	if err := enc.EncodeToken(seq); err != nil {
		return fmt.Errorf("marshal itemrefs: %w", err)
	}
	for _, uri := range r {
		li := xml.StartElement{
			Name: xml.Name{Local: "rdf:li"},
			Attr: []xml.Attr{
				{Name: xml.Name{Local: "resource"}, Value: uri},
			}, // unprefixed per spec's abbreviated syntax
		}
		if err := enc.EncodeToken(li); err != nil {
			return fmt.Errorf("marshal itemrefs: %w", err)
		}
		if err := enc.EncodeToken(li.End()); err != nil {
			return fmt.Errorf("marshal itemrefs: %w", err)
		}
	}
	if err := enc.EncodeToken(seq.End()); err != nil {
		return fmt.Errorf("marshal itemrefs: %w", err)
	}
	if err := enc.EncodeToken(start.End()); err != nil {
		return fmt.Errorf("marshal itemrefs: %w", err)
	}
	return nil
}

// UnmarshalXML implements xml.Unmarshaler for unmarshaling ItemRefs into the correct structure.
func (r *ItemRefs) UnmarshalXML(dec *xml.Decoder, start xml.StartElement) error {
	*r = nil
	depth := 0
	for {
		tok, err := dec.Token()
		if err != nil {
			return fmt.Errorf("unmarshal itemrefs: %w", err)
		}
		switch t := tok.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "Seq":
				depth++
			case "li":
				for _, a := range t.Attr {
					if a.Name.Local == "resource" {
						*r = append(*r, a.Value)
					}
				}
			default:
				if err := dec.Skip(); err != nil {
					return fmt.Errorf("unmarshal itemrefs: %w", err)
				}
			}
		case xml.EndElement:
			if t.Name.Local == "Seq" {
				depth--
			} else if t.Name.Local == start.Name.Local && depth == 0 {
				return nil
			}
		}
	}
}
