// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: 	MIT

package feeds

import (
	"bytes"
	"encoding/xml"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/immanent-tech/go-syndication/rss"
	"github.com/immanent-tech/go-syndication/validation"
	"github.com/stretchr/testify/assert"
)

type sourceTestSuite struct {
	wantInvalid   bool
	wantDecodeErr bool
	tests         func(t *testing.T, feed *rss.RSS)
}

var sourceMustPass = map[string]sourceTestSuite{
	"example.xml": {
		tests: func(t *testing.T, feed *rss.RSS) {
			assert.Nil(t, validation.ValidateStruct(feed))
			channel := feed.Channel
			assert.Equal(t, "http://scripting.com/rss.xml", *channel.SourceSelf)
			assert.Equal(
				t,
				"https://feedland.social/opml?screenname=davewiner&catname=blogroll",
				*channel.SourceBlogroll,
			)
			assert.Len(t, channel.SourceAccount, 3)
			assert.Equal(t, "bluesky", channel.SourceAccount[0].Service)
			assert.Equal(t, "@scripting.com", channel.SourceAccount[0].Value)
			assert.Equal(t, "Sat, August 1, 2026 6:49 PM EDT", *channel.SourceLocalTime)

			item := channel.Items[0]
			assert.NotNil(t, item.SourceMarkdown)
			assert.Equal(
				t,
				"A user contacted us about a potential security issue in the RSS.chat server. We responded quickly and with v0.6.11 the issue is removed. If you're running rssnetwork.js on a publicly visible server, please install the [new version](https://github.com/scripting/rss.chat/blob/main/server/code/rssnetwork.js) now. Thanks!",
				item.SourceMarkdown.Value,
			)
			assert.NotNil(t, item.SourceOutline)
			assert.Equal(t, "Sat, 01 Aug 2026 22:47:53 GMT", *item.SourceOutline.Created)
			idx := slices.IndexFunc(item.SourceOutline.Attrs, func(a xml.Attr) bool {
				return a.Name.Local == "permalink"
			})
			assert.NotEqual(t, -1, idx)
			assert.Equal(t, "http://scripting.com/2026/08/01.html#a224753", item.SourceOutline.Attrs[idx].Value)
			idx = slices.IndexFunc(item.SourceOutline.Attrs, func(a xml.Attr) bool {
				return a.Name.Local == "flInCalendar"
			})
			assert.NotEqual(t, -1, idx)
			assert.Equal(t, "true", item.SourceOutline.Attrs[idx].Value)
		},
	},
}

var sourceTests = map[string]map[string]sourceTestSuite{
	"test/extensions/source": sourceMustPass,
}

func TestExtensionsSource(t *testing.T) {
	type args struct {
		data []byte
	}
	tests := []struct {
		name  string
		args  args
		want  *rss.RSS
		suite sourceTestSuite
	}{}
	for set, testSuites := range sourceTests {
		for name, suite := range testSuites {
			testFile := filepath.Join(set, name)
			if data, err := os.ReadFile(testFile); err != nil {
				t.Error("could not read file: " + name)
			} else {
				tests = append(tests, struct {
					name  string
					args  args
					want  *rss.RSS
					suite sourceTestSuite
				}{
					name:  "file:" + testFile,
					args:  args{data: data},
					suite: suite,
				})
			}
		}
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			feed, err := Decode[*rss.RSS]("", bytes.NewReader(tt.args.data))
			if (err != nil) != tt.suite.wantDecodeErr {
				t.Fatalf("Decode() error = %v, wantDecodeErr %v", err, tt.suite.wantDecodeErr)
				return
			}

			// Run test suites.
			if tt.suite.tests != nil {
				tt.suite.tests(t, feed)
			}
			// If wantErr, make sure that occurs.
			// if tt.suite.wantInvalid {
			// 	err := feed.Validate()
			// 	if (err != nil) != tt.suite.wantInvalid {
			// 		t.Fatalf("Validate() error = %v, wantErr %v", err, tt.suite.wantInvalid)
			// 		return
			// 	}
			// }
		})
	}
}
