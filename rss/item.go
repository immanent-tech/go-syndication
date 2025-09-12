// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: 	AGPL-3.0-or-later

package rss

import (
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/immanent-tech/go-syndication/types"
)

var _ types.ItemSource = (*Item)(nil)

var ErrItemValidation = errors.New("item is invalid")

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
	return i.Link
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
	if i.Author != "" {
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
	case i.DCLanguage != "":
		return i.DCLanguage.String()
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
	switch {
	case i.Image != nil:
		return &types.ImageInfo{
			URL:   i.Image.Link,
			Title: &i.Image.Title,
		}
	case i.Enclosure != nil && types.IsImage(i.Enclosure.Type):
		return &types.ImageInfo{
			URL: i.Enclosure.URL,
		}
	case i.MediaContent != nil:
		isImage, image := i.MediaContent.IsImage()
		if isImage {
			return image
		}
		return nil
	case len(i.MediaThumbnails) > 0:
		// Use the first thumbnail found.
		return i.MediaThumbnails[0].AsImage()
	default:
		return nil
	}
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
	if i.ContentEncoded != nil {
		return i.ContentEncoded.Value.String()
	}
	return ""
}

// Validate applies custom validation to an item.
func (i *Item) Validate() error {
	// Either description or title must be set. Both cannot be empty.
	if i.Description.String() == "" && i.Title.String() == "" {
		return fmt.Errorf("%w: description or title is required", ErrItemValidation)
	}
	return nil
}
