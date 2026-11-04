// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: 	MIT

package rss_test

import (
	"testing"
	"time"

	"github.com/immanent-tech/go-syndication/extensions/syndication"
	"github.com/immanent-tech/go-syndication/rss"
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
