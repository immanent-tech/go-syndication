// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: 	MIT

package itunes

import (
	"slices"

	"github.com/immanent-tech/go-syndication/sanitization"
	"github.com/immanent-tech/go-syndication/types"
)

func (c Category) String() string {
	return sanitization.SanitizeText(c.Text, sanitization.DefaultMaxRunes)
}

// GetCategories returns all iTunes categories associated with the object.
func (c Categories) GetCategories() []string {
	var categories []string
	main := sanitization.SanitizeText(c.Text, sanitization.DefaultMaxRunes)
	categories = append(categories, main)
	if len(c.Categories) > 0 {
		for subcategory := range slices.Values(c.Categories) {
			categories = append(categories, main+"|"+subcategory.String())
		}
	}
	return categories
}

func (i Image) AsImage() *types.Image {
	safeURL, err := sanitization.SanitizeURLString(i.Href)
	if err != nil {
		return nil
	}
	return &types.Image{URL: safeURL}
}
