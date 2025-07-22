// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: 	AGPL-3.0-or-later

// Package mrss contains objects and methods defining the MediaRSS extension.
package mrss

import (
	"github.com/joshuar/go-syndication/sanitization"
	"github.com/joshuar/go-syndication/types"
)

// AsImage returns the <media:thumbnail> object as a types.Image object.
func (t *MediaThumbnail) AsImage() *types.Image {
	return &types.Image{
		Value: t.URL,
	}
}

func (c *MediaContent) GetCategory() string {
	if c.MediaCategory != nil {
		if c.MediaCategory.Label != nil {
			return *c.MediaCategory.Label
		}
		return sanitization.SanitizeString(c.MediaCategory.Value)
	}
	return ""
}

func (t *MediaText) GetText() string {
	return sanitization.SanitizeString(t.Value)
}
