// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: 	MIT

package opml

import "slices"

// NewSubscriptionOutline creates a new OPML feed outline object from the given options.
func NewSubscriptionOutline(text, url string, options ...OutlineOption) *Outline {
	outline := &Outline{
		Text:   text,
		XMLURL: url,
		Type:   "rss",
	}

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
		o.Title = title
	}
}

// WithDescription option sets description of the subscription.
func WithDescription(desc string) OutlineOption {
	return func(o *Outline) {
		o.Description = desc
	}
}

// WithHTMLURL option sets a URL for the canonical HTML location (usually the source website) of the subscription.
func WithHTMLURL(url string) OutlineOption {
	return func(o *Outline) {
		o.HTMLURL = url
	}
}

// WithLanguage option sets the language the subscription contains.
func WithLanguage(lang string) OutlineOption {
	return func(o *Outline) {
		o.Language = lang
	}
}

// WithVersion sets the subscription version.
func WithVersion(version OutlineVersion) OutlineOption {
	return func(o *Outline) {
		o.Version = version
	}
}
