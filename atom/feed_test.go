// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: 	MIT

package atom_test

import (
	"testing"
	"time"

	"github.com/immanent-tech/go-syndication/atom"
	"github.com/immanent-tech/go-syndication/extensions/syndication"
)

func TestFeed_GetUpdateInterval(t *testing.T) {
	testTime := time.Now()
	tests := []struct {
		name string
		feed atom.Feed
		want time.Duration
	}{
		{
			name: "from update period",
			feed: atom.Feed{
				SYUdatePeriod: new(syndication.SYUpdatePeriod("hourly")),
			},
			want: time.Hour,
		},
		{
			name: "from update period and frequency",
			feed: atom.Feed{
				SYUdatePeriod:     new(syndication.SYUpdatePeriod("hourly")),
				SYUpdateFrequency: new(2),
			},
			want: time.Hour / 2,
		},
		{
			name: "from items",
			feed: atom.Feed{
				Entries: []atom.Entry{
					{Published: &atom.DateConstruct{Value: testTime}},
					{Published: &atom.DateConstruct{Value: testTime.Add(2 * time.Hour)}},
					{Published: &atom.DateConstruct{Value: testTime.Add(4 * time.Hour)}},
				},
			},
			want: 2 * time.Hour,
		},
		{
			name: "default",
			feed: atom.Feed{},
			want: time.Hour,
		}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.feed.GetUpdateInterval()
			if got != tt.want {
				t.Errorf("GetUpdateInterval() = %v, want %v", got, tt.want)
			}
		})
	}
}
