// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: 	MIT

package feeds

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/immanent-tech/go-syndication/rdf"
	"github.com/stretchr/testify/assert"
)

type rdfTestSuite struct {
	wantInvalid   bool
	wantDecodeErr bool
	tests         func(t *testing.T, feed *rdf.RDF)
}

var rdfMustPass = map[string]rdfTestSuite{
	"rss10_spec_sample_noerror.xml": {
		tests: func(t *testing.T, feed *rdf.RDF) {
			// Channel.
			assert.Equal(t, "http://purl.org/rss/1.0/", feed.DefaultNamespace)
			assert.Equal(t, "http://www.xml.com/xml/news.rss", feed.Channel.About)
			assert.Equal(t, "XML.com", feed.Channel.Title)
			assert.Equal(t, "http://xml.com/pub", feed.Channel.Link)
			// Image.
			assert.Equal(t, "XML.com", feed.Image.Title)
			assert.Equal(t, "http://www.xml.com", feed.Image.Link)
			assert.Equal(t, "http://xml.com/universal/images/xml_tiny.gif", feed.Image.URL)
			assert.Equal(t, "http://xml.com/universal/images/xml_tiny.gif", feed.Channel.Image.Resource)
			// assert.Equal(
			// 	t,
			// 	`      XML.com features a rich mix of information and services \n      for the XML community.\n    `,
			// 	feed.Channel.Description,
			// )
			// Items.
			assert.Len(t, feed.Items, 2)
			assert.Equal(t, "Processing Inclusions with XSLT", feed.Items[0].Title)
			assert.Equal(t, "http://xml.com/pub/2000/08/09/xslt/xslt.html", feed.Items[0].Link)
			assert.Equal(t, "Putting RDF to Work", feed.Items[1].Title)
			assert.Equal(t, "http://xml.com/pub/2000/08/09/rdfdb/index.html", feed.Items[1].Link)
		},
	},
}

var rdfTests = map[string]map[string]rdfTestSuite{
	"test/assets/rss/must": rdfMustPass,
}

func TestNewFeedFromBytesRDF(t *testing.T) {
	type args struct {
		data []byte
	}
	tests := []struct {
		name  string
		args  args
		want  *rdf.RDF
		suite rdfTestSuite
	}{}
	for set, testSuites := range rdfTests {
		for name, suite := range testSuites {
			testFile := filepath.Join(set, name)
			data, err := os.ReadFile(testFile) // #nosec G304
			if err != nil {
				t.Error("could not read file: " + name)
			} else {
				tests = append(tests, struct {
					name  string
					args  args
					want  *rdf.RDF
					suite rdfTestSuite
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
			feed, err := Decode[*rdf.RDF]("", bytes.NewReader(tt.args.data))
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
