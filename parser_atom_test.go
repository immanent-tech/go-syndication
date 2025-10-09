// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: 	AGPL-3.0-or-later

package feeds

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/immanent-tech/go-syndication/atom"
	"github.com/immanent-tech/go-syndication/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type atomTestSuite struct {
	wantInvalid   bool
	wantDecodeErr bool
	tests         func(t *testing.T, feed *atom.Feed)
}

var atomOtherTests = map[string]atomTestSuite{
	"example.xml": {
		wantInvalid: false,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			assert.Equal(t, "Example Feed", feed.GetTitle())
			assert.Equal(t, "2003-12-13T18:30:02Z", feed.GetUpdatedDate().Format(time.RFC3339))
			assert.Equal(t, "urn:uuid:60a76c80-d399-11d9-b93C-0003939e0af6", feed.ID.Value)
			assert.Len(t, feed.Entries, 1)
			entry := feed.Entries[0]
			assert.Equal(t, "Atom-Powered Robots Run Amok", entry.GetTitle())
			assert.Equal(t, "http://example.org/2003/12/13/atom03", entry.GetLink())
			assert.Equal(t, "urn:uuid:1225c695-cfb8-4ebb-aaaa-80da344efa6a", entry.GetID())
			assert.Equal(t, "Some text.", entry.GetDescription())
			assert.Equal(t, "2003-12-13T18:30:02Z", entry.GetUpdatedDate().Format(time.RFC3339))
		},
	},
}

var atomMustTests = map[string]atomTestSuite{
	"entry_author_email_contains_plus.xml": {
		wantInvalid: false,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			entries := feed.GetItems()
			assert.Len(t, entries, 1)
			assert.Equal(t, "Valid name (valid+folder@example.com)", entries[0].GetAuthors()[0])
		},
	},
	"entry_author_email_invalid.xml": {
		wantInvalid: true,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			entries := feed.GetItems()
			assert.Len(t, entries, 1)
			failedValidations, err := getFailedValidations(feed.Entries[0].Authors[0].Validate())
			require.NoError(t, err)
			assert.Contains(t, failedValidations["PersonConstruct.Email.Value"], "email")
		},
	},
	"entry_author_email_overloaded.xml": {
		wantInvalid: true,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			entries := feed.GetItems()
			assert.Len(t, entries, 1)
			failedValidations, err := getFailedValidations(feed.Entries[0].Authors[0].Validate())
			require.NoError(t, err)
			assert.Contains(t, failedValidations["PersonConstruct.Email.Value"], "email")
		},
	},
	"entry_author_email.xml": {
		wantInvalid: false,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			entries := feed.GetItems()
			assert.Len(t, entries, 1)
			assert.Equal(t, "Valid name (valid@example.com)", entries[0].GetAuthors()[0])
		},
	},
	"entry_author_inherit_from_feed.xml": {
		wantInvalid: false,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			authors := feed.GetAuthors()
			assert.Len(t, authors, 1)
			assert.Equal(t, "Mark Pilgrim", authors[0])
		},
	},
	"entry_author_missing.xml": {
		wantInvalid: true,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			entries := feed.GetItems()
			assert.Len(t, entries, 1)
			failedValidations, err := getFailedValidations(feed.Entries[0].Validate())
			require.NoError(t, err)
			assert.Contains(t, failedValidations["Entry.Authors"], "gt")
		},
	},
	"entry_author_name.xml": {
		wantInvalid: false,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			entries := feed.GetItems()
			assert.Len(t, entries, 1)
			assert.Equal(t, "Valid name", entries[0].GetAuthors()[0])
		},
	},
	"entry_author_name_blank.xml": {
		wantInvalid: true,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			assert.Len(t, feed.GetAuthors(), 1)
			failedValidations, err := getFailedValidations(feed.Authors[0].Validate())
			require.NoError(t, err)
			assert.Contains(t, failedValidations["PersonConstruct.Name.Value"], "required")
		},
	},
	"entry_author_name_cdata.xml": {
		wantInvalid: false,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			entries := feed.GetItems()
			assert.Len(t, entries, 1)
			assert.Equal(t, "Valid name", entries[0].GetAuthors()[0])
		},
	},
	// // TODO: how to test name is NOT html encoded?
	// "entry_author_name_contains_html.xml": {
	// 	wantInvalid: true,
	// },
	// // TODO: how to test name is NOT html encoded?
	// "entry_author_name_contains_html_cdata.xml": {
	// 	wantInvalid: true,
	// },
	"entry_author_name_missing.xml": {
		wantInvalid: true,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			entries := feed.GetItems()
			assert.Len(t, entries, 1)
			failedValidations, err := getFailedValidations(feed.Entries[0].Authors[0].Validate())
			require.NoError(t, err)
			assert.Contains(t, failedValidations["PersonConstruct.Name.Value"], "required")
		},
	},
	// TODO: might require custom unmarshal logic?
	"entry_author_name_multiple.xml": {
		wantInvalid: true,
	},
	"entry_author_unknown_element.xml": {
		wantInvalid: true,
	},
	"entry_author_url_contains_comma.xml": {
		wantInvalid: false,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			entries := feed.GetItems()
			assert.Len(t, entries, 1)
			require.NoError(t, validation.Validate.Struct(feed.Entries[0].Authors[0].URI))
			assert.Equal(t, "http://www.wired.com/news/school/0,1383,54916,00.html", feed.Entries[0].Authors[0].URI.Value)
		},
	},
	"entry_author_url_ftp.xml": {
		wantInvalid: false,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			entries := feed.GetItems()
			assert.Len(t, entries, 1)
			require.NoError(t, validation.Validate.Struct(feed.Entries[0].Authors[0].URI))
			assert.Equal(t, "ftp://example.com/", feed.Entries[0].Authors[0].URI.Value)
		},
	},
	"entry_author_url_http.xml": {
		wantInvalid: false,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			entries := feed.GetItems()
			assert.Len(t, entries, 1)
			require.NoError(t, validation.Validate.Struct(feed.Entries[0].Authors[0].URI))
			assert.Equal(t, "http://example.com/", feed.Entries[0].Authors[0].URI.Value)
		},
	},
	// TODO: might require custom unmarshal logic?
	"entry_author_url_multiple.xml": {
		wantInvalid: true,
	},
	"entry_content_is_html.xml": {
		wantInvalid: false,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			entries := feed.GetItems()
			assert.Len(t, entries, 1)
			content := feed.Entries[0].Content
			require.NoError(t, validation.Validate.Struct(content))
			assert.Equal(t, "\n  <br>\n", content.Value)
		},
	},
	"entry_content_type_blank.xml": {
		wantInvalid: true,
	},
	"entry_content_type_not_mime.xml": {
		wantInvalid: true,
	},
	"entry_content_type2.xml": {
		wantInvalid: false,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			entries := feed.GetItems()
			assert.Len(t, entries, 1)
			content := feed.Entries[0].Content
			require.NoError(t, validation.Validate.Struct(content))
			assert.Equal(t, "application/xhtml+xml", content.Type)
		},
	},
	"entry_content_type3.xml": {
		wantInvalid: false,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			entries := feed.GetItems()
			assert.Len(t, entries, 1)
			content := feed.Entries[0].Content
			require.NoError(t, validation.Validate.Struct(content))
			assert.Equal(t, "image/jpeg", content.Type)
		},
	},
	"entry_content_type4.xml": {
		wantInvalid: false,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			entries := feed.GetItems()
			assert.Len(t, entries, 1)
			content := feed.Entries[0].Content
			require.NoError(t, validation.Validate.Struct(content))
			assert.Equal(t, "text/plain", content.Type)
		},
	},
	"entry_contributor_email_contains_plus.xml": {
		wantInvalid: false,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			entries := feed.GetItems()
			assert.Len(t, entries, 1)
			assert.Equal(t, "Valid name (valid+folder@example.com)", entries[0].GetContributors()[0])
		},
	},
	"entry_contributor_email_invalid.xml": {
		wantInvalid: true,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			entries := feed.GetItems()
			assert.Len(t, entries, 1)
			failedValidations, err := getFailedValidations(feed.Entries[0].Contributors[0].Validate())
			require.NoError(t, err)
			assert.Contains(t, failedValidations["PersonConstruct.Email.Value"], "email")
		},
	},
	"entry_contributor_email_overloaded.xml": {
		wantInvalid: true,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			entries := feed.GetItems()
			assert.Len(t, entries, 1)
			failedValidations, err := getFailedValidations(feed.Entries[0].Contributors[0].Validate())
			require.NoError(t, err)
			assert.Contains(t, failedValidations["PersonConstruct.Email.Value"], "email")
		},
	},
	"entry_contributor_email.xml": {
		wantInvalid: false,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			entries := feed.GetItems()
			assert.Len(t, entries, 1)
			assert.Equal(t, "Valid name (valid@example.com)", entries[0].GetContributors()[0])
		},
	},
	"entry_contributor_name.xml": {
		wantInvalid: false,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			entries := feed.GetItems()
			assert.Len(t, entries, 1)
			assert.Equal(t, "Valid name", entries[0].GetContributors()[0])
		},
	},
	"entry_contributor_name_blank.xml": {
		wantInvalid: true,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			assert.Len(t, feed.Entries[0].GetContributors(), 1)
			failedValidations, err := getFailedValidations(feed.Entries[0].Contributors[0].Validate())
			require.NoError(t, err)
			assert.Contains(t, failedValidations["PersonConstruct.Name.Value"], "required")
		},
	},
	"entry_contributor_name_cdata.xml": {
		wantInvalid: false,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			entries := feed.GetItems()
			assert.Len(t, entries, 1)
			assert.Equal(t, "Valid name", entries[0].GetContributors()[0])
		},
	},
	// TODO: how to test name is NOT html encoded?
	"entry_contributor_name_contains_html.xml": {
		wantInvalid: true,
	},
	// TODO: how to test name is NOT html encoded?
	"entry_contributor_name_contains_html_cdata.xml": {
		wantInvalid: true,
	},
	"entry_contributor_name_missing.xml": {
		wantInvalid: true,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			entries := feed.GetItems()
			assert.Len(t, entries, 1)
			failedValidations, err := getFailedValidations(feed.Entries[0].Contributors[0].Validate())
			require.NoError(t, err)
			assert.Contains(t, failedValidations["PersonConstruct.Name.Value"], "required")
		},
	},
	// TODO: might require custom unmarshal logic?
	"entry_contributor_name_multiple.xml": {
		wantInvalid: true,
	},
	"entry_contributor_unknown_element.xml": {
		wantInvalid: true,
	},
	"entry_contributor_url_contains_comma.xml": {
		wantInvalid: false,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			entries := feed.GetItems()
			assert.Len(t, entries, 1)
			require.NoError(t, validation.Validate.Struct(feed.Entries[0].Contributors[0].URI))
			assert.Equal(t, "http://www.wired.com/news/school/0,1383,54916,00.html", feed.Entries[0].Contributors[0].URI.Value)
		},
	},
	"entry_contributor_url_ftp.xml": {
		wantInvalid: false,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			entries := feed.GetItems()
			assert.Len(t, entries, 1)
			require.NoError(t, validation.Validate.Struct(feed.Entries[0].Contributors[0].URI))
			assert.Equal(t, "ftp://example.com/", feed.Entries[0].Contributors[0].URI.Value)
		},
	},
	"entry_contributor_url_http.xml": {
		wantInvalid: false,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			entries := feed.GetItems()
			assert.Len(t, entries, 1)
			require.NoError(t, validation.Validate.Struct(feed.Entries[0].Contributors[0].URI))
			assert.Equal(t, "http://example.com/", feed.Entries[0].Contributors[0].URI.Value)
		},
	},
	// TODO: might require custom unmarshal logic?
	"entry_contributor_url_multiple.xml": {
		wantInvalid: true,
	},
	"entry_id_blank.xml": {
		wantInvalid: true,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			entries := feed.GetItems()
			assert.Len(t, entries, 1)
			failedValidations, err := getFailedValidations(validation.Validate.Struct(feed.Entries[0].ID))
			require.NoError(t, err)
			assert.Contains(t, failedValidations["ID.Value"], "required")
		},
	},
	"entry_id_contains_comma.xml": {
		wantInvalid: false,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			entries := feed.GetItems()
			assert.Len(t, entries, 1)
			assert.Equal(t, "http://www.wired.com/news/school/0,1383,54916,00.html", feed.Entries[0].GetID())
		},
	},
	// TODO: might require custom unmarshal logic?
	"entry_id_duplicate_value.xml": {
		wantInvalid: true,
	},
	"entry_id_full_uri.xml": {
		wantInvalid: false,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			entries := feed.GetItems()
			assert.Len(t, entries, 1)
			assert.Equal(t, "http://example.com/1", feed.Entries[0].GetID())
		},
	},
	"entry_id_missing.xml": {
		wantInvalid: true,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			entries := feed.GetItems()
			assert.Len(t, entries, 1)
			failedValidations, err := getFailedValidations(validation.Validate.Struct(feed.Entries[0].ID))
			require.NoError(t, err)
			assert.Contains(t, failedValidations["ID.Value"], "required")
		},
	},
	// TODO: might require custom unmarshal logic?
	"entry_id_multiple.xml": {
		wantInvalid: true,
	},
	"entry_id_not_full_uri.xml": {
		wantInvalid: true,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			entries := feed.GetItems()
			assert.Len(t, entries, 1)
			failedValidations, err := getFailedValidations(validation.Validate.Struct(feed.Entries[0].ID))
			require.NoError(t, err)
			assert.Contains(t, failedValidations["ID.Value"], "uri|urn_rfc2141|uuid")
		},
	},
	"entry_id_not_tag.xml": {
		wantInvalid: true,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			entries := feed.GetItems()
			assert.Len(t, entries, 1)
			failedValidations, err := getFailedValidations(validation.Validate.Struct(feed.Entries[0].ID))
			require.NoError(t, err)
			assert.Contains(t, failedValidations["ID.Value"], "uri|urn_rfc2141|uuid")
		},
	},
	"entry_id_not_tag2.xml": {
		wantInvalid: true,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			entries := feed.GetItems()
			assert.Len(t, entries, 1)
			failedValidations, err := getFailedValidations(validation.Validate.Struct(feed.Entries[0].ID))
			require.NoError(t, err)
			assert.Contains(t, failedValidations["ID.Value"], "uri|urn_rfc2141|uuid")
		},
	},
	"entry_id_not_urn.xml": {
		wantInvalid: true,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			entries := feed.GetItems()
			assert.Len(t, entries, 1)
			failedValidations, err := getFailedValidations(validation.Validate.Struct(feed.Entries[0].ID))
			require.NoError(t, err)
			assert.Contains(t, failedValidations["ID.Value"], "uri|urn_rfc2141|uuid")
		},
	},
	"entry_id_not_urn2.xml": {
		wantInvalid: true,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			entries := feed.GetItems()
			assert.Len(t, entries, 1)
			failedValidations, err := getFailedValidations(validation.Validate.Struct(feed.Entries[0].ID))
			require.NoError(t, err)
			assert.Contains(t, failedValidations["ID.Value"], "uri|urn_rfc2141|uuid")
		},
	},
	"entry_id_urn.xml": {
		wantInvalid: false,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			entries := feed.GetItems()
			assert.Len(t, entries, 1)
			assert.Equal(t, "urn:diveintomark-org:1", feed.Entries[0].GetID())
		},
	},
	"entry_id_tag.xml": {
		wantInvalid: false,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			entries := feed.GetItems()
			assert.Len(t, entries, 1)
			assert.Equal(t, "tag:diveintomark.org,2003:blog-14:post-19", feed.Entries[0].GetID())
		},
	},
	"entry_issued_bad_day.xml": {
		wantDecodeErr: true,
	},
	"entry_issued_bad_day2.xml": {
		wantDecodeErr: true,
	},
	"entry_issued_bad_hours.xml": {
		wantDecodeErr: true,
	},
	"entry_issued_bad_minutes.xml": {
		wantDecodeErr: true,
	},
	"entry_issued_bad_month.xml": {
		wantDecodeErr: true,
	},
	"entry_issued_bad_seconds.xml": {
		wantDecodeErr: true,
	},
	"entry_issued_extra_spaces.xml": {
		wantDecodeErr: true,
	},
	"entry_issued_extra_spaces2.xml": {
		wantDecodeErr: true,
	},
	"entry_issued_extra_spaces3.xml": {
		wantDecodeErr: true,
	},
	"entry_issued_extra_spaces4.xml": {
		wantDecodeErr: true,
	},
	"entry_issued_extra_spaces5.xml": {
		wantDecodeErr: true,
	},
	"entry_issued_fractional_second.xml": {
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			entries := feed.GetItems()
			assert.Len(t, entries, 1)
			assert.Equal(t, "2002-12-31 19:20:30.45 +0100 +0100", feed.Entries[0].GetPublishedDate().String())
		},
	},
	// TODO: this should fail but it doesn't (probably because types.Datetime excepts more formats than the Atom spec).
	"entry_issued_hours_minutes.xml": {
		wantDecodeErr: true,
	},
	// TODO: might require custom unmarshal logic?
	// "entry_issued_multiple.xml": {
	// 	wantDecodeErr: true,
	// },
	"entry_issued_no_colons.xml": {
		wantDecodeErr: true,
	},
	"entry_issued_no_hyphens.xml": {
		wantDecodeErr: true,
	},
	"entry_issued_no_t.xml": {
		wantDecodeErr: true,
	},
	"entry_issued_no_timezone_colon.xml": {
		wantDecodeErr: true,
	},
	"entry_issued_no_year.xml": {
		wantDecodeErr: true,
	},
	"entry_issued_seconds.xml": {
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			entries := feed.GetItems()
			assert.Len(t, entries, 1)
			assert.Equal(t, "2002-12-31 19:20:30 +0100 +0100", feed.Entries[0].GetPublishedDate().String())
		},
	},
	"entry_issued_utc.xml": {
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			entries := feed.GetItems()
			assert.Len(t, entries, 1)
			assert.Equal(t, "2002-12-31 19:20:30 +0000 UTC", feed.Entries[0].GetPublishedDate().String())
		},
	},
	// TODO: this should fail but it doesn't (probably because types.Datetime excepts more formats than the Atom spec).
	"entry_issued_wrong_format.xml": {
		wantDecodeErr: true,
	},
	"entry_issued.xml": {
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			entries := feed.GetItems()
			assert.Len(t, entries, 1)
			assert.Equal(t, "2003-07-01 01:55:07 -0500 -0500", feed.Entries[0].GetPublishedDate().String())
		},
	},
}

var atomTests = map[string]map[string]atomTestSuite{
	"test/assets/atom/other": atomOtherTests,
	"test/assets/atom/must":  atomMustTests,
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
			feed, err := Decode[*atom.Feed]("", tt.args.data)
			if (err != nil) != tt.suite.wantDecodeErr {
				t.Fatalf("Decode() error = %v, wantDecodeErr %v", err, tt.suite.wantDecodeErr)
				return
			}

			// Run test suites.
			if tt.suite.tests != nil {
				tt.suite.tests(t, feed)
			}
			// If wantErr, make sure that occurs.
			if tt.suite.wantInvalid {
				err := feed.Validate()
				if (err != nil) != tt.suite.wantInvalid {
					t.Fatalf("Validate() error = %v, wantErr %v", err, tt.suite.wantInvalid)
					return
				}
			}
		})
	}
}
