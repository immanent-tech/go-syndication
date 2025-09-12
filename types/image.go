// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: 	AGPL-3.0-or-later

package types

import (
	"slices"
	"strings"
)

// GetTitle returns the title (if any) of the image.
func (i *ImageInfo) GetTitle() string {
	if i != nil {
		if i.Title != nil && *i.Title != "" {
			return *i.Title
		}
	}
	return ""
}

// GetURL returns the URL of the image.
func (i *ImageInfo) GetURL() string {
	if i != nil {
		return i.URL
	}
	return ""
}

// IsImage will return a boolean indicating whether the given mimetype represents an image.
func IsImage(mimetype string) bool {
	return slices.ContainsFunc(MimeTypesImage, func(v string) bool {
		return strings.Contains(mimetype, v)
	})
}
