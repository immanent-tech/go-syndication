// Copyright 2026 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: 	AGPL-3.0-or-later

package rdf

import (
	"slices"
	"strings"
	"time"

	"github.com/immanent-tech/go-syndication/sanitization"
	"github.com/immanent-tech/go-syndication/types"
)

var _ types.ItemSource = (*Item)(nil)

func (i *Item) GetAuthors() []string {
	if len(i.DcCreator) > 0 {
		return i.DcCreator
	}
	return nil
}

func (i *Item) GetContributors() []string {
	if len(i.DcContributor) > 0 {
		return i.DcContributor
	}
	return nil
}

func (i *Item) GetCategories() []string {
	categories := make([]string, 0)
	if len(i.DcSubject) > 0 {
		categories = append(categories, i.DcSubject...)
	}
	// Treat <slash:department> as a category.
	if i.SlashDepartment != nil {
		categories = append(categories, sanitization.SanitizeText(*i.SlashDepartment, sanitization.DefaultMaxRunes))
	}
	// Treat <slash:section> as a category.
	if i.SlashSection != nil {
		categories = append(categories, sanitization.SanitizeText(*i.SlashSection, sanitization.DefaultMaxRunes))
	}
	categories = slices.DeleteFunc(categories, func(str string) bool {
		return strings.TrimSpace(str) == ""
	})
	slices.Sort(categories)
	return slices.CompactFunc(categories, func(a, b string) bool {
		return strings.EqualFold(a, b)
	})
}

func (i *Item) GetDescription() string {
	if i.Description != nil {
		return *i.Description
	}
	return ""
}

func (i *Item) GetTitle() string {
	return i.Title
}

func (i *Item) GetLanguage() *string {
	if len(i.DcLanguage) > 0 {
		return new(strings.Join(i.DcLanguage, " "))
	}
	return nil
}

func (i *Item) GetLink() string {
	return i.Link
}

func (i *Item) GetContent() *string {
	return i.Description
}

func (i *Item) GetID() string {
	return ""
}

func (i *Item) GetImage() *types.Image {
	return nil
}

func (i *Item) GetPublishedDate() *time.Time {
	if len(i.DcDate) > 0 {
		v := (i.DcDate)[0].Value
		return &v
	}
	if len(i.DcDate) == 1 {
		dcDate := i.DcDate[0]
		if !dcDate.Value.IsZero() {
			return new(dcDate.Value)
		}
	}
	return nil
}

func (i *Item) GetUpdatedDate() *time.Time {
	return nil
}

func (i *Item) GetRights() *string {
	if len(i.DcRights) > 0 {
		return new(strings.Join(i.DcRights, " "))
	}
	return nil
}

func (i *Item) GetGeoInfo() *types.GeoInfo {
	return nil
}

// GetCommentRSS returns whichever spelling was present after decoding a feed preferring the correct one if a
// (spec-violating) feed somehow supplied both.
func (i *Item) GetCommentRSS() string {
	switch {
	case i.CommentRSS != nil:
		return *i.CommentRSS
	case i.CommentRSSLegacy != nil:
		return *i.CommentRSSLegacy
	default:
		return ""
	}
}
