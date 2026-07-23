// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: 	MIT

// Package types contains methods and objects that are shared across different Feed schemas/specifications.
package types

import (
	"slices"
)

var (
	// MimeTypesRSS contains canonical/standard mimetypes for RSS feeds.
	MimeTypesRSS = []string{"application/rss+xml", "application/rdf+xml"}
	// MimeTypesAtom contains canonical/standard mimetypes for Atom feeds.
	MimeTypesAtom = []string{"application/atom+xml"}
	// MimeTypesIndeterminate contains mimetypes that can be used for either RSS/Atom feeds and don't give any clues to
	// the actual type.
	MimeTypesIndeterminate = []string{"application/xml", "text/xml"}
	// MimeTypesJSONFeed contains canonical/standard mimetypes for JSONFeed feeds.
	MimeTypesJSONFeed = []string{"application/feed+json", "application/json"}
	// MimeTypesFeed is the concatenation of all feed mime types.
	MimeTypesFeed = slices.Concat(MimeTypesAtom, MimeTypesRSS, MimeTypesIndeterminate)
	// MimeTypesHTML contains canonical/standard mimetypes for HTML.
	MimeTypesHTML = []string{"text/html", "application/xhtml+xml"}
	// MimeTypesImage contains canonical/standard/common mimetypes for images.
	MimeTypesImage = []string{"image/avif", "image/gif", "image/jpeg", "image/png", "image/svg+xml", "image/webp"}
	// MediaImageExt contains canonical/standard/common file extensions for images.
	MediaImageExt = []string{"jpg", "jpeg", "png", "webp", "gif"}
)

const (
	// MimeTypeOPML indicates the canonical mimetype for an OPML file.
	MimeTypeOPML = "text/x-opml+xml"
)
