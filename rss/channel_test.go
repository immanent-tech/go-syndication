// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: 	MIT

package rss_test

import (
	"testing"
	"time"

	"github.com/immanent-tech/go-syndication/atom"
	"github.com/immanent-tech/go-syndication/extensions/dc"
	"github.com/immanent-tech/go-syndication/extensions/googleplay"
	"github.com/immanent-tech/go-syndication/extensions/itunes"
	"github.com/immanent-tech/go-syndication/extensions/media"
	"github.com/immanent-tech/go-syndication/extensions/syndication"
	"github.com/immanent-tech/go-syndication/rss"
	"github.com/stretchr/testify/assert"
)

func TestChannel_GetUpdateInterval(t *testing.T) {
	testTime := time.Now()
	tests := []struct {
		name    string
		channel rss.Channel
		want    time.Duration
	}{
		{
			name: "from update period",
			channel: rss.Channel{
				SYUdatePeriod: new(syndication.SYUpdatePeriod("hourly")),
			},
			want: time.Hour,
		},
		{
			name: "from update period and frequency",
			channel: rss.Channel{
				SYUdatePeriod:     new(syndication.SYUpdatePeriod("hourly")),
				SYUpdateFrequency: new(2),
			},
			want: time.Hour / 2,
		},
		{
			name: "from items",
			channel: rss.Channel{
				Items: []rss.Item{
					{PubDate: &rss.PubDate{Value: testTime}},
					{PubDate: &rss.PubDate{Value: testTime.Add(2 * time.Hour)}},
					{PubDate: &rss.PubDate{Value: testTime.Add(4 * time.Hour)}},
				},
			},
			want: 2 * time.Hour,
		},
		{
			name:    "default",
			channel: rss.Channel{},
			want:    time.Hour,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.channel.GetUpdateInterval()
			if got != tt.want {
				t.Errorf("GetUpdateInterval() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestChannel_GetPublishedDate(t *testing.T) {
	testTime := time.Now()
	tests := []struct {
		name    string
		channel rss.Channel
		want    *time.Time
		wantNil bool
	}{
		{
			name: "use <published>",
			channel: rss.Channel{
				PubDate: rss.NewPubDate(testTime),
			},
			want: &testTime,
		},
		{
			name: "use <dc:date>",
			channel: rss.Channel{
				DcDate: dc.Date{
					{Value: testTime},
				},
			},
			want: &testTime,
		},
		{
			name:    "no date",
			channel: rss.Channel{},
			wantNil: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.channel.GetPublishedDate()
			if !tt.wantNil {
				assert.WithinDuration(t, *tt.want, *got, time.Minute, "GetUpdatedDate")
			} else {
				assert.Nil(t, got, "no date")
			}
		})
	}
}

func TestChannel_GetUpdatedDate(t *testing.T) {
	testTime := time.Now()
	tests := []struct {
		name    string
		channel rss.Channel
		want    *time.Time
	}{
		{
			name: "use items",
			channel: rss.Channel{
				Items: []rss.Item{
					{PubDate: &rss.PubDate{Value: testTime.Add(-time.Hour)}},
					{PubDate: &rss.PubDate{Value: testTime.Add(-2 * time.Hour)}},
					{PubDate: &rss.PubDate{Value: testTime.Add(-4 * time.Hour)}},
				},
			},
			want: new(testTime.Add(-time.Hour)),
		},
		{
			name: "unreliable items",
			channel: rss.Channel{
				Items: []rss.Item{
					{PubDate: &rss.PubDate{Value: testTime.Add(-time.Hour)}},
					{},
					{PubDate: &rss.PubDate{Value: testTime.Add(-2 * time.Hour)}},
					{PubDate: &rss.PubDate{Value: testTime.Add(-4 * time.Hour)}},
					{},
				},
			},
			want: new(testTime.Add(-time.Hour)),
		},
		{
			name: "use <lastBuildDate>",
			channel: rss.Channel{
				LastBuildDate: &rss.LastBuildDate{Value: testTime},
			},
			want: &testTime,
		},
		{
			name: "use <pubDate>",
			channel: rss.Channel{
				PubDate: &rss.LastBuildDate{Value: testTime},
			},
			want: &testTime,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.channel.GetUpdatedDate()
			assert.WithinDuration(t, *tt.want, *got, time.Minute, "GetUpdatedDate")
		})
	}
}

func TestChannel_GetImage(t *testing.T) {
	tests := []struct {
		name      string
		channel   rss.Channel
		wantURL   string
		wantTitle string
		wantNil   bool
	}{
		{
			name: "use <image>",
			channel: rss.Channel{
				Image: &rss.Image{
					URL: "http://example.com/image.webp",
				},
			},
			wantURL: "http://example.com/image.webp",
		},
		{
			name: "use <media:content>",
			channel: rss.Channel{
				MediaContent: &media.MediaContent{
					Medium: new(media.MediaContentMediumImage),
					URL:    "http://example.com/image.webp",
				},
			},
			wantURL: "http://example.com/image.webp",
		},
		{
			name: "use <media:thumbnail>",
			channel: rss.Channel{
				MediaThumbnails: media.MediaThumbnails{
					{URL: "http://example.com/image.webp"},
				},
			},
			wantURL: "http://example.com/image.webp",
		},
		{
			name: "use <itunes:image>",
			channel: rss.Channel{
				ItunesImage: &itunes.Image{
					Href: "http://example.com/image.webp",
				},
			},
			wantURL: "http://example.com/image.webp",
		},
		{
			name:    "no image",
			channel: rss.Channel{},
			wantNil: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.channel.GetImage()
			if tt.wantNil {
				assert.Nil(t, got)
			} else {
				assert.Equal(t, tt.wantURL, got.GetURL(), "GetImage URL")
				assert.Equal(t, tt.wantTitle, got.GetTitle(), "GetImage Title")
			}
		})
	}
}

func TestChannel_GetSourceURL(t *testing.T) {
	tests := []struct {
		name    string
		channel rss.Channel
		want    string
	}{
		{
			name: "multiple <atom:link>",
			channel: rss.Channel{
				AtomLinks: []rss.AtomLink{
					{Rel: new(atom.LinkRelRelated), Href: "http://example.com"},
					{Rel: new(atom.LinkRelSelf), Href: "http://example.com/rss"},
				},
			},
			want: "http://example.com/rss",
		},
		{
			name: "no rel=self <atom:link>",
			channel: rss.Channel{
				AtomLinks: []rss.AtomLink{
					{Rel: new(atom.LinkRelRelated), Href: "http://example.com"},
				},
			},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.channel.GetSourceURL()
			if got != tt.want {
				t.Errorf("GetSourceURL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestChannel_GetLink(t *testing.T) {
	tests := []struct {
		name    string
		channel rss.Channel
		want    string
	}{
		{
			name: "has <link>",
			channel: rss.Channel{
				Link: "http://example.com",
			},
			want: "http://example.com",
		},
		{
			name: "use rel=alt <atom:link>",
			channel: rss.Channel{
				AtomLinks: []rss.AtomLink{
					{Rel: new(atom.LinkRelAlternate), Href: "http://example.com"},
					{Rel: new(atom.LinkRelSelf), Href: "http://example.com/rss"},
				},
			},
			want: "http://example.com",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.channel.GetLink()
			if got != tt.want {
				t.Errorf("GetSourceURL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestChannel_GetAuthors(t *testing.T) {
	tests := []struct {
		name    string
		channel rss.Channel
		want    []string
	}{
		{
			name: "with authors",
			channel: rss.Channel{
				WebMaster: new("author1"),
			},
			want: []string{"author1"},
		},
		{
			name: "multiple sources",
			channel: rss.Channel{
				WebMaster: new("author1"),
				DcPublisher: dc.Publisher{
					"author2",
				},
				DcCreator: dc.Creator{
					"author3",
				},
			},
			want: []string{"author1", "author2", "author3"},
		},
		{
			name: "no duplicates",
			channel: rss.Channel{
				WebMaster: new("author1"),
				DcPublisher: dc.Publisher{
					"author1",
				},
			},
			want: []string{"author1"},
		},
		{
			name:    "no authors",
			channel: rss.Channel{},
			want:    nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.channel.GetAuthors()
			assert.EqualValues(t, tt.want, got)
		})
	}
}

func TestChannel_GetContributors(t *testing.T) {
	tests := []struct {
		name    string
		channel rss.Channel
		want    []string
	}{
		{
			name: "with contributors",
			channel: rss.Channel{
				ManagingEditor: new("contributor1"),
			},
			want: []string{"contributor1"},
		},
		{
			name: "with contributors",
			channel: rss.Channel{
				ManagingEditor: new("contributor1"),
				DcContributor: dc.Contributor{
					"contributor2",
				},
			},
			want: []string{"contributor1", "contributor2"},
		},
		{
			name: "no duplicates",
			channel: rss.Channel{
				ManagingEditor: new("contributor1"),
				DcContributor: dc.Contributor{
					"contributor1",
				},
			},
			want: []string{"contributor1"},
		},
		{
			name:    "no contributors",
			channel: rss.Channel{},
			want:    nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.channel.GetContributors()
			assert.EqualValues(t, tt.want, got)
		})
	}
}

func TestChannel_GetRights(t *testing.T) {
	tests := []struct {
		name    string
		channel rss.Channel
		want    string
		wantNil bool
	}{
		{
			name: "with <copyright>",
			channel: rss.Channel{
				Copyright: new("rights"),
			},
			want: "rights",
		},
		{
			name: "with <dc:rights>",
			channel: rss.Channel{
				DcRights: dc.Rights{
					"rights",
				},
			},
			want: "rights",
		},
		{
			name:    "no rights",
			channel: rss.Channel{},
			wantNil: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.channel.GetRights()
			if !tt.wantNil {
				if *got != tt.want {
					t.Errorf("GetRights() = %v, want %v", got, tt.want)
				}
			} else {
				assert.Nil(t, got)
			}
		})
	}
}

func TestChannel_GetLanguage(t *testing.T) {
	tests := []struct {
		name    string
		channel rss.Channel
		want    string
		wantNil bool
	}{
		{
			name: "with <language>",
			channel: rss.Channel{
				Language: new("en-us"),
			},
			want: "en-us",
		},
		{
			name: "with <dc:language>",
			channel: rss.Channel{
				DcLanguage: dc.Language{
					"en-us",
				},
			},
			want: "en-us",
		},
		{
			name:    "no language",
			channel: rss.Channel{},
			wantNil: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.channel.GetLanguage()
			if !tt.wantNil {
				if *got != tt.want {
					t.Errorf("GetLanguage() = %v, want %v", got, tt.want)
				}
			} else {
				assert.Nil(t, got)
			}
		})
	}
}

func TestChannel_GetCategories(t *testing.T) {
	tests := []struct {
		name    string
		channel rss.Channel
		want    []string
	}{
		{
			name: "has categories",
			channel: rss.Channel{
				Categories: []rss.Category{
					{Value: "category1"},
					{Value: "category2"},
				},
			},
			want: []string{"category1", "category2"},
		},
		{
			name: "no duplicates",
			channel: rss.Channel{
				Categories: []rss.Category{
					{Value: "category1"},
					{Value: "category2"},
					{Value: "category1"},
				},
			},
			want: []string{"category1", "category2"},
		},
		{
			name: "multiple sources",
			channel: rss.Channel{
				Categories: []rss.Category{
					{Value: "category1"},
					{Value: "category2"},
				},
				DcSubject: dc.Subject{
					"category3",
				},
				GooglePlayCategory: &googleplay.Category{Text: "category4"},
				ItunesCategory: &itunes.Categories{
					Text: "test",
					Categories: []itunes.Category{
						{Text: "category5"},
					},
				},
				MediaCategory: &media.MediaCategory{Value: "category6"},
			},
			want: []string{"category1", "category2", "category3", "category4", "category6", "test", "test|category5"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.channel.GetCategories()
			assert.EqualValues(t, got, tt.want)
		})
	}
}
