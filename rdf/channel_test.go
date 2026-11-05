// Copyright 2026 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: 	AGPL-3.0-or-later

package rdf_test

import (
	"testing"
	"time"

	"github.com/immanent-tech/go-syndication/extensions/dc"
	"github.com/immanent-tech/go-syndication/extensions/media"
	"github.com/immanent-tech/go-syndication/extensions/syndication"
	"github.com/immanent-tech/go-syndication/rdf"
	"github.com/stretchr/testify/assert"
)

func TestChannel_GetUpdateInterval(t *testing.T) {
	tests := []struct {
		name    string
		channel rdf.Channel
		want    time.Duration
	}{
		{
			name: "from update period",
			channel: rdf.Channel{
				SYUdatePeriod: new(syndication.SYUpdatePeriod("hourly")),
			},
			want: time.Hour,
		},
		{
			name: "from update period and frequency",
			channel: rdf.Channel{
				SYUdatePeriod:     new(syndication.SYUpdatePeriod("hourly")),
				SYUpdateFrequency: new(2),
			},
			want: time.Hour / 2,
		},
		{
			name:    "default",
			channel: rdf.Channel{},
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
		channel rdf.Channel
		want    *time.Time
	}{
		{
			name: "use dcDate",
			channel: rdf.Channel{
				DcDate: dc.Date{dc.DCDate{Value: testTime}},
			},
			want: &testTime,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.channel.GetPublishedDate()
			assert.WithinDuration(t, *tt.want, *got, time.Minute, "GetUpdatedDate")
		})
	}
}

func TestChannel_GetImage(t *testing.T) {
	tests := []struct {
		name      string
		channel   rdf.Channel
		wantURL   string
		wantTitle string
	}{
		{
			name: "use <media:content>",
			channel: rdf.Channel{
				MediaContent: &media.MediaContent{
					Medium: new(media.MediaContentMediumImage),
					URL:    "http://example.com/image.webp",
				},
			},
			wantURL: "http://example.com/image.webp",
		},
		{
			name: "use <media:thumbnail>",
			channel: rdf.Channel{
				MediaThumbnails: media.MediaThumbnails{
					{URL: "http://example.com/image.webp"},
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
