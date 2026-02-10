// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: 	MIT

// Package types contains methods and objects that are shared across different Feed schemas/specifications.
package types

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"html"
	"slices"
	"time"

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

var (
	// DefaultFeedUpdateInterval defines the update interval for feeds that do not define an update interval or where
	// one cannot be calculated based off item frequency.
	DefaultFeedUpdateInterval = time.Hour
)

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

// String is custom string type that handles "malformed" string fields containing whitespace or forbidden input.
//
//nolint:recvcheck // required for unmarshal to work correctly.
type String string

// UnmarshalText provides custom unmarshaling of String that will sanitize, unescape and trim whitespace from the value.
func (s *String) UnmarshalText(data []byte) error {
	safeData := sanitization.SanitizeBytes(data)
	*s = String(safeData)
	return nil
}

func (s String) String() string {
	return html.UnescapeString(string(s))
}

// CharData is a custom type for xml.CharData that can additionally sanitize the data.
type CharData xml.CharData

// UnmarshalText provides custom unmarshaling of CharData that will sanitize, unescape and trim whitespace from the
// value.
func (c *CharData) UnmarshalText(data []byte) error {
	*c = sanitization.SanitizeBytes(data)
	return nil
}

// UnmarshalJSON provides custom unmarshaling of CharData that will sanitize, unescape and trim whitespace from the
// value.
func (c *CharData) UnmarshalJSON(data []byte) error {
	var chardata struct {
		CharData []byte `json:"CharData"`
	}

	if err := json.Unmarshal(data, &chardata); err != nil {
		return fmt.Errorf("cannot unmarshal chardata: %w", err)
	}

	*c = sanitization.SanitizeBytes(chardata.CharData)

	return nil
}

func (c *CharData) String() string {
	return html.UnescapeString(string(*c))
}
