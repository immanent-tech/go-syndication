// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: 	AGPL-3.0-or-later

// Package mrss contains objects and methods defining the MediaRSS extension.
package mrss

import (
	"github.com/immanent-tech/go-syndication/sanitization"
	"github.com/immanent-tech/go-syndication/types"
)

// AsImage returns the <media:thumbnail> object as a types.Image object.
func (t *MediaThumbnail) AsImage() *types.Image {
	return &types.Image{
		Value: t.URL,
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

// IsImage will return a boolean indicating whether the media:content element represents an image, and if it does, also
// return a generic types.Image object.
func (c *MediaContent) IsImage() (bool, *types.Image) {
	// Check if medium attr indicates an image.
	if c.Medium == MediaContentMediumImage {
		return true, &types.Image{
			Value: c.Url,
		}
	}
	// Check if mimetype attr indicates an image.
	if types.IsImage(c.Type) {
		return true, &types.Image{
			Value: c.Url,
		}
	}
	// Not an image.
	return false, nil
}
