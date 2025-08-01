// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: 	AGPL-3.0-or-later

package feeds

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/immanent-tech/go-syndication/atom"
	"github.com/stretchr/testify/assert"
)

type atomTestSuite struct {
	wantErr bool
	tests   func(t *testing.T, feed *Feed)
}

var atomOtherTests = map[string]atomTestSuite{
	"example.xml": {
		wantErr: false,
		tests: func(t *testing.T, feed *Feed) {
			t.Helper()
			atom := toAtom(t, feed)
			assert.Equal(t, "Example Feed", atom.GetTitle())
			assert.Equal(t, "2003-12-13T18:30:02Z", atom.GetUpdatedDate().Format(time.RFC3339))
			assert.Equal(t, "urn:uuid:60a76c80-d399-11d9-b93C-0003939e0af6", atom.ID.Value)
			assert.Len(t, atom.Entries, 1)
			entry := atom.Entries[0]
			assert.Equal(t, "Atom-Powered Robots Run Amok", entry.GetTitle())
			assert.Equal(t, "http://example.org/2003/12/13/atom03", entry.GetLink())
			assert.Equal(t, "urn:uuid:1225c695-cfb8-4ebb-aaaa-80da344efa6a", entry.GetID())
			assert.Equal(t, "Some text.", entry.GetDescription())
			assert.Equal(t, "2003-12-13T18:30:02Z", entry.GetUpdatedDate().Format(time.RFC3339))
		},
	},
}

var atomTests = map[string]map[string]atomTestSuite{
	"test/assets/atom/other": atomOtherTests,
}

func toAtom(t *testing.T, source *Feed) *atom.Feed {
	t.Helper()
	r, ok := source.FeedSource.(*atom.Feed)
	if !ok {
		t.Fatal("Unable to convert to Atom")
	}
	return r
}

func TestNewFeedFromBytesAtom(t *testing.T) {
	type args struct {
		data []byte
	}
	tests := []struct {
		name  string
		args  args
		want  *Feed
		suite atomTestSuite
	}{}
	for set, testSuites := range atomTests {
		for name, suite := range testSuites {
			testFile := filepath.Join(set, name)
			data, err := os.ReadFile(testFile) // #nosec G304
			if err != nil {
				t.Error("could not read file: " + name)
			} else {
				tests = append(tests, struct {
					name  string
					args  args
					want  *Feed
					suite atomTestSuite
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
			feed, err := NewFeedFromBytes[*atom.Feed](tt.args.data)
			// Check test suite error condition.
			if (err != nil) != tt.suite.wantErr {
				t.Fatalf("NewFeedFromBytes() error = %v, wantErr %v", err, tt.suite.wantErr)
				return
			}
			// Run test suites.
			if tt.suite.tests != nil {
				tt.suite.tests(t, feed)
			}
		})
	}
}
