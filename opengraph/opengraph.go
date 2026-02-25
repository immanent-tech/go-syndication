// Copyright 2026 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: 	AGPL-3.0-or-later

package opengraph

import (
	"encoding/xml"
	"slices"
	"strings"
)

// New creates a new open graph object with the given required values and any additional values as options.
func New(title Title, objectType ObjectType, url URL, image Image, options ...Option) *OpenGraph {
	og := &OpenGraph{
		Title:                title,
		ObjectType:           objectType,
		URL:                  url,
		Image:                image,
		AdditionalProperties: make(map[string]string),
	}
	for option := range slices.Values(options) {
		option(og)
	}

	return og
}

type Option func(*OpenGraph)

func WithDescription(desc Description) Option {
	return func(og *OpenGraph) {
		og.Description = &desc
	}
}

func WithSiteName(name SiteName) Option {
	return func(og *OpenGraph) {
		og.SiteName = &name
	}
}

func WithAudio(audio Audio) Option {
	return func(og *OpenGraph) {
		og.Audio = &audio
	}
}

func WithVideo(video Video) Option {
	return func(og *OpenGraph) {
		og.Video = &video
	}
}

func WithLocale(locale Locale) Option {
	return func(og *OpenGraph) {
		og.Locale = &locale
	}
}

func WithAdditionalProperty(key, value string) Option {
	return func(og *OpenGraph) {
		og.Set(key, value)
	}
}

// UnmarshalXML implements xml.Unmarshaler.
// It scans all <meta> elements anywhere in the document and populates
// OGMeta fields from those whose property/name attribute starts with "og:".
func (og *OpenGraph) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	if og.AdditionalProperties == nil {
		og.AdditionalProperties = make(map[string]string)
	}

	for {
		tok, err := d.Token()
		if err != nil {
			// io.EOF is expected at end of document
			break
		}

		se, ok := tok.(xml.StartElement)
		if !ok {
			continue
		}

		if !strings.EqualFold(se.Name.Local, "meta") {
			continue
		}

		var m metaTag
		// Decode the element (handles self-closing tags too).
		if err := d.DecodeElement(&m, &se); err != nil {
			continue
		}

		// Use whichever attribute is set: property takes precedence over name.
		key := m.Property
		if key == "" {
			key = m.Name
		}
		key = strings.ToLower(strings.TrimSpace(key))

		if !strings.HasPrefix(key, "og:") {
			continue
		}

		og.extract(key, m.Content)
	}

	return nil
}

// extract assigns a parsed og: property to the appropriate struct field.
func (og *OpenGraph) extract(property, content string) {
	switch property {
	case "og:title":
		og.Title = content
	case "og:description":
		og.Description = &content
	case "og:image":
		og.Image = content
	case "og:url":
		og.URL = content
	case "og:type":
		og.ObjectType = content
	case "og:site_name":
		og.SiteName = &content
	case "og:audio":
		og.Audio = &content
	case "og:video":
		og.Video = &content
	case "og:locale":
		og.Locale = &content
	default:
		if og.AdditionalProperties == nil {
			og.AdditionalProperties = make(map[string]string)
		}
		og.Set(property, content)
	}
}

// metaTag is a minimal representation of a <meta> element for XML decoding.
type metaTag struct {
	Property string `xml:"property,attr"`
	Name     string `xml:"name,attr"`
	Content  string `xml:"content,attr"`
}
