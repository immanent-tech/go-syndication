// Copyright 2026 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: 	MIT

package extensions

// WellKnownNamespaces is a convenience registry of namespace URIs commonly seen in RSS feeds. It's just a lookup table
// that can be used to lookup commonly used namespaces. It does not reflect all known namespaces and can be overridden.
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

// NewNamespace builds a Namespace. NewNamespace("content") looks up the canonical URI from the well-known registry
// above. NewNamespace("foo", "http://example.com/foo") declares an arbitrary namespace not in the registry.
func NewNamespace(prefix string, uri ...string) Namespace {
	if len(uri) > 0 {
		return Namespace{Prefix: prefix, URI: uri[0]}
	}
	return Namespace{Prefix: prefix, URI: WellKnownNamespaces[prefix]}
}
