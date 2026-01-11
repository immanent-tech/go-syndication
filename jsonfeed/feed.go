// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: 	MIT

package jsonfeed

import (
	"slices"
	"time"

	"github.com/immanent-tech/go-syndication/sanitization"
	"github.com/immanent-tech/go-syndication/types"
	"github.com/immanent-tech/go-syndication/validation"
)

var _ types.FeedSource = (*Feed)(nil)

// GetTitle retrieves the title of the Feed.
func (f *Feed) GetTitle() string {
	return sanitization.SanitizeString(f.Title)
}

// GetDescription retrieves the description (if any) of the Feed.
func (f *Feed) GetDescription() string {
	if f.Description != nil {
		return sanitization.SanitizeString(*f.Description)
	}
	return ""
}

// GetSourceURL retrieves the URL that links to the JSONFeed file for the Feed.
func (f *Feed) GetSourceURL() string {
	if f.FeedURL != nil {
		return *f.FeedURL
	}
	return ""
}

// SetSourceURL will set a source URL, indicating the URL of the JSONFeed document, in the Feed.
func (f *Feed) SetSourceURL(url string) {
	f.FeedURL = &url
}

// GetLink retrieves the link of the Feed. This is the link to the website associated with the JSONFeed.
func (f *Feed) GetLink() string {
	if f.HomePageURL != nil {
		return *f.HomePageURL
	}
	return ""
}

// GetAuthors retrieves the authors (if any) of the Feed. This will be the list of values from any <author> and
// <dc:creator> elements.
func (f *Feed) GetAuthors() []string {
	authors := make([]string, 0, len(f.Authors))
	if f.Author != nil {
		authors = append(authors, f.Author.String())
	}
	for author := range slices.Values(f.Authors) {
		if author.String() != "" {
			authors = append(authors, author.String())
		}
	}
	return authors
}

// GetContributors is a no-op for a Feed.
func (f *Feed) GetContributors() []string {
	return nil
}

// GetRights is a no-op for a Feed.
func (f *Feed) GetRights() string {
	return ""
}

// GetLanguage retrieves the language (if any) of the Feed.
func (f *Feed) GetLanguage() string {
	if f.Language != nil {
		return *f.Language
	}
	return ""
}

// GetCategories is a no-op for a Feed.
func (f *Feed) GetCategories() []string {
	return nil
}

// GetImage retrieves the image (if any) for the Feed. It will retrieve the icon or favicon, whichever is found first,
// or an empty string if neither is found.
func (f *Feed) GetImage() *types.ImageInfo {
	var url string
	switch {
	case f.Icon != nil:
		url = *f.Icon
	case f.Favicon != nil:
		url = *f.Favicon
	}
	if url != "" {
		return &types.ImageInfo{
			URL: url,
		}
	}
	return nil
}

// SetImage sets an image for the Feed. This will set the icon value.
func (f *Feed) SetImage(image *types.ImageInfo) {
	url := image.GetURL()
	f.Icon = &url
}

// GetPublishedDate is tricky. We try to ascertain a published date from the newest published item. Otherwise this will
// be unix epoch.
func (f *Feed) GetPublishedDate() time.Time {
	published := time.Unix(0, 0)
	for item := range slices.Values(f.Items) {
		if item.GetPublishedDate().After(published) {
			published = item.GetPublishedDate()
		}
	}
	return published
}

// GetUpdatedDate is tricky. We try to ascertain a updated date from the newest modified item. Otherwise this will
// be unix epoch.
func (f *Feed) GetUpdatedDate() time.Time {
	modified := time.Unix(0, 0)
	for item := range slices.Values(f.Items) {
		if item.GetUpdatedDate().After(modified) {
			modified = item.GetUpdatedDate()
		}
	}
	return modified
}

func (f *Feed) GetUpdateInterval() time.Duration {
	if items := f.GetItems(); len(items) > 2 {
		var intervals []time.Duration
		for idx := range items {
			if idx < len(items)-1 {
				if items[idx].GetUpdatedDate() != types.UnixEpoch && items[idx+1].GetUpdatedDate() != types.UnixEpoch {
					intervals = append(intervals, items[idx].GetUpdatedDate().Sub(items[idx+1].GetUpdatedDate()))
				}
			}
		}
		if len(intervals) > 0 {
			return types.GetMedianInterval(intervals)
		}
	}
	return 5 * time.Minute
}

// GetItems returns a slice of Entry values for the Feed.
func (f *Feed) GetItems() []types.ItemSource {
	items := make([]types.ItemSource, 0, len(f.Items))
	for item := range slices.Values(f.Items) {
		items = append(items, &item)
	}
	return items
}

// Validate applies custom validation to an feed.
func (f *Feed) Validate() error {
	return validation.Validate.Struct(f)
}
