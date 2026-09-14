// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: 	MIT

// Package media contains objects and methods defining the MediaRSS extension.
package media

import (
	"encoding/xml"
	"fmt"
	"net/url"
	"slices"
	"strconv"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/immanent-tech/go-syndication/sanitization"
	"github.com/immanent-tech/go-syndication/types"
	"github.com/immanent-tech/go-syndication/validation"
)

func init() {
	validation.RegisterStructValidation(mediaContentCustomValidation, MediaContent{})
	validation.RegisterStructValidation(mediaRestrictionCustomValidation, MediaRestriction{})
	validation.RegisterStructValidation(mediaGroupCustomValidation, MediaGroup{})
}

var (
	// ImageExt contains canonical/standard/common file extensions for images.
	ImageExt = []string{"jpg", "jpeg", "png", "webp", "gif"}
)

// AsImage returns the <media:thumbnail> object as a types.ImageInfo object.
func (t *MediaThumbnail) AsImage() *types.Image {
	return &types.Image{
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
func (c *MediaContent) AsImage() *types.Image {
	// Check if medium attr indicates an image.
	if c.Medium != nil && *c.Medium == MediaContentMediumImage {
		return &types.Image{
			URL: c.URL,
		}
	}
	// Check if mimetype attr indicates an image.
	if c.Type != nil && types.IsImage(*c.Type) {
		return &types.Image{
			URL: c.URL,
		}
	}
	// Ugh, maybe try parsing the URL and see if it ends in a well-known image file extension...
	if url, err := url.Parse(c.URL); err == nil {
		for imgext := range slices.Values(ImageExt) {
			if strings.HasSuffix(url.Path, imgext) {
				return &types.Image{
					URL: c.URL,
				}
			}
		}
	}

	return nil
}

func mediaContentCustomValidation(sl validator.StructLevel) {
	c := sl.Current().Interface().(MediaContent)
	if c.URL == "" && c.MediaPlayer == nil {
		sl.ReportError(c, "URL", "URL", "required", "either url or a media:player child is required")
	}
}

func (g *MediaGroup) GetDescription() string {
	if g.MediaDescription != nil {
		return sanitization.SanitizeString(g.MediaDescription.Value)
	}
	return ""
}

func mediaGroupCustomValidation(sl validator.StructLevel) {
	g := sl.Current().Interface().(MediaGroup)
	if len(g.Content) <= 1 {
		sl.ReportError(g.Content, "Content", "Content", "gt=1", "must have multiple media:content children")
	}
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

func mediaRestrictionCustomValidation(sl validator.StructLevel) {
	r := sl.Current().Interface().(MediaRestriction)
	if r.Relationship != "allow" && r.Relationship != "deny" {
		sl.ReportError(
			r.Relationship,
			"Relationship",
			"Relationship",
			"oneof",
			fmt.Sprintf("relationship must be \"allow\" or \"deny\", got %q", r.Relationship),
		)
	}
	if v := strings.TrimSpace(r.Value); v == "all" || v == "none" {
		return // type may legitimately be omitted for these reserved literals
	}
	switch *r.Type {
	case "country", "uri", "sharing":
		return
	default:
		sl.ReportError(
			r.Type,
			"Type",
			"Type",
			"oneof",
			fmt.Sprintf(
				"type must be \"country\", \"uri\", or \"sharing\" unless value is \"all\"/\"none\", got %q",
				*r.Type,
			),
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
