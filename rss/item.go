// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: 	MIT

//nolint:sloglint // ignore bare slog usage in pkg.
package rss

import (
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"strings"
	"time"

	"github.com/immanent-tech/go-syndication/extensions/rss"
	"github.com/immanent-tech/go-syndication/types"
	"golang.org/x/net/html"
)

var _ types.ItemSource = (*Item)(nil)

var ErrItemValidation = errors.New("item is invalid")

// NewItem creates a new Item with the given options.
func NewItem(options ...ItemOption) *Item {
	item := &Item{
		PubDate: &types.DateTime{Time: time.Now().UTC()},
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
		i.Title = types.String(title)
	}
}

// WithItemDescription option sets the item description. Note to be value, an item needs either the title or description
// set.
func WithItemDescription(desc string) ItemOption {
	return func(i *Item) {
		i.Description = types.String(desc)
	}
}

// WithItemLink option sets the URL to the original page displaying the item.
func WithItemLink(link string) ItemOption {
	return func(i *Item) {
		i.Link = types.String(link)
	}
}

// WithItemGUID option assigns the given GUID to the item.
func WithItemGUID(guid *GUID) ItemOption {
	return func(i *Item) {
		i.GUID = guid
	}
}

// WithItemContent option sets the item content.
func WithItemContent(content []byte) ItemOption {
	return func(i *Item) {
		i.ContentEncoded = (*rss.ContentEncoded)(&content)
	}
}

// WithItemPublishedDate option sets the published date of the item.
func WithItemPublishedDate(ts time.Time) ItemOption {
	return func(i *Item) {
		if ts.IsZero() {
			// Ignore zero value.
			return
		}
		i.PubDate.Time = ts
	}
}

// GetID returns an "id" for the item. This will be the value of the <guid> element, if present, or an empty string if
// not present.
func (i *Item) GetID() string {
	return i.GUID.Value.String()
}

// GetTitle retrieves the <title> (if any) of the Item.
func (i *Item) GetTitle() string {
	switch {
	case i.DCTitle != nil:
		return i.DCTitle.String()
	default:
		return i.Title.String()
	}
}

// GetLink retrieves the <link> (if any) of the Item.
func (i *Item) GetLink() string {
	return i.Link.String()
}

// GetDescription retrieves the <description> (if any) of the Item.
func (i *Item) GetDescription() string {
	switch {
	case i.DCDescription != nil:
		return i.DCDescription.String()
	default:
		return i.Description.String()
		// return sanitization.SanitizeString(i.Description)
	}
}

// GetAuthors retrieves the authors (if any) of the Item. This will be the list of values from any <author> and
// <dc:creator> elements.
func (i *Item) GetAuthors() []string {
	var authors []string
	if i.Author != nil && i.Author.String() != "" {
		authors = append(authors, i.Author.String())
	}
	if i.DCCreator != nil {
		authors = append(authors, i.DCCreator.String())
	}
	return authors
}

// GetContributors retrieves the contributors (if any) of the Item. This will be the list of values from the
// <dc:contributor> element.
func (i *Item) GetContributors() []string {
	var contributors []string
	if i.DCContributor != nil {
		contributors = append(contributors, i.DCContributor.String())
	}
	return contributors
}

// GetRights retrieves the rights (copyright) of the Channel. This will be the value of <dc:rights>, if found.
func (i *Item) GetRights() string {
	if i.DCRights != nil {
		return i.DCRights.String()
	}
	return ""
}

// GetLanguage retrieves the language of the Item. This will be the value found from the <dc:language> element, if
// present.
func (i *Item) GetLanguage() string {
	switch {
	case i.DCLanguage != nil:
		return *i.DCLanguage
	default:
		return ""
	}
}

// GetCategories retrieves the categories (if any) of the Item. The categories are returned as strings.
func (i *Item) GetCategories() []string {
	categories := make([]string, 0, len(i.Categories))
	for category := range slices.Values(i.Categories) {
		categories = append(categories, category.String())
	}
	return categories
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

// GetPublishedDate returns the <pubDate> of the Item (if any). If there is no publish date, it will return a
// DateTime equal to Unix epoch.
func (i *Item) GetPublishedDate() time.Time {
	if i.PubDate != nil {
		return i.PubDate.Time
	}
	return time.Unix(0, 0)
}

// GetUpdatedDate returns the <pubDate> of the Item (if any). If there is no publish date, it will return a
// DateTime equal to Unix epoch.
func (i *Item) GetUpdatedDate() time.Time {
	return i.GetPublishedDate()
}

// GetContent returns the content of the Item (if any). This will be taken from any <content:encoded> element.
func (i *Item) GetContent() string {
	if i.ContentEncoded == nil || i.ContentEncoded.String() == "" {
		return ""
	}
	// Parse the value.
	doc, err := html.Parse(strings.NewReader(i.ContentEncoded.String()))
	if err != nil {
		slog.Error("Unable to parse content:encoded.",
			slog.Any("error", err),
		)
		return ""
	}
	// Write out.
	var out strings.Builder
	err = html.Render(&out, doc)
	if err != nil {
		slog.Error("Unable to render content:encoded.",
			slog.Any("error", err),
		)
		return ""
	}
	return out.String()
}

// Validate applies custom validation to an item.
func (i *Item) Validate() error {
	// Either description or title must be set. Both cannot be empty.
	if i.Description.String() == "" && i.Title.String() == "" {
		return fmt.Errorf("%w: description or title is required", ErrItemValidation)
	}
	return nil
}

// GenerateGUID creates a GUID from the given value, with the given permalink status.
func GenerateGUID(value string, permalink bool) *GUID {
	return &GUID{
		IsPermaLink: permalink,
		Value:       types.String(value),
	}
}
