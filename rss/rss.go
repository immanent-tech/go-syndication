// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: 	MIT

// Package rss contains objects and methods defining the RSS syndication format.
//
//revive:disable:exported // function definitions can be ascertained from Channel.
package rss

import (
	"encoding/xml"
	"fmt"
	"slices"
	"time"

	"github.com/immanent-tech/go-syndication/extensions/rss"
	"github.com/immanent-tech/go-syndication/types"
	"github.com/immanent-tech/go-syndication/validation"
)

var _ types.FeedSource = (*RSS)(nil)

// String returns the value of the Category.
func (c *Category) String() string {
	return c.Value.String()
}

// NewRSS creates a new RSS version 2.0 object with the required title, description and link values and any given
// options.
func NewRSS(title, description, link string, options ...RSSOption) *RSS {
	rss := &RSS{
		XMLName: xml.Name{Local: "rss"},
		Version: N20,
		Channel: Channel{
			Title:         Title(title),
			Description:   Description(description),
			Link:          Link(link),
			LastBuildDate: &types.DateTime{Time: time.Now().UTC()},
			Generator:     "go-syndication",
			Docs:          types.String("https://www.rssboard.org/rss-specification"),
		},
	}

	for option := range slices.Values(options) {
		option(rss)
	}

	return rss
}

// RSSOption is a functional applied to an RSS object.
type RSSOption func(*RSS)

// WithCopyright option sets the RSS channel copyright.
func WithCopyright(copyright string) RSSOption {
	return func(r *RSS) {
		r.Channel.Copyright = types.String(copyright)
	}
}

// WithGenerator options sets the generator. This will default to "go-syndication".
func WithGenerator(generator string) RSSOption {
	return func(r *RSS) {
		r.Channel.Generator = types.String(generator)
	}
}

// WithManagingEditor option sets the RSS channel managingEditor.
func WithManagingEditor(editor string) RSSOption {
	return func(r *RSS) {
		r.Channel.ManagingEditor = types.String(editor)
	}
}

// WithWebmaster option sets the RSS channel webmaster.
func WithWebmaster(webmaster string) RSSOption {
	return func(r *RSS) {
		r.Channel.WebMaster = types.String(webmaster)
	}
}

// WithLastBuildDate option sets the last build date of the RSS object. This will default to time.Now().UTC().
func WithLastBuildDate(ts time.Time) RSSOption {
	return func(r *RSS) {
		r.Channel.LastBuildDate.Time = ts
	}
}

// WithPublishedDate option sets the published date of the RSS object.
func WithPublishedDate(ts time.Time) RSSOption {
	return func(r *RSS) {
		r.Channel.PubDate = &types.DateTime{}
		r.Channel.PubDate.Time = ts
	}
}

// WithChannelLanguage option sets the RSS channel language. Should be an ISO country code to be valid.
func WithChannelLanguage(lang string) RSSOption {
	return func(r *RSS) {
		r.Channel.Language = types.String(lang)
	}
}

// WithChannelImage option sets an image.
func WithChannelImage(image *Image) RSSOption {
	return func(r *RSS) {
		r.Channel.Image = image
	}
}

// WithUpdatePeriod option sets the update period of the feed.
func WithUpdatePeriod(up string) RSSOption {
	return func(r *RSS) {
		r.Channel.SYUdatePeriod = &rss.SYUpdatePeriod{
			Value: up,
		}
	}
}

func (r *RSS) GetTitle() string {
	return r.Channel.GetTitle()
}

func (r *RSS) GetDescription() string {
	return r.Channel.GetDescription()
}

func (r *RSS) GetSourceURL() string {
	return r.Channel.GetSourceURL()
}

func (r *RSS) SetSourceURL(url string) {
	r.Channel.SetSourceURL(url)
}

func (r *RSS) GetLink() string {
	return r.Channel.GetLink()
}

func (r *RSS) GetUpdatedDate() time.Time {
	return r.Channel.GetUpdatedDate()
}

func (r *RSS) GetPublishedDate() time.Time {
	return r.Channel.GetPublishedDate()
}

func (r *RSS) GetCategories() []string {
	return r.Channel.GetCategories()
}

func (r *RSS) GetAuthors() []string {
	return r.Channel.GetAuthors()
}

func (r *RSS) GetContributors() []string {
	return r.Channel.GetContributors()
}

func (r *RSS) GetRights() string {
	return r.Channel.GetRights()
}

func (r *RSS) GetLanguage() string {
	return r.Channel.GetLanguage()
}

func (r *RSS) GetImage() *types.ImageInfo {
	return r.Channel.GetImage()
}

func (r *RSS) SetImage(image *types.ImageInfo) {
	r.Channel.SetImage(image)
}

func (r *RSS) GetItems() []types.ItemSource {
	return r.Channel.GetItems()
}

func (r *RSS) GetUpdateInterval() time.Duration {
	return r.Channel.GetUpdateInterval()
}

// Validate applies custom validation to an feed.
func (r *RSS) Validate() error {
	if err := validation.ValidateStruct(r); err != nil {
		return fmt.Errorf("rss validation failed: %w", err)
	}
	return nil
}
