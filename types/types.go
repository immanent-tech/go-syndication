// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: 	AGPL-3.0-or-later

// Package types contains methods and objects that are shared across different Feed schemas/specifications.
package types

import (
	"encoding/json"
	"fmt"
	"html"
	"slices"

	"encoding/xml"

	"github.com/immanent-tech/go-syndication/sanitization"
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
	MimeTypesFeed = slices.Concat(MimeTypesAtom, MimeTypesRSS, MimeTypesIndeterminate, MimeTypesJSONFeed)
	// MimeTypesHTML contains canonical/standard mimetypes for HTML.
	MimeTypesHTML = []string{"text/html", "application/xhtml+xml"}
	// MimeTypesImage contains canonical/standard/common mimetypes for images.
	MimeTypesImage = []string{"image/avif", "image/gif", "image/jpeg", "image/png", "image/svg+xml", "image/webp"}
)

// String will return the value of the object.
func (c *CustomTypeBase) String() string {
	if c != nil {
		return c.Value
	}
	return ""
}

// NewXMLAttr is a convienience function to create an xml.Attr from a name/value/namespace combination. The namespace
// value is optional, but the name and value should be provided.
func NewXMLAttr(name, value, namespace string) xml.Attr {
	return xml.Attr{
		Name: xml.Name{
			Space: namespace,
			Local: name,
		},
		Value: value,
	}
}

func (c *CharData) UnmarshalText(data []byte) error {
	c.Value = sanitization.SanitizeBytes(data)
	return nil
}

func (c *CharData) UnmarshalJSON(data []byte) error {
	var chardata struct {
		CharData []byte `json:"CharData"`
	}

	err := json.Unmarshal(data, &chardata)
	if err != nil {
		return fmt.Errorf("cannot unmarshal chardata: %w", err)
	}

	c.Value = sanitization.SanitizeBytes(chardata.CharData)

	return nil
}

func (c *CharData) String() string {
	return html.UnescapeString(string(c.Value))
}

type StringData string

func (s *StringData) UnmarshalText(data []byte) error {
	safeData := sanitization.SanitizeBytes(data)
	*s = StringData(string(safeData))
	return nil
}

func (s *StringData) String() string {
	return string(*s)
}
