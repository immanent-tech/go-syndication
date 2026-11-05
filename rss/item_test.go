// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: 	MIT

package rss_test

import (
	"encoding/xml"
	"testing"
	"time"

	"github.com/immanent-tech/go-syndication/extensions/basicgeo"
	"github.com/immanent-tech/go-syndication/extensions/dc"
	"github.com/immanent-tech/go-syndication/extensions/georss"
	"github.com/immanent-tech/go-syndication/extensions/media"
	"github.com/immanent-tech/go-syndication/extensions/slash"
	"github.com/immanent-tech/go-syndication/rss"
	"github.com/immanent-tech/go-syndication/types"
	"github.com/stretchr/testify/assert"
)

func TestItem_GetDescription(t *testing.T) {
	tests := []struct {
		name string
		item rss.Item
		want string
	}{
		{
			name: "using description",
			item: rss.Item{
				Description: rss.NewItemDescription("test description", false),
			},
			want: "test description",
		},
		{
			name: "using <media:group>",
			item: rss.Item{
				MediaGroup: &media.MediaGroup{
					MediaDescription: &media.MediaDescription{
						Value: "test description",
					},
				},
			},
			want: "test description",
		},
		{
			name: "no description",
			item: rss.Item{},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.item.GetDescription()
			if got != tt.want {
				t.Errorf("GetDescription() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestItem_GetAuthors(t *testing.T) {
	tests := []struct {
		name string
		item rss.Item
		want []string
	}{
		{
			name: "with authors",
			item: rss.Item{
				Author: new("author1"),
			},
			want: []string{"author1"},
		},
		{
			name: "multiple sources",
			item: rss.Item{
				Author: new("author1"),
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
			item: rss.Item{
				Author: new("author1"),
				DcPublisher: dc.Publisher{
					"author1",
				},
			},
			want: []string{"author1"},
		},
		{
			name: "no authors",
			item: rss.Item{},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.item.GetAuthors()
			assert.EqualValues(t, tt.want, got)
		})
	}
}

func TestItem_GetContributors(t *testing.T) {
	tests := []struct {
		name string
		item rss.Item
		want []string
	}{
		{
			name: "with contributors",
			item: rss.Item{
				DcContributor: dc.Contributor{
					"contributor1",
				},
			},
			want: []string{"contributor1"},
		},
		{
			name: "no contributors",
			item: rss.Item{},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.item.GetContributors()
			assert.EqualValues(t, tt.want, got)
		})
	}
}

func TestItem_GetRights(t *testing.T) {
	tests := []struct {
		name    string
		item    rss.Item
		want    string
		wantNil bool
	}{
		{
			name: "with rights",
			item: rss.Item{
				DcRights: dc.Rights{
					"rights",
				},
			},
			want: "rights",
		},
		{
			name:    "no rights",
			item:    rss.Item{},
			wantNil: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.item.GetRights()
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

func TestItem_GetLanguage(t *testing.T) {
	tests := []struct {
		name    string
		item    rss.Item
		want    string
		wantNil bool
	}{
		{
			name: "with language",
			item: rss.Item{
				DcLanguage: dc.Language{
					"en-us",
				},
			},
			want: "en-us",
		},
		{
			name:    "no language",
			item:    rss.Item{},
			wantNil: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.item.GetLanguage()
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

func TestItem_GetCategories(t *testing.T) {
	tests := []struct {
		name string
		item rss.Item
		want []string
	}{
		{
			name: "has categories",
			item: rss.Item{
				Categories: []rss.Category{
					{Value: "category1"},
					{Value: "category2"},
				},
			},
			want: []string{"category1", "category2"},
		},
		{
			name: "no duplicates",
			item: rss.Item{
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
			item: rss.Item{
				Categories: []rss.Category{
					{Value: "category1"},
					{Value: "category2"},
				},
				DcSubject: dc.Subject{
					"category3",
				},
				SlashDepartment: new(slash.Department("category4")),
				SlashSection:    new(slash.Section("category5")),
			},
			want: []string{"category1", "category2", "category3", "category4", "category5"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.item.GetCategories()
			assert.EqualValues(t, got, tt.want)
		})
	}
}

func TestItem_GetImage(t *testing.T) {
	tests := []struct {
		name      string
		item      rss.Item
		wantNil   bool
		wantURL   string
		wantTitle string
	}{
		{
			name: "use image",
			item: rss.Item{
				Image: &rss.Image{
					URL:   "http://example.com/image.png",
					Title: "test image",
				},
			},
			wantURL:   "http://example.com/image.png",
			wantTitle: "test image",
		},
		{
			name: "use enclosure",
			item: rss.Item{
				Enclosure: &rss.Enclosure{
					Type: "image/png",
					URL:  "http://example.com/image.png",
				},
			},
			wantURL: "http://example.com/image.png",
		},
		{
			name: "use <media:thumbnail>",
			item: rss.Item{
				MediaThumbnails: []media.MediaThumbnail{
					{URL: "http://example.com/image.png"},
				},
			},
			wantURL:   "http://example.com/image.png",
			wantTitle: "",
		},
		{
			name: "use <media:content>",
			item: rss.Item{
				MediaContent: &media.MediaContent{
					Medium: new(media.MediaContentMediumImage),
					URL:    "http://example.com/image.png",
				},
			},
			wantURL:   "http://example.com/image.png",
			wantTitle: "",
		},
		{
			name:    "no image",
			item:    rss.Item{},
			wantNil: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.item.GetImage()
			if !tt.wantNil {
				assert.Equal(t, tt.wantURL, got.GetURL(), "GetImage URL")
				assert.Equal(t, tt.wantTitle, got.GetTitle(), "GetImage Title")
			} else {
				assert.Nil(t, got, "no image")
			}
		})
	}
}

func TestItem_GetPublishedDate(t *testing.T) {
	testTime := time.Now()
	tests := []struct {
		name    string
		item    rss.Item
		want    *time.Time
		wantNil bool
	}{
		{
			name: "use <published>",
			item: rss.Item{
				PubDate: rss.NewPubDate(testTime),
			},
			want: &testTime,
		},
		{
			name: "use <dc:date>",
			item: rss.Item{
				DcDate: dc.Date{
					{Value: testTime},
				},
			},
			want: &testTime,
		},
		{
			name:    "no date",
			item:    rss.Item{},
			wantNil: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.item.GetPublishedDate()
			if !tt.wantNil {
				assert.WithinDuration(t, *tt.want, *got, time.Minute, "GetUpdatedDate")
			} else {
				assert.Nil(t, got, "no date")
			}
		})
	}
}

func TestItem_HasBasicGeo(t *testing.T) {
	tests := []struct {
		name string
		item rss.Item
		want bool
	}{
		{
			name: "has geo",
			item: rss.Item{
				Lat: new(basicgeo.Latitude(1)),
				Lon: new(basicgeo.Longitude(1)),
			},
			want: true,
		},
		{
			name: "no geo",
			item: rss.Item{},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.item.HasBasicGeo()
			if got != tt.want {
				t.Errorf("HasBasicGeo() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestItem_HasGeoRSS(t *testing.T) {
	tests := []struct {
		name string
		item rss.Item
		want bool
	}{
		{
			name: "has geo",
			item: rss.Item{
				Point: &georss.Point{Lat: 1, Lon: 1},
			},
			want: true,
		},
		{
			name: "no geo",
			item: rss.Item{},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.item.HasGeoRSS()
			if got != tt.want {
				t.Errorf("HasGeoRSS() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestItem_GetGeoInfo(t *testing.T) {
	tests := []struct {
		name    string
		item    rss.Item
		want    *types.GeoInfo
		wantNil bool
	}{
		{
			name: "geo point",
			item: rss.Item{
				Point: &georss.Point{Lat: 1, Lon: 1},
			},
		},
		{
			name:    "no geo",
			item:    rss.Item{},
			wantNil: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.item.GetGeoInfo()
			assert := assert.New(t)
			if !tt.wantNil {
				assert.InEpsilon(1, got.Point.Lat, 0.001)
				assert.InEpsilon(1, got.Point.Lon, 0.001)
			} else {
				assert.Nil(got)
			}
		})
	}
}

func TestItem_GetCommentRSS(t *testing.T) {
	tests := []struct {
		name string
		item rss.Item
		want string
	}{
		{
			name: "has comment RSS",
			item: rss.Item{
				CommentRSS: new("https://example.com/comments/rss"),
			},
			want: "https://example.com/comments/rss",
		},
		{
			name: "has comment RSS (legacy)",
			item: rss.Item{
				CommentRSSLegacy: new("https://example.com/comments/rss"),
			},
			want: "https://example.com/comments/rss",
		},
		{
			name: "no comment RSS",
			item: rss.Item{},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.item.GetCommentRSS()
			if got != tt.want {
				t.Errorf("GetCommentRSS() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestItemDescription_MarshalXML(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		cdata   bool
		want    []byte
		wantErr bool
	}{
		{
			name:    "not cdata",
			value:   "test",
			cdata:   false,
			want:    []byte(`<description>test</description>`),
			wantErr: false,
		},
		{
			name:    "as cdata",
			value:   "<div>test</div>",
			cdata:   true,
			want:    []byte(`<description><![CDATA[<div>test</div>]]></description>`),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := rss.NewItemDescription(tt.value, tt.cdata)
			got, gotErr := xml.Marshal(c)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("MarshalXML() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("MarshalXML() succeeded unexpectedly")
			} else {
				assert.Equal(t, string(tt.want), string(got))
			}
		})
	}
}
