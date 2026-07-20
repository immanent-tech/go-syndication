// Copyright 2026 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: 	AGPL-3.0-or-later

package rdf

import (
	"strings"
	"time"

	"github.com/immanent-tech/go-syndication/types"
)

var _ types.ItemSource = (*Item)(nil)

func (i *Item) GetAuthors() []string {
	if i.Creator != nil {
		return *i.Creator
	}
	return nil
}

func (i *Item) GetContributors() []string {
	if i.Contributor != nil {
		return *i.Contributor
	}
	return nil
}

func (i *Item) GetCategories() []string {
	if i.Subject != nil {
		return *i.Subject
	}
	return nil
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
	if i.Language != nil {
		return new(strings.Join(*i.Language, " "))
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

func (i *Item) GetImage() *types.ImageInfo {
	return nil
}

func (i *Item) GetPublishedDate() *time.Time {
	if i.Date != nil {
		v := (*i.Date)[0].Value
		return &v
	}
	return nil
}

func (i *Item) GetUpdatedDate() *time.Time {
	return nil
}

func (i *Item) GetRights() *string {
	if i.Rights != nil {
		return new(strings.Join(*i.Rights, " "))
	}
	return nil
}
