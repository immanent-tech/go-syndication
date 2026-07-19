// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: 	MIT

//nolint:sloglint // ignore bare slog usage in pkg.
package rss

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"log/slog"
	"slices"
	"strings"
	"time"

	"github.com/immanent-tech/go-syndication/extensions/media"
	"github.com/immanent-tech/go-syndication/extensions/rss"
	"github.com/immanent-tech/go-syndication/types"
	"github.com/immanent-tech/go-syndication/validation"
	"golang.org/x/net/html"
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
func WithItemImage(img *types.ImageInfo) ItemOption {
	return func(i *Item) {
		i.MediaThumbnails = media.MediaThumbnails{
			media.MediaThumbnail{
				XMLName: xml.Name{Local: "media:thumbnail"},
				URL:     img.GetURL(),
			},
		}
	}
}

// WithItemContent option sets the item content.
func WithItemContent(content string, cdata bool) ItemOption {
	return func(i *Item) {
		i.ContentEncoded = new(rss.NewContentEncoded(content, cdata))
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
func (i *Item) GetID() string {
	if i.GUID != nil {
		return i.GUID.Value
	}
	return ""
}

// GetTitle retrieves the <title> (if any) of the Item.
func (i *Item) GetTitle() string {
	return i.Title
}

// GetLink retrieves the <link> (if any) of the Item.
func (i *Item) GetLink() string {
	return i.Link
}

// GetDescription retrieves the <description> (if any) of the Item.
func (i *Item) GetDescription() string {
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
func (i *Item) GetAuthors() []string {
	var authors []string
	if i.Author != nil && *i.Author != "" {
		authors = append(authors, *i.Author)
	}
	if i.Creator != nil {
		authors = append(authors, *i.Creator...)
	}
	return authors
}

// GetContributors retrieves the contributors (if any) of the Item. This will be the list of values from the
// <dc:contributor> element.
func (i *Item) GetContributors() []string {
	var contributors []string
	if i.Contributor != nil {
		contributors = append(contributors, *i.Contributor...)
	}
	return contributors
}

// GetRights retrieves the rights (copyright) of the Channel. This will be the value of <dc:rights>, if found.
func (i *Item) GetRights() *string {
	if i.Rights != nil {
		return new(strings.Join(*i.Rights, " "))
	}
	return nil
}

// GetLanguage retrieves the language of the Item. This will be the value found from the <dc:language> element, if
// present.
func (i *Item) GetLanguage() *string {
	switch {
	case i.Language != nil:
		return new(strings.Join(*i.Language, " "))
	default:
		return nil
	}
}

// GetCategories retrieves the categories (if any) of the Item. The categories are returned as strings.
func (i *Item) GetCategories() []string {
	categories := make([]string, 0, len(i.Categories))
	for category := range slices.Values(i.Categories) {
		categories = append(categories, category.String())
	}
	slices.Sort(categories)
	return slices.Compact(categories)
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
func (i *Item) GetImage() *types.ImageInfo {
	var img *types.ImageInfo
	switch {
	case i.Image != nil:
		// Item has an <image> element, use it.
		img = &types.ImageInfo{
			URL:   i.Image.URL,
			Title: i.Image.Title,
		}
	case i.Enclosure != nil && types.IsImage(i.Enclosure.Type):
		// Item has an <enclosure> element, check if it contains an image and use it.
		img = &types.ImageInfo{
			URL: i.Enclosure.URL,
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
		if img.Title == "" {
			img.Title = i.GetTitle()
		}
	}
	return img
}

// GetMediaGroup returns any media.MediaGroup object for the entry.
func (i *Item) GetMediaGroup() *media.MediaGroup {
	return i.MediaGroup
}

// GetPublishedDate returns the <pubDate> of the Item (if any). If there is no publish date, it will return a
// DateTime equal to Unix epoch.
func (i *Item) GetPublishedDate() *time.Time {
	if i.PubDate != nil {
		return &i.PubDate.Value
	}
	return nil
}

// GetUpdatedDate returns the <pubDate> of the Item (if any). If there is no publish date, it will return a
// DateTime equal to Unix epoch.
func (i *Item) GetUpdatedDate() *time.Time {
	return nil
}

// GetContent returns the content of the Item (if any). This will be taken from any <content:encoded> element.
func (i *Item) GetContent() *string {
	if i.ContentEncoded == nil || i.ContentEncoded.String() == "" {
		return nil
	}
	// Parse the value.
	doc, err := html.Parse(strings.NewReader(i.ContentEncoded.String()))
	if err != nil {
		slog.Error("Unable to parse content:encoded.",
			slog.Any("error", err),
		)
		return nil
	}
	// Write out.
	var out strings.Builder
	err = html.Render(&out, doc)
	if err != nil {
		slog.Error("Unable to render content:encoded.",
			slog.Any("error", err),
		)
		return nil
	}
	return new(out.String())
}

// Validate applies custom validation to an item.
func (i *Item) Validate() error {
	// Either description or title must be set. Both cannot be empty.
	if i.Description.String() == "" && i.Title == "" {
		return fmt.Errorf("%w: description or title is required", validation.ErrInvalidStruct)
	}
	return nil
}

// NewGUID creates a GUID from the given value, with the given permalink status.
func NewGUID(value string, permalink bool) *GUID {
	return &GUID{
		IsPermaLink: permalink,
		Value:       value,
	}
}

func NewItemDescription(value string, cdata bool) ItemDescription {
	return ItemDescription{
		Value: value,
		CDATA: cdata,
	}
}

func (c ItemDescription) String() string {
	return c.Value
}

// MarshalXML implements xml.Marshaler.
func (c ItemDescription) MarshalXML(enc *xml.Encoder, start xml.StartElement) error {
	// Force the literal element name "content:encoded". Go's xml package
	// doesn't manage namespace prefixes well on marshal, so the common
	// workaround is to declare xmlns:content on the root element.
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
// Note: Go's decoder does not distinguish a CDATA section from ordinary
// character data at the token level -- both come back as CharData and get
// concatenated into a plain ",chardata" field. That means this single
// implementation correctly reads content:encoded whether the source feed
// used CDATA-escaping or entity-encoding, per the spec's "entity-encoded or
// CDATA-escaped" wording. We can't reliably recover which form was
// originally used, so CDATA is left at its zero value (false) after
// decoding; set it yourself before re-marshaling if it matters.
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
