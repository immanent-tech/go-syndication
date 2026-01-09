// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: 	MIT

package types

import (
	"slices"
	"strings"
)

// GetTitle returns the title (if any) of the image.
func (i *ImageInfo) GetTitle() string {
	return i.Title
}

// GetURL returns the URL of the image.
func (i *ImageInfo) GetURL() string {
	return i.URL
}

// IsImage will return a boolean indicating whether the given mimetype represents an image.
func IsImage(mimetype string) bool {
	return slices.ContainsFunc(MimeTypesImage, func(v string) bool {
		return strings.Contains(mimetype, v)
	})
}
