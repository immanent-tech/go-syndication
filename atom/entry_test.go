// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: 	MIT

package atom_test

import (
	"encoding/xml"
	"testing"
	"time"

	"github.com/immanent-tech/go-syndication/atom"
	"github.com/immanent-tech/go-syndication/extensions/dc"
	"github.com/immanent-tech/go-syndication/extensions/media"
	"github.com/immanent-tech/go-syndication/extensions/slash"
	"github.com/stretchr/testify/assert"
)

func TestEntry_GetRights(t *testing.T) {
	tests := []struct {
		name    string
		entry   atom.Entry
		wantNil bool
		want    string
	}{
		{
			name: "use <rights>",
			entry: atom.Entry{
				Rights: &atom.Rights{Value: "test rights"},
			},
			want: "test rights",
		},
		{
			name: "use <dc:rights>",
			entry: atom.Entry{
				DcRights: dc.Rights{
					"test rights",
				},
			},
			want: "test rights",
		},
		{
			name:    "no rights",
			entry:   atom.Entry{},
			wantNil: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.entry.GetRights()
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

func TestEntry_GetLanguage(t *testing.T) {
	tests := []struct {
		name    string
		entry   atom.Entry
		want    string
		wantNil bool
	}{
		{
			name: "use <lang>",
			entry: atom.Entry{
				Lang: new("en"),
			},
			want: "en",
		},
		{
			name: "use <dc:lang>",
			entry: atom.Entry{
				DcLanguage: dc.Language{
					"en",
				},
			},
			want: "en",
		},
		{
			name:    "no lang",
			entry:   atom.Entry{},
			wantNil: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.entry.GetLanguage()
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

func TestEntry_GetCategories(t *testing.T) {
	tests := []struct {
		name  string
		entry atom.Entry
		want  []string
	}{
		{
			name: "has categories",
			entry: atom.Entry{
				Categories: atom.Categories{
					atom.Category{Label: &xml.Attr{Value: "category1"}},
					atom.Category{Label: &xml.Attr{Value: "category2label"}, Term: xml.Attr{Value: "category2"}},
				},
			},
			want: []string{"category1", "category2label"},
		},
		{
			name: "no duplicates",
			entry: atom.Entry{
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
			entry: atom.Entry{
				Categories: atom.Categories{
					atom.Category{Label: &xml.Attr{Value: "category1"}},
					atom.Category{Label: &xml.Attr{Value: "category2label"}, Term: xml.Attr{Value: "category2"}},
				},
				DcSubject: dc.Subject{
					"category3",
				},
				SlashDepartment: new(slash.Department("category4")),
				SlashSection:    new(slash.Section("category5")),
			},
			want: []string{"category1", "category2label", "category3", "category4", "category5"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.entry.GetCategories()
			assert.EqualValues(t, got, tt.want)
		})
	}
}

func TestEntry_GetDescription(t *testing.T) {
	tests := []struct {
		name  string
		entry atom.Entry
		want  string
	}{
		{
			name: "use <summary>",
			entry: atom.Entry{
				Summary: &atom.Summary{Value: "test summary"},
			},
			want: "test summary",
		},
		{
			name: "use <dc:description>",
			entry: atom.Entry{
				DcDescription: dc.Description{
					"test summary",
				},
			},
			want: "test summary",
		},
		{
			name: "use <media:group>",
			entry: atom.Entry{
				MediaGroup: &media.MediaGroup{
					MediaDescription: &media.MediaDescription{
						Value: "test summary",
					},
				},
			},
			want: "test summary",
		},
		{
			name:  "no description",
			entry: atom.Entry{},
			want:  "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.entry.GetDescription()
			if got != tt.want {
				t.Errorf("GetDescription() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEntry_GetImage(t *testing.T) {
	tests := []struct {
		name      string
		entry     atom.Entry
		wantNil   bool
		wantURL   string
		wantTitle string
	}{
		{
			name: "use <media:thumbnail>",
			entry: atom.Entry{
				MediaThumbnails: []media.MediaThumbnail{
					{URL: "http://example.com/image.webp"},
				},
			},
			wantURL:   "http://example.com/image.webp",
			wantTitle: "",
		},
		{
			name: "use <media:group>",
			entry: atom.Entry{
				MediaGroup: &media.MediaGroup{
					MediaThumbnails: []media.MediaThumbnail{
						{URL: "http://example.com/image.webp"},
					},
				},
			},
			wantURL:   "http://example.com/image.webp",
			wantTitle: "",
		},
		{
			name:    "no image",
			entry:   atom.Entry{},
			wantNil: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.entry.GetImage()
			if !tt.wantNil {
				assert.Equal(t, tt.wantURL, got.GetURL(), "GetImage URL")
				assert.Equal(t, tt.wantTitle, got.GetTitle(), "GetImage Title")
			} else {
				assert.Nil(t, got, "no image")
			}
		})
	}
}

func TestEntry_GetPublishedDate(t *testing.T) {
	testTime := time.Now()
	tests := []struct {
		name    string
		entry   atom.Entry
		want    *time.Time
		wantNil bool
	}{
		{
			name: "use <published>",
			entry: atom.Entry{
				Published: &atom.DateConstruct{Value: testTime},
			},
			want: &testTime,
		},
		{
			name: "use <dc:date>",
			entry: atom.Entry{
				DcDate: dc.Date{
					{Value: testTime},
				},
			},
			want: &testTime,
		},
		{
			name:    "no date",
			entry:   atom.Entry{},
			wantNil: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.entry.GetPublishedDate()
			if !tt.wantNil {
				assert.WithinDuration(t, *tt.want, *got, time.Minute, "GetUpdatedDate")
			} else {
				assert.Nil(t, got, "no date")
			}
		})
	}
}

func TestEntry_GetUpdatedDate(t *testing.T) {
	testTime := time.Now()
	tests := []struct {
		name    string
		entry   atom.Entry
		want    *time.Time
		wantNil bool
	}{
		{
			name: "use <updated>",
			entry: atom.Entry{
				Updated: atom.DateConstruct{Value: testTime},
			},
			want: &testTime,
		},
		{
			name: "published fallback",
			entry: atom.Entry{
				Published: &atom.DateConstruct{Value: testTime},
			},
			want: &testTime,
		},
		{
			name:    "no date",
			entry:   atom.Entry{},
			wantNil: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.entry.GetUpdatedDate()
			if !tt.wantNil {
				assert.WithinDuration(t, *tt.want, *got, time.Minute, "GetUpdatedDate")
			} else {
				assert.Nil(t, got, "no date")
			}
		})
	}
}

func TestStandaloneEntry_MarshalXML(t *testing.T) {
	testTime := time.Unix(0, 0)
	tests := []struct {
		name    string
		entry   atom.StandaloneEntry
		want    []byte
		wantErr bool
	}{
		{
			name: "ok",
			entry: atom.StandaloneEntry{
				Title: atom.Title{Value: "Atom-Powered Robots Run Amok"},
				Links: atom.Links{
					atom.Link{Href: "http://example.org/2003/12/13/atom03"},
				},
				ID:      atom.ID{Value: "urn:uuid:1225c695-cfb8-4ebb-aaaa-80da344efa6a"},
				Updated: atom.Updated{Value: testTime},
				Summary: &atom.Summary{Value: "Some text."},
			},
			want: []byte(
				`<entry xmlns="http://www.w3.org/1999/xhtml"><id xmlns="http://www.w3.org/2005/Atom">urn:uuid:1225c695-cfb8-4ebb-aaaa-80da344efa6a</id><link href="http://example.org/2003/12/13/atom03"></link><summary xmlns="http://www.w3.org/2005/Atom" type="text">Some text.</summary><title xmlns="http://www.w3.org/2005/Atom" type="text">Atom-Powered Robots Run Amok</title><updated xmlns="http://www.w3.org/2005/Atom">1970-01-01T10:00:00+10:00</updated></entry>`,
			),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := xml.Marshal(tt.entry)
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

func TestStandaloneEntry_UnmarshalXML(t *testing.T) {
	tests := []struct {
		name    string // description of this test case
		data    []byte
		wantErr bool
	}{
		{
			name: "ok",
			data: []byte(
				`<entry xmlns="http://www.w3.org/1999/xhtml"><id xmlns="http://www.w3.org/2005/Atom">urn:uuid:1225c695-cfb8-4ebb-aaaa-80da344efa6a</id><link href="http://example.org/2003/12/13/atom03"></link><summary xmlns="http://www.w3.org/2005/Atom" type="text">Some text.</summary><title xmlns="http://www.w3.org/2005/Atom" type="text">Atom-Powered Robots Run Amok</title><updated xmlns="http://www.w3.org/2005/Atom">1970-01-01T10:00:00+10:00</updated></entry>`,
			),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var entry atom.StandaloneEntry
			gotErr := xml.Unmarshal(tt.data, &entry)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("UnmarshalXML() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("UnmarshalXML() succeeded unexpectedly")
			}
			assert := assert.New(t)
			assert.Equal("Atom-Powered Robots Run Amok", entry.Title.Value)
			assert.Equal("http://example.org/2003/12/13/atom03", entry.Links[0].Href)
			assert.Equal("urn:uuid:1225c695-cfb8-4ebb-aaaa-80da344efa6a", entry.ID.Value)
			assert.Equal(time.Unix(0, 0), entry.Updated.Value)
			assert.Equal("Some text.", entry.Summary.Value)
		})
	}
}
