// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: 	AGPL-3.0-or-later

package itunes

import (
	"slices"

	"github.com/immanent-tech/go-syndication/sanitization"
)

func (c *Category) String() string {
	return sanitization.SanitizeString(c.Text)
}

// GetCategories returns all iTunes categories associated with the object.
func (c *Categories) GetCategories() []string {
	var categories []string
	main := sanitization.SanitizeString(c.Text)
	categories = append(categories, main)
	if len(c.Categories) > 0 {
		for subcategory := range slices.Values(c.Categories) {
			categories = append(categories, main+"|"+subcategory.String())
		}
	}
	return categories
}
