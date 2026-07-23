// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: 	MIT

package opml

import (
	"os"
	"slices"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/immanent-tech/go-syndication/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testSuite struct {
	wantValid bool
	tests     func(t *testing.T, opml *OPML)
}

var opmlTests = map[string]testSuite{
	"../test/feedvalidator/testcases/opml/clean/subscriptionList.opml": {
		wantValid: true,
		tests: func(t *testing.T, opml *OPML) {
			t.Helper()
			validate := validator.New()
			for feed := range slices.Values(opml.Body) {
				require.NoError(t, validate.Struct(feed))
			}
			assert.Len(t, opml.Body, 13)
			feed := opml.Body[0]
			assert.Equal(t, "CNET News.com", feed.Text)
			assert.Equal(t, "http://news.com.com/", feed.HTMLURL)
			assert.Equal(t, "http://news.com.com/2547-1_3-0-5.xml", feed.XMLURL)
		},
	},
	"../test/feedvalidator/testcases/opml/errors/subscriptionListErrors1.opml": {
		wantValid: false,
		tests: func(t *testing.T, opml *OPML) {
			t.Helper()
			validate := validator.New()
			require.Error(t, validate.Struct(opml.Body[0]))
			require.Error(t, validate.Struct(opml.Body[1]))
			require.NoError(t, validate.Struct(opml.Body[2]))
			// require.Error(t, validate.Struct(opml.Body[3]))
			// require.Error(t, validate.Struct(opml.Body[4]))
		},
	},
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
		data, err := os.ReadFile(testFile) // #nosec G304
		if err != nil {
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
			if err != nil {
				t.Fatalf("New() error = %v, wantErr %v", err, tt.suite.wantValid)
				return
			}
			if !tt.suite.wantValid {
				assert.Error(t, validation.ValidateStruct(opml))
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
			assert.Equal(t, "test-subscription", opml.Head.Title)
			feed := opml.Body[0]
			assert.Equal(t, "CNET News.com", feed.Text)
			assert.Equal(t, "http://news.com.com/2547-1_3-0-5.xml", feed.XMLURL)
		})
	}
}
