// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: 	MIT

package atom_test

import (
	"encoding/xml"
	"testing"
	"time"

	"github.com/immanent-tech/go-syndication/atom"
	"github.com/immanent-tech/go-syndication/extensions"
	"github.com/immanent-tech/go-syndication/extensions/dc"
	"github.com/immanent-tech/go-syndication/extensions/media"
	"github.com/immanent-tech/go-syndication/extensions/syndication"
	"github.com/immanent-tech/go-syndication/rss"
	"github.com/stretchr/testify/assert"
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

func TestFeed_GetUpdatedDate(t *testing.T) {
	testTime := time.Now()
	tests := []struct {
		name    string
		feed    atom.Feed
		want    *time.Time
		wantNil bool
	}{
		{
			name: "use items",
			feed: atom.Feed{
				Entries: []atom.Entry{
					{Updated: atom.DateConstruct{Value: testTime.Add(-time.Hour)}},
					{Updated: atom.DateConstruct{Value: testTime.Add(-2 * time.Hour)}},
					{Updated: atom.DateConstruct{Value: testTime.Add(-4 * time.Hour)}},
				},
			},
			want: new(testTime.Add(-time.Hour)),
		},
		{
			name: "use <updated>",
			feed: atom.Feed{
				Updated: atom.DateConstruct{Value: testTime},
			},
			want: &testTime,
		},
		{
			name: "use <published>",
			feed: atom.Feed{
				Published: &atom.DateConstruct{Value: testTime},
			},
			want: &testTime,
		},
		{
			name:    "no date",
			feed:    atom.Feed{},
			wantNil: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.feed.GetUpdatedDate()
			if !tt.wantNil {
				assert.WithinDuration(t, *tt.want, *got, time.Minute, "GetUpdatedDate")
			} else {
				assert.Nil(t, got, "no date")
			}
		})
	}
}

func TestFeed_GetPublishedDate(t *testing.T) {
	testTime := time.Now()
	tests := []struct {
		name    string
		feed    atom.Feed
		want    *time.Time
		wantNil bool
	}{
		{
			name: "use <published>",
			feed: atom.Feed{
				Published: &atom.DateConstruct{Value: testTime},
			},
			want: &testTime,
		},
		{
			name: "use <dc:date>",
			feed: atom.Feed{
				DcDate: dc.Date{
					{Value: testTime},
				},
			},
			want: &testTime,
		},
		{
			name:    "no date",
			feed:    atom.Feed{},
			wantNil: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.feed.GetPublishedDate()
			if !tt.wantNil {
				assert.WithinDuration(t, *tt.want, *got, time.Minute, "GetPublishedDate")
			} else {
				assert.Nil(t, got, "no date")
			}
		})
	}
}

func TestFeed_GetImage(t *testing.T) {
	tests := []struct {
		name      string
		feed      atom.Feed
		wantNil   bool
		wantURL   string
		wantTitle string
	}{
		{
			name: "use <logo>",
			feed: atom.Feed{
				Logo: &atom.Logo{
					Value: "http://example.com/image.webp",
				},
			},
			wantURL:   "http://example.com/image.webp",
			wantTitle: "",
		},
		{
			name: "use <icon>",
			feed: atom.Feed{
				Icon: &atom.Icon{
					Value: "http://example.com/image.webp",
				},
			},
			wantURL:   "http://example.com/image.webp",
			wantTitle: "",
		},
		{
			name: "use <media:thumbnail>",
			feed: atom.Feed{
				MediaThumbnails: []media.MediaThumbnail{
					{URL: "http://example.com/image.webp"},
				},
			},
			wantURL:   "http://example.com/image.webp",
			wantTitle: "",
		},
		{
			name:    "no image",
			feed:    atom.Feed{},
			wantNil: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.feed.GetImage()
			if !tt.wantNil {
				assert.Equal(t, tt.wantURL, got.GetURL(), "GetImage URL")
				assert.Equal(t, tt.wantTitle, got.GetTitle(), "GetImage Title")
			} else {
				assert.Nil(t, got, "no image")
			}
		})
	}
}

func TestFeed_GetSourceURL(t *testing.T) {
	tests := []struct {
		name string
		feed atom.Feed
		want string
	}{
		{
			name: "multiple <atom:link>",
			feed: atom.Feed{
				Links: []rss.AtomLink{
					{Rel: new(atom.LinkRelRelated), Href: "http://example.com"},
					{Rel: new(atom.LinkRelSelf), Type: &atom.MimeTypes[0], Href: "http://example.com/rss"},
				},
			},
			want: "http://example.com/rss",
		},
		{
			name: "no rel=self <atom:link>",
			feed: atom.Feed{
				Links: []rss.AtomLink{
					{Rel: new(atom.LinkRelRelated), Href: "http://example.com"},
				},
			},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.feed.GetSourceURL()
			if got != tt.want {
				t.Errorf("GetSourceURL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFeed_GetLink(t *testing.T) {
	tests := []struct {
		name string
		feed atom.Feed
		want string
	}{
		{
			name: "multiple <atom:link> rel=self",
			feed: atom.Feed{
				Links: []rss.AtomLink{
					{Rel: new(atom.LinkRelSelf), Type: &atom.MimeTypes[0], Href: "http://example.com/rss"},
					{Rel: new(atom.LinkRelSelf), Type: new("text/html"), Href: "http://example.com"},
				},
			},
			want: "http://example.com",
		},
		{
			name: "<atom:link> rel=alt",
			feed: atom.Feed{
				Links: []rss.AtomLink{
					{Rel: new(atom.LinkRelSelf), Type: &atom.MimeTypes[0], Href: "http://example.com/rss"},
					{Rel: new(atom.LinkRelAlternate), Type: new("text/html"), Href: "http://example.com"},
				},
			},
			want: "http://example.com",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.feed.GetLink()
			if got != tt.want {
				t.Errorf("GetLink() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFeed_GetDescription(t *testing.T) {
	tests := []struct {
		name string
		feed atom.Feed
		want string
	}{
		{
			name: "use <subtitle>",
			feed: atom.Feed{
				Subtitle: &atom.Subtitle{Value: "test description"},
			},
			want: "test description",
		},
		{
			name: "use <dc:description>",
			feed: atom.Feed{
				DcDescription: dc.Description{
					"test description",
				},
			},
			want: "test description",
		},
		{
			name: "no description",
			feed: atom.Feed{},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.feed.GetDescription()
			if got != tt.want {
				t.Errorf("GetDescription() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFeed_GetRights(t *testing.T) {
	tests := []struct {
		name    string
		feed    atom.Feed
		wantNil bool
		want    string
	}{
		{
			name: "use <rights>",
			feed: atom.Feed{
				Rights: &atom.Rights{Value: "test rights"},
			},
			want: "test rights",
		},
		{
			name: "use <dc:rights>",
			feed: atom.Feed{
				DcRights: dc.Rights{
					"test rights",
				},
			},
			want: "test rights",
		},
		{
			name:    "no rights",
			feed:    atom.Feed{},
			wantNil: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.feed.GetRights()
			if !tt.wantNil {
				assert.NotNil(t, got, "has rights")
				if *got != tt.want {
					t.Errorf("GetRights() = %v, want %v", *got, tt.want)
				}
			} else {
				assert.Nil(t, got, "no rights")
			}
		})
	}
}

func TestFeed_GetLanguage(t *testing.T) {
	tests := []struct {
		name    string
		feed    atom.Feed
		want    string
		wantNil bool
	}{
		{
			name: "use <lang>",
			feed: atom.Feed{
				Lang: new("en"),
			},
			want: "en",
		},
		{
			name: "use <dc:lang>",
			feed: atom.Feed{
				DcLanguage: dc.Language{
					"en",
				},
			},
			want: "en",
		},
		{
			name:    "no lang",
			feed:    atom.Feed{},
			wantNil: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.feed.GetLanguage()
			if !tt.wantNil {
				assert.NotNil(t, got, "has lang")
				if *got != tt.want {
					t.Errorf("GetLanguage() = %v, want %v", *got, tt.want)
				}
			} else {
				assert.Nil(t, got, "no lang")
			}
		})
	}
}

func TestFeed_GetCategories(t *testing.T) {
	tests := []struct {
		name string
		feed atom.Feed
		want []string
	}{
		{
			name: "has categories",
			feed: atom.Feed{
				Categories: atom.Categories{
					atom.Category{Label: &xml.Attr{Value: "category1"}},
					atom.Category{Label: &xml.Attr{Value: "category2label"}, Term: xml.Attr{Value: "category2"}},
				},
			},
			want: []string{"category1", "category2label"},
		},
		{
			name: "no duplicates",
			feed: atom.Feed{
				Categories: atom.Categories{
					atom.Category{Label: &xml.Attr{Value: "category1"}},
					atom.Category{Label: &xml.Attr{Value: "category2label"}, Term: xml.Attr{Value: "category2"}},
					atom.Category{Label: &xml.Attr{Value: "category1"}},
				},
			},
			want: []string{"category1", "category2label"},
		},
		{
			name: "multiple sources",
			feed: atom.Feed{
				Categories: atom.Categories{
					atom.Category{Label: &xml.Attr{Value: "category1"}},
					atom.Category{Label: &xml.Attr{Value: "category2label"}, Term: xml.Attr{Value: "category2"}},
				},
				DcSubject: dc.Subject{
					"category3",
				},
			},
			want: []string{"category1", "category2label", "category3"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.feed.GetCategories()
			assert.EqualValues(t, got, tt.want)
		})
	}
}

func TestFeed_AutoDeclareNamespaces(t *testing.T) {
	tests := []struct {
		name string
		feed *atom.Feed
		want []extensions.Namespace
	}{
		{
			name: "no namespaces",
			feed: &atom.Feed{},
		},
		{
			name: "with declared namespace",
			feed: &atom.Feed{
				Namespaces: []extensions.Namespace{
					{Prefix: "syn", URI: "http://purl.org/rss/1.0/modules/syndication/"},
				},
			},
			want: []extensions.Namespace{
				{Prefix: "syn", URI: "http://purl.org/rss/1.0/modules/syndication/"},
			},
		},
		{
			name: "with undeclared namespace",
			feed: &atom.Feed{
				Extensions: []extensions.Extension{
					{XMLName: xml.Name{Space: "http://purl.org/rss/1.0/modules/syndication/"}},
				},
			},
			want: []extensions.Namespace{
				{Prefix: "syn", URI: "http://purl.org/rss/1.0/modules/syndication/"},
			},
		},
		{
			name: "with entry namespace",
			feed: &atom.Feed{
				Entries: []atom.Entry{
					{
						Extensions: []extensions.Extension{
							{XMLName: xml.Name{Space: "http://purl.org/rss/1.0/modules/syndication/"}},
						},
					},
				},
			},
			want: []extensions.Namespace{
				{Prefix: "syn", URI: "http://purl.org/rss/1.0/modules/syndication/"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.feed.AutoDeclareNamespaces()
			got := tt.feed.Namespaces
			assert.EqualValues(t, tt.want, got)
		})
	}
}

func TestFeed_MarshalXML(t *testing.T) {
	testTime := time.Unix(0, 0).UTC()
	tests := []struct {
		name    string
		feed    atom.Feed
		want    []byte
		wantErr bool
	}{
		{
			name: "ok",
			feed: atom.Feed{
				Title: atom.Title{Value: "Example Feed"},
				Links: atom.Links{
					atom.Link{Href: "http://example.org"},
				},
				Updated: atom.Updated{Value: testTime},
				Authors: atom.Authors{
					atom.PersonConstruct{Name: "John Doe"},
				},
				ID: atom.ID{Value: "urn:uuid:60a76c80-d399-11d9-b93C-0003939e0af6"},
				Entries: []atom.Entry{
					{
						Title: atom.Title{Value: "Atom-Powered Robots Run Amok"},
						Links: atom.Links{
							atom.Link{Href: "http://example.org/2003/12/13/atom03"},
						},
						ID:      atom.ID{Value: "urn:uuid:1225c695-cfb8-4ebb-aaaa-80da344efa6a"},
						Updated: atom.Updated{Value: testTime},
						Summary: &atom.Summary{Value: "Some text."},
					},
				},
			},
			want: []byte(
				`<feed xmlns="http://www.w3.org/1999/xhtml"><id xmlns="http://www.w3.org/2005/Atom">urn:uuid:60a76c80-d399-11d9-b93C-0003939e0af6</id><author xmlns="http://www.w3.org/2005/Atom"><name>John Doe</name></author><entry><id xmlns="http://www.w3.org/2005/Atom">urn:uuid:1225c695-cfb8-4ebb-aaaa-80da344efa6a</id><link href="http://example.org/2003/12/13/atom03"></link><summary xmlns="http://www.w3.org/2005/Atom" type="text">Some text.</summary><title xmlns="http://www.w3.org/2005/Atom" type="text">Atom-Powered Robots Run Amok</title><updated xmlns="http://www.w3.org/2005/Atom">1970-01-01T00:00:00Z</updated></entry><link href="http://example.org"></link><title xmlns="http://www.w3.org/2005/Atom" type="text">Example Feed</title><updated xmlns="http://www.w3.org/2005/Atom">1970-01-01T00:00:00Z</updated></feed>`,
			),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := xml.Marshal(tt.feed)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("MarshalXML() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("MarshalXML() succeeded unexpectedly")
			}
			assert.Equal(t, string(tt.want), string(got))
		})
	}
}
