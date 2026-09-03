// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: 	MIT

package types

import (
	"slices"
	"strings"
)

var (
	MimeTypesImage = []string{"image/avif", "image/gif", "image/jpeg", "image/png", "image/svg+xml", "image/webp"}
)

// GetTitle returns the title (if any) of the image.
func (i *Image) GetTitle() string {
	if i.Title != nil {
		return *i.Title
	}
	return ""
}

// GetURL returns the URL of the image.
func (i *Image) GetURL() string {
	return i.URL
}

// IsImage will return a boolean indicating whether the given mimetype represents an image.
func IsImage(mimetype string) bool {
	return slices.ContainsFunc(MimeTypesImage, func(v string) bool {
		return strings.Contains(mimetype, v)
	})
}
