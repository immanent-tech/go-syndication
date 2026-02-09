// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: 	MIT

// Package media contains objects and methods defining the MediaRSS extension.
package media

import (
	"net/url"
	"slices"
	"strings"

	"github.com/immanent-tech/go-syndication/sanitization"
	"github.com/immanent-tech/go-syndication/types"
)

// AsImage returns the <media:thumbnail> object as a types.ImageInfo object.
func (t *MediaThumbnail) AsImage() *types.ImageInfo {
	return &types.ImageInfo{
		URL: t.URL,
	}
}

// GetCategory retrieves the category assigned to the media:content element (if any).
func (c *MediaContent) GetCategory() string {
	if c.MediaCategory != nil {
		if c.MediaCategory.Label != "" {
			return c.MediaCategory.Label
		}
		return sanitization.SanitizeString(c.MediaCategory.Value)
	}
	return ""
}

// GetText retrieves the text of media:content element (if any).
func (t *MediaText) GetText() string {
	return sanitization.SanitizeString(t.Value)
}

// AsImage will return a types.ImageInfo if the <media:content> element represents an image. If not, it will return nil.
func (c *MediaContent) AsImage() *types.ImageInfo {
	// Check if medium attr indicates an image.
	if c.Medium == MediaContentMediumImage {
		return &types.ImageInfo{
			URL: c.Url,
		}
	}
	// Check if mimetype attr indicates an image.
	if types.IsImage(c.Type) {
		return &types.ImageInfo{
			URL: c.Url,
		}
	}
	// Ugh, maybe try parsing the URL and see if it ends in a well-known image file extension...
	if url, err := url.Parse(c.Url); err == nil {
		for imgext := range slices.Values(types.MediaImageExt) {
			if strings.HasSuffix(url.Path, imgext) {
				return &types.ImageInfo{
					URL: c.Url,
				}
			}
		}
	}

	return nil
}
