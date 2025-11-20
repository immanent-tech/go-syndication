// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: 	AGPL-3.0-or-later

package atom

import (
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/immanent-tech/go-syndication/extensions/media"
	"github.com/immanent-tech/go-syndication/types"
	"github.com/immanent-tech/go-syndication/validation"
)

var _ types.FeedSource = (*Feed)(nil)

var ErrFeedValidation = errors.New("feed is invalid")

// GetTitle retrieves the <title> of the Feed.
func (f *Feed) GetTitle() string {
	switch {
	case f.DCTitle != nil:
		return f.DCTitle.String()
	case f.Title.String() != "":
		return f.Title.String()
	default:
		return ""
	}
}

// GetDescription retrieves the <description> (if any) of the Feed.
func (f *Feed) GetDescription() string {
	switch {
	case f.DCDescription != nil:
		return f.DCDescription.String()
	case f.Subtitle.String() != "":
		return f.Subtitle.String()
	default:
		return ""
	}
}

// GetSourceURL retrieves the URL that links to the Atom file for the Feed. This will be any <link> element
// present with a "rel" attribute of "self" and ideally with a mime-type indicating Atom content.
func (f *Feed) GetSourceURL() string {
	for link := range slices.Values(f.Links) {
		if link.Rel != "" && link.Rel == LinkRelSelf {
			if link.Type != "" && slices.Contains(types.MimeTypesAtom, link.Type) {
				return link.Href
			}
		}
	}
	return ""
}

// SetSourceURL will set a source URL, indicating the URL of the Atom document, in the Feed.
func (f *Feed) SetSourceURL(url string) {
	rel := LinkRelSelf
	f.Links = append(f.Links, Link{Href: url, Rel: rel, Type: types.MimeTypesAtom[0]})
}

// GetLink retrieves the <link> of the Feed. This is the link to the website associated with the Atom feed. Even the
// spec is ambiguous about what link attributes constitute the correct combination to indicate the site, so we apply
// some guesses here.
func (f *Feed) GetLink() string {
	for link := range slices.Values(f.Links) {
		// If there is a rel=self link that does not point to an atom document, use that.
		if link.Rel == LinkRelSelf {
			if !slices.Contains(types.MimeTypesAtom, link.Type) {
				return link.Href
			}
		}
		// If there is a rel=alt, use that.
		if link.Rel == LinkRelAlternate {
			return link.Href
		}
	}
	return ""
}

// GetAuthors retrieves the authors (if any) of the Feed. This will be the list of values from any <author> and
// <dc:creator> elements.
func (f *Feed) GetAuthors() []string {
	var authors []string
	if len(f.Authors) > 0 {
		for author := range slices.Values(f.Authors) {
			authors = append(authors, author.String())
		}
	}
	if f.DCCreator != nil {
		authors = append(authors, f.DCCreator.String())
	}
	return authors
}

// GetContributors retrieves the contributors (if any) of the Feed. This will be the list of values from any
// <contributor> and <dc:contributor> elements.
func (f *Feed) GetContributors() []string {
	var contributors []string
	if len(f.Contributors) > 0 {
		for contributor := range slices.Values(f.Contributors) {
			contributors = append(contributors, contributor.String())
		}
	}
	if f.DCContributor != nil {
		contributors = append(contributors, f.DCCreator.String())
	}
	return contributors
}

// GetRights retrieves the rights (copyright) of the Feed. This will be the first value found from either <dc:rights>
// or <rights> elements.
func (f *Feed) GetRights() string {
	switch {
	case f.DCRights != nil:
		return f.DCRights.String()
	case f.Rights.Value != "":
		return f.Rights.Value
	default:
		return ""
	}
}

// GetLanguage retrieves the language of the Feed. This will be the first value found from either <dc:language>
// or <lang> elements.
func (f *Feed) GetLanguage() string {
	switch {
	case f.DCLanguage != "":
		return f.DCLanguage
	case f.Lang != "":
		return f.Lang
	default:
		return ""
	}
}

// GetCategories retrieves the categories (if any) of the Feed. The categories are returned as strings.
func (f *Feed) GetCategories() []string {
	categories := make([]string, 0, len(f.Categories))
	for category := range slices.Values(f.Categories) {
		categories = append(categories, category.String())
	}
	return categories
}

// GetImage retrieves the image (if any) for the Feed. The image is returned as a types.ImageInfo object. The value will be
// the first found of <media:thumbnail> element.
func (f *Feed) GetImage() *types.ImageInfo {
	if len(f.MediaThumbnails) > 0 {
		thumbnail := f.MediaThumbnails[0]
		return &types.ImageInfo{
			URL:   thumbnail.URL,
			Title: f.GetTitle(),
		}
	}
	return nil
}

// SetImage sets an image for the Channel.
func (f *Feed) SetImage(image *types.ImageInfo) {
	f.MediaThumbnails = media.MediaThumbnails{
		media.MediaThumbnail{URL: image.GetURL()},
	}
}

// GetPublishedDate returns the <published> element of the Feed.
func (f *Feed) GetPublishedDate() time.Time {
	return f.Updated.Value.Time
}

// GetUpdatedDate returns the <updated> of the Feed.
func (f *Feed) GetUpdatedDate() time.Time {
	return f.Updated.Value.Time
}

// GetItems returns a slice of Entry values for the Feed.
func (f *Feed) GetItems() []types.ItemSource {
	items := make([]types.ItemSource, 0, len(f.Entries))
	for item := range slices.Values(f.Entries) {
		items = append(items, &item)
	}
	return items
}

// Validate applies custom validation to an feed.
func (f *Feed) Validate() error {
	// Check for all entries having authors.
	var missingEntryAuthors bool
	for entry := range slices.Values(f.GetItems()) {
		if len(entry.GetAuthors()) == 0 {
			missingEntryAuthors = true
			break
		}
	}
	// atom:feed elements MUST contain one or more atom:author elements, unless all of the atom:feed element's child
	// atom:entry elements  contain at least one atom:author element.
	//
	// https://www.rfc-editor.org/rfc/rfc4287#page-11
	if len(f.GetAuthors()) == 0 && missingEntryAuthors {
		return fmt.Errorf("%w: must have at least one author or all entries with authors", ErrFeedValidation)
	}
	err := validation.Validate.Struct(f)
	if err != nil {
		return fmt.Errorf("feed validation failed: %w", err)
	}
	return nil
}
