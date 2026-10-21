// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: 	MIT

package jsonfeed

import (
	"slices"
	"strings"
	"time"

	"github.com/immanent-tech/go-syndication/sanitization"
	"github.com/immanent-tech/go-syndication/types"
)

var _ types.ItemSource = (*Item)(nil)

// GetID returns an "id" for the Item.
func (i *Item) GetID() string {
	return i.ID
}

// GetTitle retrieves title (if any) of the Item.
func (i *Item) GetTitle() string {
	if i.Title != nil {
		return sanitization.SanitizeText(*i.Title, sanitization.DefaultMaxRunes)
	}
	return ""
}

// GetLink retrieves URL associated with the item (if any). It will retrieve either the URL or ExternalURL, whichever is
// found first, or an empty string if neither is found.
func (i *Item) GetLink() string {
	var rawURL string
	switch {
	case i.URL != nil:
		rawURL = *i.URL
	case i.ExternalURL != nil:
		rawURL = *i.ExternalURL
	default:
		return ""
	}
	safeURL, err := sanitization.SanitizeURLString(rawURL)
	if err != nil {
		return ""
	}
	return safeURL
}

// GetDescription retrieves the summary (if any) of the Item.
func (i *Item) GetDescription() string {
	if i.Summary != nil {
		return sanitization.SanitizeText(*i.Summary, 600)
	}
	return ""
}

// GetAuthors retrieves the authors (if any) of the Item.
func (i *Item) GetAuthors() []string {
	authors := make([]string, 0, len(i.Authors))
	if i.Author != nil {
		authors = append(authors, i.Author.String())
	}
	for author := range slices.Values(i.Authors) {
		if author.String() != "" {
			authors = append(authors, author.String())
		}
	}
	return authors
}

// GetContributors is a no-op for JSONFeed items.
func (i *Item) GetContributors() []string {
	return nil
}

// GetRights is a no-op for JSONFeed items.
func (i *Item) GetRights() *string {
	return nil
}

// GetLanguage retrieves the language (if any) of the item.
func (i *Item) GetLanguage() *string {
	if i.Language != nil {
		lang := sanitization.SanitizeText(*i.Language, sanitization.DefaultMaxRunes)
		return &lang
	}
	return nil
}

// GetCategories retrieves the categories (if any) of the Item.
func (i *Item) GetCategories() []string {
	slices.Sort(i.Tags)
	i.Tags = slices.DeleteFunc(i.Tags, func(str string) bool {
		return strings.TrimSpace(str) == ""
	})
	for idx, tag := range i.Tags {
		i.Tags[idx] = sanitization.SanitizeText(tag, sanitization.DefaultMaxRunes)
	}
	return slices.CompactFunc(i.Tags, func(a, b string) bool {
		return strings.EqualFold(a, b)
	})
}

// GetImage retrieves the image (if any) for the Item.
func (i *Item) GetImage() *types.Image {
	if i.Image != nil {
		safeURL, err := sanitization.SanitizeURLString(*i.Image)
		if err != nil {
			return nil
		}
		return &types.Image{
			URL: safeURL,
		}
	}
	return nil
}

// GetPublishedDate returns the published date of the Item.
func (i *Item) GetPublishedDate() *time.Time {
	if i.DatePublished != nil {
		ts, err := time.Parse(time.RFC3339, *i.DatePublished)
		if err != nil {
			return nil
		}
		return &ts
	}
	return nil
}

// GetUpdatedDate returns the updated (modified) date of the Item.
func (i *Item) GetUpdatedDate() *time.Time {
	if i.DateModified != nil {
		ts, err := time.Parse(time.RFC3339, *i.DateModified)
		if err != nil {
			return nil
		}
		return &ts
	}
	return nil
}

// GetContent returns the content of the Item (if any). This will be either the html or text content, whichever is found
// first.
func (i *Item) GetContent() *string {
	var content string
	switch {
	case i.ContentHTML != nil:
		content = sanitization.SanitizeHTML(*i.ContentHTML)
	case i.ContentText != nil:
		content = sanitization.SanitizeHTML(*i.ContentText)
	}
	if content != "" {
		return &content
	}
	return nil
}

func (i *Item) GetGeoInfo() *types.GeoInfo {
	return nil
}

// GetCommentRSS returns whichever spelling was present after decoding a feed preferring the correct one if a
// (spec-violating) feed somehow supplied both.
func (i *Item) GetCommentRSS() string {
	return ""
}
