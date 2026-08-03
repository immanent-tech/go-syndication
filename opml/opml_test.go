// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: 	MIT

package opml

import (
	"os"
	"testing"

	"github.com/immanent-tech/go-syndication/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testSuite struct {
	wantValid     bool
	wantDecodeErr bool

	tests func(t *testing.T, opml *OPML)
}

var opmlTests = map[string]testSuite{
	"../test/feedvalidator/testcases/opml/clean/subscriptionList.opml": {
		wantValid: true,
		tests: func(t *testing.T, opml *OPML) {
			t.Helper()
			assert.Equal(t, "mySubscriptions.opml", *opml.Head.Title)
			assert.Equal(t, "Sat, 18 Jun 2005 12:11:52 GMT", opml.Head.DateCreated.String())
			assert.Equal(t, "Tue, 02 Aug 2005 21:42:48 GMT", opml.Head.DateModified.String())
			assert.Equal(t, "Dave Winer", *opml.Head.OwnerName)
			assert.Equal(t, "dave@scripting.com", *opml.Head.OwnerEmail)
			item := opml.Body[0]
			assert.Equal(t, "CNET News.com", item.Text)
			description, ok := item.FeedDescription()
			assert.True(t, ok)
			assert.Equal(
				t,
				"Tech news and business reports by CNET News.com. Focused on information technology, core topics include computers, hardware, software, networking, and Internet media.",
				description,
			)
			htmlURL, ok := item.HTMLURL()
			assert.True(t, ok)
			assert.Equal(t, "http://news.com.com/", htmlURL)
			xmlURL, ok := item.XMLURL()
			assert.True(t, ok)
			assert.Equal(t, "http://news.com.com/2547-1_3-0-5.xml", xmlURL)
			lang, ok := item.FeedLanguage()
			assert.True(t, ok)
			assert.Equal(t, "unknown", lang)
			title, ok := item.FeedTitle()
			assert.True(t, ok)
			assert.Equal(t, "CNET News.com", title)
			version, ok := item.FeedVersion()
			assert.True(t, ok)
			assert.Equal(t, "RSS2", version)
			assert.Equal(t, "rss", item.EffectiveType())
		},
	},
	"../test/feedvalidator/testcases/opml/clean/category.opml":                 {wantValid: true},
	"../test/feedvalidator/testcases/opml/clean/directory.opml":                {wantValid: true},
	"../test/feedvalidator/testcases/opml/clean/outlineDocument.opml":          {wantValid: true},
	"../test/feedvalidator/testcases/opml/clean/ownerId.opml":                  {wantValid: true},
	"../test/feedvalidator/testcases/opml/clean/playlist.opml":                 {wantValid: true},
	"../test/feedvalidator/testcases/opml/clean/presentation.opml":             {wantValid: true},
	"../test/feedvalidator/testcases/opml/clean/simpleScript.opml":             {wantValid: true},
	"../test/feedvalidator/testcases/opml/clean/specification.opml":            {wantValid: true},
	"../test/feedvalidator/testcases/opml/clean/states.opml":                   {wantValid: true},
	"../test/feedvalidator/testcases/opml/clean/unknownOutlineType2.opml":      {wantValid: true},
	"../test/feedvalidator/testcases/opml/errors/badlyFormedAttrCreated.opml":  {wantDecodeErr: true},
	"../test/feedvalidator/testcases/opml/errors/badlyFormedDateCreated.opml":  {wantDecodeErr: true},
	"../test/feedvalidator/testcases/opml/errors/badlyFormedDateModified.opml": {wantDecodeErr: true},
	"../test/feedvalidator/testcases/opml/errors/badlyFormedEmail.opml":        {wantValid: false},
	"../test/feedvalidator/testcases/opml/errors/closedbody.opml":              {wantValid: false},
	"../test/feedvalidator/testcases/opml/errors/incorrectOpmlVersion.opml":    {wantValid: false},
	"../test/feedvalidator/testcases/opml/errors/invalidCharacters.opml":       {wantDecodeErr: true},
	"../test/feedvalidator/testcases/opml/errors/invalidIsBreakpoint.opml":     {wantDecodeErr: true},
	"../test/feedvalidator/testcases/opml/errors/invalidIsComment.opml":        {wantDecodeErr: true},
	"../test/feedvalidator/testcases/opml/errors/missingBody.opml":             {wantValid: false},
	"../test/feedvalidator/testcases/opml/errors/missingHead.opml":             {wantValid: false},
	"../test/feedvalidator/testcases/opml/errors/missingOpml.opml":             {wantValid: false},
	"../test/feedvalidator/testcases/opml/errors/missingOpmlVersion.opml":      {wantValid: false},
	"../test/feedvalidator/testcases/opml/errors/missingOutline.opml":          {wantValid: false},
	"../test/feedvalidator/testcases/opml/errors/missingTextAttributes.opml":   {wantValid: false},
	"../test/feedvalidator/testcases/opml/errors/missingTitleAttribute.opml":   {wantValid: false},
	"../test/feedvalidator/testcases/opml/errors/missingUrlAttribute.opml":     {wantValid: false},
	"../test/feedvalidator/testcases/opml/errors/notEncoded.opml":              {wantDecodeErr: true},
	"../test/feedvalidator/testcases/opml/errors/notwellformed.opml":           {wantDecodeErr: true},
	"../test/feedvalidator/testcases/opml/errors/outlinesInWrongPlaces.opml":   {wantValid: false},
	"../test/feedvalidator/testcases/opml/errors/subscriptionListErrors1.opml": {wantValid: false},
	"../test/feedvalidator/testcases/opml/errors/subscriptionListErrors2.opml": {wantValid: false},
	"../test/feedvalidator/testcases/opml/errors/subscriptionListErrors3.opml": {wantValid: false},
	"../test/feedvalidator/testcases/opml/errors/subscriptionListErrors4.opml": {wantValid: false},
	"../test/feedvalidator/testcases/opml/errors/unknownOutlineType.opml":      {wantValid: false},
}

func TestNewOPMLFromBytes(t *testing.T) {
	type args struct {
		data []byte
	}
	tests := []struct {
		name  string
		args  args
		want  *OPML
		suite testSuite
	}{}
	for testFile, testSuites := range opmlTests {
		if data, err := os.ReadFile(testFile); err != nil {
			t.Error("could not read file: " + testFile)
		} else {
			tests = append(tests, struct {
				name  string
				args  args
				want  *OPML
				suite testSuite
			}{
				name:  "file:" + testFile,
				args:  args{data: data},
				suite: testSuites,
			})
		}
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opml, err := NewOPMLFromBytes(tt.args.data)
			if (err != nil) != tt.suite.wantDecodeErr {
				t.Fatalf("Decode() error = %v, wantDecodeErr %v", err, tt.suite.wantDecodeErr)
				return
			}
			if !tt.suite.wantValid {
				require.Error(t, validation.ValidateStruct(opml))
			} else {
				assert.Nil(t, validation.ValidateStruct(opml))
			}
			// Run test suites.
			if tt.suite.tests != nil {
				tt.suite.tests(t, opml)
			}
		})
	}
}

func TestNewOPML(t *testing.T) {
	type args struct {
		options []Option
	}
	tests := []struct {
		name      string
		args      args
		testSuite func(t *testing.T, opml *OPML)
	}{
		{
			name: "valid OPML file",
			args: args{
				[]Option{
					WithTitle("test-subscription"),
					WithOutlines(
						*NewSubscriptionOutline("CNET News.com", "http://news.com.com/2547-1_3-0-5.xml"),
					),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opml := NewOPML(tt.args.options...)
			assert.Equal(t, "test-subscription", *opml.Head.Title)
			feed := opml.Body[0]
			assert.Equal(t, "CNET News.com", feed.Text)
			xmlURL, ok := feed.XMLURL()
			assert.True(t, ok)
			assert.Equal(t, "http://news.com.com/2547-1_3-0-5.xml", xmlURL)
		})
	}
}
