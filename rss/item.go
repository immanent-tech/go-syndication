// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: 	MIT

//nolint:sloglint // ignore bare slog usage in pkg.
package rss

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/url"
	"slices"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/immanent-tech/go-syndication/extensions/content"
	"github.com/immanent-tech/go-syndication/extensions/media"
	"github.com/immanent-tech/go-syndication/sanitization"
	"github.com/immanent-tech/go-syndication/types"
	"github.com/immanent-tech/go-syndication/validation"
)

var _ types.ItemSource = (*Item)(nil)

// NewItem creates a new Item with the given options.
func NewItem(options ...ItemOption) *Item {
	item := &Item{
		PubDate: NewTimestamp(time.Now().UTC()),
	}

	for option := range slices.Values(options) {
		option(item)
	}

	return item
}

// ItemOption is a functional option applied to an Item.
type ItemOption func(*Item)

// WithItemTitle option sets the item title. Note to be value, an item needs either the title or description set.
func WithItemTitle(title string) ItemOption {
	return func(i *Item) {
		i.Title = title
	}
}

// WithItemDescription option sets the item description. Note to be value, an item needs either the title or description
// set.
func WithItemDescription(desc string, cdata bool) ItemOption {
	return func(i *Item) {
		i.Description = NewItemDescription(desc, cdata)
	}
}

// WithItemLink option sets the URL to the original page displaying the item.
func WithItemLink(link string) ItemOption {
	return func(i *Item) {
		i.Link = link
	}
}

// WithItemGUID option assigns the given GUID to the item.
func WithItemGUID(guid *GUID) ItemOption {
	return func(i *Item) {
		i.GUID = guid
	}
}

// WithItemImage option sets the item image.
func WithItemImage(img *types.Image) ItemOption {
	return func(i *Item) {
		i.MediaThumbnails = media.MediaThumbnails{
			media.MediaThumbnail{
				URL: img.GetURL(),
			},
		}
	}
}

// WithItemContent option sets the item content.
func WithItemContent(itemContent string, cdata bool) ItemOption {
	return func(i *Item) {
		i.ContentEncoded = new(content.NewContentEncoded(itemContent, cdata))
	}
}

// WithItemPublishedDate option sets the published date of the item.
func WithItemPublishedDate(ts time.Time) ItemOption {
	return func(i *Item) {
		if ts.IsZero() {
			// Ignore zero value.
			return
		}
		i.PubDate = NewTimestamp(ts)
	}
}

// GetID returns an "id" for the item. This will be the value of the <guid> element, if present, or an empty string if
// not present.
func (i Item) GetID() string {
	if i.GUID != nil {
		return i.GUID.String()
	}
	return ""
}

// GetTitle retrieves the <title> (if any) of the Item.
func (i Item) GetTitle() string {
	return sanitization.SanitizeText(sanitization.SanitizeHTML(i.Title), sanitization.DefaultMaxRunes)
}

// GetLink retrieves the <link> (if any) of the Item.
func (i Item) GetLink() string {
	safe, err := sanitization.SanitizeURLString(i.Link)
	if err != nil {
		return ""
	}
	return safe
}

// GetDescription retrieves the <description> (if any) of the Item.
func (i Item) GetDescription() string {
	// Use the nonempty description.
	if i.Description.String() != "" {
		return i.Description.String()
	}
	// Else, use a description from one of these:
	switch {
	case i.MediaGroup != nil:
		return i.MediaGroup.GetDescription()
	default:
		return ""
	}
}

// GetAuthors retrieves the authors (if any) of the Item. This will be the list of values from any <author> and
// <dc:creator> elements.
func (i Item) GetAuthors() []string {
	authors := make([]string, 0)
	if i.Author != nil && *i.Author != "" {
		authors = append(authors, sanitization.SanitizeText(*i.Author, sanitization.DefaultMaxRunes))
	}
	if len(i.DcPublisher) > 0 {
		authors = append(authors, i.DcPublisher...)
	}
	if len(i.DcCreator) > 0 {
		authors = append(authors, i.DcCreator...)
	}
	slices.Sort(authors)
	return slices.Compact(authors)
}

// GetContributors retrieves the contributors (if any) of the Item. This will be the list of values from the
// <dc:contributor> element.
func (i Item) GetContributors() []string {
	var contributors []string
	if len(i.DcContributor) > 0 {
		contributors = append(contributors, i.DcContributor...)
	}
	return contributors
}

// GetRights retrieves the rights (copyright) of the Channel. This will be the value of <dc:rights>, if found.
func (i Item) GetRights() *string {
	if rights := i.DcRights.String(); rights != "" {
		return &rights
	}
	return nil
}

// GetLanguage retrieves the language of the Item. This will be the value found from the <dc:language> element, if
// present.
func (i Item) GetLanguage() *string {
	if lang := i.DcLanguage.String(); lang != "" {
		return &lang
	}
	return nil
}

// GetCategories retrieves the categories (if any) of the [Item]. Case insensitive duplicate categories are collapsed.
// The categories are returned as strings.
func (i Item) GetCategories() []string {
	categories := make([]string, 0, len(i.Categories))
	for category := range slices.Values(i.Categories) {
		if str := sanitization.SanitizeText(category.String(), sanitization.DefaultMaxRunes); str != "" {
			categories = append(categories, str)
		}
	}
	// Treat <dc:subject> values as categories.
	if len(i.DcSubject) > 0 {
		categories = append(categories, i.DcSubject...)
	}
	// Treat <slash:department> as a category.
	if i.SlashDepartment != nil {
		categories = append(categories, sanitization.SanitizeText(*i.SlashDepartment, sanitization.DefaultMaxRunes))
	}
	// Treat <slash:section> as a category.
	if i.SlashSection != nil {
		categories = append(categories, sanitization.SanitizeText(*i.SlashSection, sanitization.DefaultMaxRunes))
	}
	slices.Sort(categories)
	return slices.CompactFunc(categories, func(a, b string) bool {
		return strings.EqualFold(a, b)
	})
}

// GetImage retrieves the image (if any) for the Item. The image is returned as a types.ImageInfo object. There are many
// places/elements that could represent the item's image, or rather, many ways various feeds indicate an image:
//
// - an <image> element in the item.
//
// - an <enclosure> element in the item with a mimetype that is an image.
//
// - a <media:content> element with medium=image or mimetype indicating an image.
//
// - a single <media:thumbnail> element.
//
// This method tries to retrieve one of these, first one wins, in the order above.
func (i Item) GetImage() *types.Image {
	var img *types.Image
	switch {
	case i.Image != nil:
		// Item has an <image> element, use it.
		img = i.Image.AsImage()
	case i.Enclosure != nil && types.IsImage(i.Enclosure.Type):
		// Item has an <enclosure> element, check if it contains an image and use it.
		safeURL, err := sanitization.SanitizeURLString(i.Enclosure.URL)
		if err == nil {
			img = &types.Image{URL: safeURL}
		}
	case i.MediaContent != nil && i.MediaContent.AsImage() != nil:
		// Item has a <media:content> element, extract the image.
		img = i.MediaContent.AsImage()
	case len(i.MediaThumbnails) > 0:
		// Check for a <media:thumbnails> element and assume the first element is an appropriate image.
		img = i.MediaThumbnails[0].AsImage()
	default:
		return nil
	}
	// If the image does not have a title, set it to the item title.
	if img != nil {
		if img.Title == nil {
			img.Title = new(i.GetTitle())
		}
	}
	if err := validation.ValidateStruct(img); err != nil {
		return nil
	}
	return img
}

// GetMediaGroup returns any media.MediaGroup object for the entry.
func (i Item) GetMediaGroup() *media.MediaGroup {
	return i.MediaGroup
}

// GetPublishedDate returns the <pubDate> of the Item (if any). If there is no publish date, it will return a
// DateTime equal to Unix epoch.
func (i Item) GetPublishedDate() *time.Time {
	if i.PubDate != nil {
		return &i.PubDate.Value
	}
	if len(i.DcDate) == 1 {
		dcDate := i.DcDate[0]
		if !dcDate.Value.IsZero() {
			return new(dcDate.Value)
		}
	}
	return nil
}

// GetUpdatedDate returns the <pubDate> of the Item (if any). If there is no publish date, it will return a
// DateTime equal to Unix epoch.
func (i Item) GetUpdatedDate() *time.Time {
	return nil
}

// GetContent returns the content of the Item (if any). This will be taken from any <content:encoded> element.
func (i Item) GetContent() *string {
	if i.ContentEncoded == nil || i.ContentEncoded.String() == "" {
		return nil
	}
	content := i.ContentEncoded.String()
	return &content
}

// HasBasicGeo checks whether an item has fields from the basic_geo extension.
func (i Item) HasBasicGeo() bool {
	return i.Alt != nil || i.Lat != nil || i.Lon != nil
}

// HasGeoRSS checks whether an item has fields from the georss extension.
func (i Item) HasGeoRSS() bool {
	return i.Point != nil || i.Line != nil || i.Polygon != nil || i.Box != nil || i.Circle != nil || i.Elev != nil ||
		i.Floor != nil ||
		i.Radius != nil ||
		i.FeatureName != nil ||
		i.FeatureTypeTag != nil ||
		i.RelationshipTag != nil ||
		i.Where != nil
}

func (i Item) GetGeoInfo() *types.GeoInfo {
	info := types.GeoInfo{}
	data, err := json.Marshal(i)
	if err != nil {
		return nil
	}
	err = json.Unmarshal(data, &info)
	if err != nil {
		return nil
	}
	if !info.IsZero() {
		return &info
	}
	return nil
}

// GetCommentRSS returns whichever spelling was present after decoding a feed preferring the correct one if a
// (spec-violating) feed somehow supplied both.
func (i Item) GetCommentRSS() string {
	switch {
	case i.CommentRSS != nil:
		return *i.CommentRSS
	case i.CommentRSSLegacy != nil:
		return *i.CommentRSSLegacy
	default:
		return ""
	}
}

func itemCustomValidation(sl validator.StructLevel) {
	i := sl.Current().Interface().(Item)
	if i.Description.String() == "" && i.Title == "" {
		sl.ReportError(i, "Item", "Item", "required", "descrition or title is required")
	}
}

func NewItemDescription(value string, cdata bool) ItemDescription {
	return ItemDescription{
		Value: value,
		CDATA: cdata,
	}
}

func (c ItemDescription) String() string {
	return sanitization.SanitizeHTML(c.Value)
}

// MarshalXML implements xml.Marshaler.
func (c ItemDescription) MarshalXML(enc *xml.Encoder, start xml.StartElement) error {
	// Force the literal element name "content:encoded". Go's xml package doesn't manage namespace prefixes well on
	// marshal, so the common workaround is to declare xmlns:content on the root element.
	start.Name = xml.Name{Local: "description"}

	if c.CDATA {
		if err := enc.EncodeElement(struct {
			Value string `xml:",cdata"`
		}{c.Value}, start); err != nil {
			return fmt.Errorf("encode description: %w", err)
		}
	}
	if err := enc.EncodeElement(struct {
		Value string `xml:",chardata"`
	}{c.Value}, start); err != nil {
		return fmt.Errorf("encode description: %w", err)
	}

	return nil
}

// UnmarshalXML implements xml.Unmarshaler.
//
// Note: Go's decoder does not distinguish a CDATA section from ordinary character data at the token level -- both come
// back as CharData and get concatenated into a plain ",chardata" field. That means this single implementation correctly
// reads content:encoded whether the source feed used CDATA-escaping or entity-encoding, per the spec's "entity-encoded
// or CDATA-escaped" wording. We can't reliably recover which form was originally used, so CDATA is left at its zero
// value (false) after decoding; set it yourself before re-marshaling if it matters.
func (c *ItemDescription) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var valueStruct struct {
		Value string `xml:",chardata"`
	}
	if err := d.DecodeElement(&valueStruct, &start); err != nil {
		return fmt.Errorf("decode item description: %w", err)
	}
	c.Value = valueStruct.Value
	return nil
}

func (c ItemDescription) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(c.Value)
	if err != nil {
		return nil, fmt.Errorf("marshal item description: %w", err)
	}
	return data, nil
}

func (c *ItemDescription) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("unmarshal item description: %w", err)
	}
	c.Value = s
	return nil
}

// NewGUID creates a GUID from the given value, with the given permalink status.
func NewGUID(value string, permalink bool) *GUID {
	return &GUID{
		IsPermaLink: permalink,
		Value:       value,
	}
}

func guidCustomValidation(sl validator.StructLevel) {
	i := sl.Current().Interface().(GUID)
	if i.IsPermaLink {
		u, err := url.Parse(i.Value)
		if err != nil {
			sl.ReportError(i.Value, "Value", "Value", "url", fmt.Sprintf("parse url: %s", err.Error()))
			return
		}
		if !u.IsAbs() {
			sl.ReportError(i.Value, "Value", "Value", "url", "url is not absolute")
		}
	}
}
