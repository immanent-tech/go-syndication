// Copyright 2026 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: 	MIT

package extensions

// WellKnownNamespaces is a convenience registry of namespace URIs commonly
// seen in RSS feeds. It's just a lookup table for NS(prefix) below --
// declaring a namespace is never limited to entries in this list.
var WellKnownNamespaces = map[string]string{
	"content": "http://purl.org/rss/1.0/modules/content/",
	"media":   "http://search.yahoo.com/mrss/",
	"atom":    "http://www.w3.org/2005/Atom",
	"dc":      "http://purl.org/dc/elements/1.1/",
	"slash":   "http://purl.org/rss/1.0/modules/slash/",
	"syn":     "http://purl.org/rss/1.0/modules/syndication/",
	"itunes":  "http://www.itunes.com/dtds/podcast-1.0.dtd",
	"georss":  "http://www.georss.org/georss",
	"wfw":     "http://wellformedweb.org/CommentAPI/",
}


