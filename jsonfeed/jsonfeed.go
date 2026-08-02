// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: 	MIT

// Package jsonfeed contains objects and methods defining the JSONFeed syndication format.
package jsonfeed

var (
	// MimeTypes contains canonical/standard mimetypes for JSONFeed feeds.
	MimeTypes = []string{"application/feed+json", "application/json"}
)

func (a Author) String() string {
	if a.Name != nil {
		return *a.Name
	}
	return ""
}
