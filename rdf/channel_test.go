// Copyright 2026 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: 	AGPL-3.0-or-later

package rdf_test

import (
	"testing"
	"time"

	"github.com/immanent-tech/go-syndication/extensions/syndication"
	"github.com/immanent-tech/go-syndication/rdf"
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
