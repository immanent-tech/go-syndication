// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: 	MIT

package rss_test

import (
	"testing"
	"time"

	"github.com/immanent-tech/go-syndication/atom"
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.channel.GetImage()
			assert.Equal(t, tt.wantURL, got.GetURL(), "GetImage URL")
			assert.Equal(t, tt.wantTitle, got.GetTitle(), "GetImage Title")
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
