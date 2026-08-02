// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: 	MIT

// Package types contains methods and objects that are shared across different Feed schemas/specifications.
package feeds

import (
	"slices"

	"github.com/immanent-tech/go-syndication/atom"
	"github.com/immanent-tech/go-syndication/jsonfeed"
	"github.com/immanent-tech/go-syndication/rss"
)

var (
	// MimeTypesIndeterminate contains mimetypes that can be used for either RSS/Atom feeds and don't give any clues to
	// the actual type.
	MimeTypesIndeterminate = []string{"application/xml", "text/xml"}
	// MimeTypesFeed is the concatenation of all feed mime types.
	MimeTypesFeed = slices.Concat(atom.MimeTypes, rss.MimeTypes, jsonfeed.MimeTypes, MimeTypesIndeterminate)
	// MimeTypesHTML contains canonical/standard mimetypes for HTML.
	MimeTypesHTML = []string{"text/html", "application/xhtml+xml"}
	// MimeTypesImage contains canonical/standard/common mimetypes for images.
	MimeTypesImage = []string{"image/avif", "image/gif", "image/jpeg", "image/png", "image/svg+xml", "image/webp"}
)
