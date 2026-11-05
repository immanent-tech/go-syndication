// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: 	MIT

package syndication_test

import (
	"testing"
	"time"

	"github.com/immanent-tech/go-syndication/extensions/syndication"
	"github.com/stretchr/testify/assert"
)

func TestSYUpdatePeriod_UnmarshalText(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		want    syndication.SYUpdatePeriod
		wantErr bool
	}{
		{
			name: "whitespace cleaned",
			data: []byte(`
				hourly
			`),
			want: syndication.UpdatedHourly,
		},
		{
			name:    "invalid value",
			data:    []byte(`invalid`),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// TODO: construct the receiver type.
			var p syndication.SYUpdatePeriod
			gotErr := p.UnmarshalText(tt.data)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("UnmarshalText() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("UnmarshalText() succeeded unexpectedly")
			}
			assert.Equal(t, tt.want, p)
		})
	}
}

func TestSYUpdatePeriod_Interval(t *testing.T) {
	tests := []struct {
		name string
		freq *syndication.SYUpdateFrequency
		data syndication.SYUpdatePeriod
		want time.Duration
	}{
		{
			name: "hourly",
			freq: nil,
			data: syndication.UpdatedHourly,
			want: time.Hour,
		},
		{
			name: "daily",
			freq: nil,
			data: syndication.UpdatedDaily,
			want: 24 * time.Hour,
		},
		{
			name: "weekly",
			freq: nil,
			data: syndication.UpdatedWeekly,
			want: 7 * 24 * time.Hour,
		},
		{
			name: "yearly",
			freq: nil,
			data: syndication.UpdatedYearly,
			want: 365 * 24 * time.Hour,
		},
		{
			name: "with frequency",
			freq: new(syndication.SYUpdateFrequency(4)),
			data: syndication.UpdatedHourly,
			want: 15 * time.Minute,
		},
		{
			name: "invalid value",
			freq: nil,
			data: syndication.SYUpdatePeriod("invalid"),
			want: 5 * time.Hour,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.data.Interval(tt.freq)
			if got != tt.want {
				t.Errorf("Interval() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSYUpdatePeriod_MarshalText(t *testing.T) {
	tests := []struct {
		name    string
		data    syndication.SYUpdatePeriod
		want    []byte
		wantErr bool
	}{
		{
			name: "ok",
			data: syndication.UpdatedHourly,
			want: []byte(`hourly`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := tt.data.MarshalText()
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("MarshalText() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("MarshalText() succeeded unexpectedly")
			}
			assert.Equal(t, string(tt.want), string(got))
		})
	}
}
