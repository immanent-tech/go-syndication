// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: 	MIT

// Package media contains objects and methods defining the MediaRSS extension.
package media

import (
	"encoding/xml"
	"errors"
	"fmt"
	"net/url"
	"slices"
	"strconv"
	"strings"

	"github.com/immanent-tech/go-syndication/sanitization"
	"github.com/immanent-tech/go-syndication/types"
)

// AsImage returns the <media:thumbnail> object as a types.ImageInfo object.
func (t *MediaThumbnail) AsImage() *types.ImageInfo {
	return &types.ImageInfo{
		URL: t.URL,
	}
}

// GetCategory retrieves the category assigned to the media:content element (if any).
func (c *MediaContent) GetCategory() string {
	if c.MediaCategory != nil {
		if c.MediaCategory.Label != nil {
			return *c.MediaCategory.Label
		}
		return sanitization.SanitizeString(c.MediaCategory.Value)
	}
	return ""
}

// GetText retrieves the text of media:content element (if any).
func (t *MediaText) GetText() string {
	return sanitization.SanitizeString(t.Value)
}

// AsImage will return a types.ImageInfo if the <media:content> element represents an image. If not, it will return nil.
func (c *MediaContent) AsImage() *types.ImageInfo {
	// Check if medium attr indicates an image.
	if c.Medium != nil && *c.Medium == MediaContentMediumImage {
		return &types.ImageInfo{
			URL: c.URL,
		}
	}
	// Check if mimetype attr indicates an image.
	if c.Type != nil && types.IsImage(*c.Type) {
		return &types.ImageInfo{
			URL: c.URL,
		}
	}
	// Ugh, maybe try parsing the URL and see if it ends in a well-known image file extension...
	if url, err := url.Parse(c.URL); err == nil {
		for imgext := range slices.Values(types.MediaImageExt) {
			if strings.HasSuffix(url.Path, imgext) {
				return &types.ImageInfo{
					URL: c.URL,
				}
			}
		}
	}

	return nil
}

// Validate enforces "URL should specify the direct URL... If not included, a media:player element must be specified.".
func (c MediaContent) Validate() error {
	if c.URL == "" && c.MediaPlayer == nil {
		return errors.New("media:content: either url or a media:player child is required")
	}
	return nil
}

func (g *MediaGroup) GetDescription() string {
	if g.MediaDescription != nil {
		return sanitization.SanitizeString(g.MediaDescription.Value)
	}
	return ""
}

func (k MediaKeywords) MarshalXML(enc *xml.Encoder, start xml.StartElement) error {
	if len(k) == 0 {
		return nil
	}
	if err := enc.EncodeToken(start); err != nil {
		return fmt.Errorf("encode media keywords: %w", err)
	}
	if err := enc.EncodeToken(xml.CharData(strings.Join(k, ", "))); err != nil {
		return fmt.Errorf("encode media keywords: %w", err)
	}
	if err := enc.EncodeToken(start.End()); err != nil {
		return fmt.Errorf("encode media keywords: %w", err)
	}
	return nil
}

func (k *MediaKeywords) UnmarshalXML(dec *xml.Decoder, start xml.StartElement) error {
	var valueStruct struct {
		Value string `xml:",chardata"`
	}
	if err := dec.DecodeElement(&valueStruct, &start); err != nil {
		return fmt.Errorf("decode media keywords: %w", err)
	}
	*k = nil
	for part := range strings.SplitSeq(valueStruct.Value, ",") {
		if s := strings.TrimSpace(part); s != "" {
			*k = append(*k, s)
		}
	}
	return nil
}

// Validate enforces the rules the struct shape alone can't:
// relationship is required and must be allow/deny; type is required
// unless the value is exactly the reserved literal "all" or "none".
func (r MediaRestriction) Validate() error {
	if r.Relationship != "allow" && r.Relationship != "deny" {
		return fmt.Errorf("media:restriction: relationship must be \"allow\" or \"deny\", got %q", r.Relationship)
	}
	v := strings.TrimSpace(r.Value)
	if v == "all" || v == "none" {
		return nil // type may legitimately be omitted for these reserved literals
	}
	switch *r.Type {
	case "country", "uri", "sharing":
		return nil
	default:
		return fmt.Errorf(
			"media:restriction: type must be \"country\", \"uri\", or \"sharing\" unless value is \"all\"/\"none\", got %q",
			r.Type,
		)
	}
}

func (t MediaTags) MarshalXML(enc *xml.Encoder, start xml.StartElement) error {
	if len(t) == 0 {
		return nil
	}
	parts := make([]string, 0, len(t))
	for _, tag := range t {
		if tag.Weight != 0 && tag.Weight != 1 {
			parts = append(parts, fmt.Sprintf("%s:%d", tag.Name, tag.Weight))
		} else {
			parts = append(parts, tag.Name)
		}
	}
	if err := enc.EncodeToken(start); err != nil {
		return fmt.Errorf("marshal media tags: %w", err)
	}
	if err := enc.EncodeToken(xml.CharData(strings.Join(parts, ", "))); err != nil {
		return fmt.Errorf("marshal media tags: %w", err)
	}
	if err := enc.EncodeToken(start.End()); err != nil {
		return fmt.Errorf("marshal media tags: %w", err)
	}
	return nil
}

func (t *MediaTags) UnmarshalXML(dec *xml.Decoder, start xml.StartElement) error {
	var valueStruct struct {
		Value string `xml:",chardata"`
	}
	if err := dec.DecodeElement(&valueStruct, &start); err != nil {
		return fmt.Errorf("unmarshal media tags: %w", err)
	}
	*t = nil
	for part := range strings.SplitSeq(valueStruct.Value, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		name, weightStr, hasWeight := strings.Cut(part, ":")
		weight := 1
		if hasWeight {
			if w, err := strconv.Atoi(strings.TrimSpace(weightStr)); err == nil {
				weight = w
			}
		}
		*t = append(*t, MediaTag{Name: strings.TrimSpace(name), Weight: weight})
	}
	return nil
}
