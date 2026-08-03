// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: 	MIT

package opml

import "slices"

// NewSubscriptionOutline creates a new OPML feed outline object from the given options.
func NewSubscriptionOutline(text, url string, options ...OutlineOption) *Outline {
	outline := &Outline{
		Text: text,
		Type: new("rss"),
	}
	outline.SetAttr("xmlUrl", url)

	for option := range slices.Values(options) {
		option(outline)
	}

	return outline
}

// OutlineOption is a functional option to apply to an outline.
type OutlineOption func(*Outline)

// WithOutlineTitle option sets the title of the subscription.
func WithOutlineTitle(title string) OutlineOption {
	return func(o *Outline) {
		o.SetAttr("title", title)
	}
}

// WithDescription option sets description of the subscription.
func WithDescription(desc string) OutlineOption {
	return func(o *Outline) {
		o.SetAttr("description", desc)
	}
}

// WithHTMLURL option sets a URL for the canonical HTML location (usually the source website) of the subscription.
func WithHTMLURL(url string) OutlineOption {
	return func(o *Outline) {
		o.SetAttr("htmlUrl", url)
	}
}

// WithLanguage option sets the language the subscription contains.
func WithLanguage(lang string) OutlineOption {
	return func(o *Outline) {
		o.SetAttr("language", lang)
	}
}

// WithVersion sets the subscription version.
func WithVersion(version RSSOutlineVersion) OutlineOption {
	return func(o *Outline) {
		o.SetAttr("version", string(version))
	}
}
