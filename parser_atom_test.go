// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: 	MIT

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
			require.NoError(t, validation.ValidateStruct(feed.Entries[0].Authors[0].URI))
			assert.Equal(
				t,
				"http://www.wired.com/news/school/0,1383,54916,00.html",
				feed.Entries[0].Authors[0].URI.Value,
			)
		},
	},
	"entry_author_url_ftp.xml": {
		wantInvalid: false,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			entries := feed.GetItems()
			assert.Len(t, entries, 1)
			require.NoError(t, validation.ValidateStruct(feed.Entries[0].Authors[0]))
			assert.Equal(t, "ftp://example.com/", feed.Entries[0].Authors[0].String())
		},
	},
	"entry_author_url_http.xml": {
		wantInvalid: false,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			entries := feed.GetItems()
			assert.Len(t, entries, 1)
			require.NoError(t, validation.ValidateStruct(feed.Entries[0].Authors[0]))
			assert.Equal(t, "http://example.com/", feed.Entries[0].Authors[0].String())
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
			require.NoError(t, validation.ValidateStruct(content))
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
			require.NoError(t, validation.ValidateStruct(content))
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
			require.NoError(t, validation.ValidateStruct(content))
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
			require.NoError(t, validation.ValidateStruct(content))
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
			require.NoError(t, validation.ValidateStruct(feed.Entries[0].Contributors[0].URI))
			assert.Equal(
				t,
				"http://www.wired.com/news/school/0,1383,54916,00.html",
				feed.Entries[0].Contributors[0].URI.Value,
			)
		},
	},
	"entry_contributor_url_ftp.xml": {
		wantInvalid: false,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			entries := feed.GetItems()
			assert.Len(t, entries, 1)
			require.NoError(t, validation.ValidateStruct(feed.Entries[0].Contributors[0].URI))
			assert.Equal(t, "ftp://example.com/", feed.Entries[0].Contributors[0].URI.Value)
		},
	},
	"entry_contributor_url_http.xml": {
		wantInvalid: false,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			entries := feed.GetItems()
			assert.Len(t, entries, 1)
			require.NoError(t, validation.ValidateStruct(feed.Entries[0].Contributors[0].URI))
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
			failedValidations, err := getFailedValidations(validation.ValidateStruct(feed.Entries[0].ID))
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
			failedValidations, err := getFailedValidations(validation.ValidateStruct(feed.Entries[0].ID))
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
			failedValidations, err := getFailedValidations(validation.ValidateStruct(feed.Entries[0].ID))
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
			failedValidations, err := getFailedValidations(validation.ValidateStruct(feed.Entries[0].ID))
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
			failedValidations, err := getFailedValidations(validation.ValidateStruct(feed.Entries[0].ID))
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
			failedValidations, err := getFailedValidations(validation.ValidateStruct(feed.Entries[0].ID))
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
			failedValidations, err := getFailedValidations(validation.ValidateStruct(feed.Entries[0].ID))
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
	"entry_link_contains_comma.xml": {
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			entries := feed.GetItems()
			assert.Len(t, entries, 1)
			assert.Equal(t, "http://www.wired.com/news/school/0,1383,54916,00.html", feed.Entries[0].Links[0].Href)
		},
	},
	"entry_link_ftp.xml": {
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			entries := feed.GetItems()
			assert.Len(t, entries, 1)
			assert.Equal(t, "ftp://example.com/", feed.Entries[0].Links[0].Href)
		},
	},
	"entry_link_href_missing.xml": {
		wantInvalid: true,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			entries := feed.GetItems()
			assert.Len(t, entries, 1)
			failedValidations, err := getFailedValidations(validation.ValidateStruct(feed.Entries[0].Links[0]))
			require.NoError(t, err)
			assert.Contains(t, failedValidations["Link.Href"], "required")
		},
	},
	"entry_link_http.xml": {
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			entries := feed.GetItems()
			assert.Len(t, entries, 1)
			assert.Equal(t, "http://example.com/", feed.Entries[0].Links[0].Href)
		},
	},
	// TODO: add custom validation for these multiple link tests
	// "entry_link_multiple.xml":
	// "entry_link_multiple2.xml":
	// "entry_link_multiple3.xml":
	// "entry_link_multiple4.xml":
	// "entry_link_multiple5.xml":
	// "entry_link_multiple6.xml":
	// "entry_link_not_multiple.xml":
	// "entry_link_not_multiple2.xml":
	// "entry_link_not_multiple3.xml":
	// TODO: why does an atom link in RSS allow a value and not the atom spec?
	// "entry_link_not_empty.xml":
	"entry_link_rel_alternate.xml": {
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			entries := feed.GetItems()
			assert.Len(t, entries, 1)
			assert.Equal(t, atom.LinkRelAlternate, feed.Entries[0].Links[0].Rel)
		},
	},
	// TODO: how to validate this?
	// "entry_link_rel_blank.xml":
	"entry_link_rel_invalid.xml": {
		wantInvalid: true,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			entries := feed.GetItems()
			assert.Len(t, entries, 1)
			failedValidations, err := getFailedValidations(validation.ValidateStruct(feed.Entries[0].Links[0]))
			require.NoError(t, err)
			assert.Contains(t, failedValidations["Link.Rel"], "oneof")
		},
	},
	"entry_link_rel_invalid2.xml": {
		wantInvalid: true,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			entries := feed.GetItems()
			assert.Len(t, entries, 1)
			failedValidations, err := getFailedValidations(validation.ValidateStruct(feed.Entries[0].Links[0]))
			require.NoError(t, err)
			assert.Contains(t, failedValidations["Link.Rel"], "oneof")
		},
	},
	"entry_link_rel_related.xml": {
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			entries := feed.GetItems()
			assert.Len(t, entries, 1)
			assert.Equal(t, atom.LinkRelRelated, feed.Entries[0].Links[0].Rel)
		},
	},
	"entry_link_rel_via.xml": {
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			entries := feed.GetItems()
			assert.Len(t, entries, 1)
			assert.Equal(t, atom.LinkRelVia, feed.Entries[0].Links[0].Rel)
		},
	},
	// TODO: how to validate this?
	// "entry_link_title_blank.xml":
	"entry_link_title.xml": {
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			entries := feed.GetItems()
			assert.Len(t, entries, 1)
			assert.Equal(t, "pretty much anything is OK here", feed.Entries[0].Links[0].Title)
		},
	},
	// TODO: how to validate this?
	// "entry_link_type_blank.xml":
	"entry_link_type_not_mime.xml": {
		wantInvalid: true,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			entries := feed.GetItems()
			assert.Len(t, entries, 1)
			failedValidations, err := getFailedValidations(validation.ValidateStruct(feed.Entries[0].Links[0]))
			require.NoError(t, err)
			assert.Contains(t, failedValidations["Link.Type"], "mimetype")
		},
	},
	"entry_link_type.xml": {
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			entries := feed.GetItems()
			assert.Len(t, entries, 1)
			assert.Equal(t, "text/html", feed.Entries[0].Links[0].Type)
		},
	},
	"entry_link_type2.xml": {
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			entries := feed.GetItems()
			assert.Len(t, entries, 1)
			assert.Equal(t, "application/xhtml+xml", feed.Entries[0].Links[0].Type)
		},
	},
	"entry_link_type3.xml": {
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			entries := feed.GetItems()
			assert.Len(t, entries, 1)
			assert.Equal(t, "image/jpeg", feed.Entries[0].Links[0].Type)
		},
	},
	"entry_link_type4.xml": {
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			entries := feed.GetItems()
			assert.Len(t, entries, 1)
			assert.Equal(t, "text/plain", feed.Entries[0].Links[0].Type)
		},
	},
	"entry_modified_bad_day.xml": {
		wantDecodeErr: true,
	},
	"entry_modified_bad_day2.xml": {
		wantDecodeErr: true,
	},
	"entry_modified_bad_hours.xml": {
		wantDecodeErr: true,
	},
	"entry_modified_bad_minutes.xml": {
		wantDecodeErr: true,
	},
	"entry_modified_bad_month.xml": {
		wantDecodeErr: true,
	},
	"entry_modified_bad_seconds.xml": {
		wantDecodeErr: true,
	},
	"entry_modified_extra_spaces.xml": {
		wantDecodeErr: true,
	},
	"entry_modified_extra_spaces2.xml": {
		wantDecodeErr: true,
	},
	"entry_modified_extra_spaces3.xml": {
		wantDecodeErr: true,
	},
	"entry_modified_extra_spaces4.xml": {
		wantDecodeErr: true,
	},
	"entry_modified_extra_spaces5.xml": {
		wantDecodeErr: true,
	},
	"entry_modified_fractional_second.xml": {
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			entries := feed.GetItems()
			assert.Len(t, entries, 1)
			assert.Equal(t, "2002-12-31 19:20:30.45 +0100 +0100", feed.Entries[0].GetUpdatedDate().String())
		},
	},
	// TODO: this should fail but it doesn't (probably because types.Datetime excepts more formats than the Atom spec).
	"entry_modified_hours_minutes.xml": {
		wantDecodeErr: true,
	},
	// TODO: might require custom unmarshal logic?
	// "entry_issued_multiple.xml": {
	// 	wantDecodeErr: true,
	// },
	"entry_modified_no_colons.xml": {
		wantDecodeErr: true,
	},
	"entry_modified_no_hyphens.xml": {
		wantDecodeErr: true,
	},
	"entry_modified_no_t.xml": {
		wantDecodeErr: true,
	},
	"entry_modified_no_timezone_colon.xml": {
		wantDecodeErr: true,
	},
	"entry_modified_no_year.xml": {
		wantDecodeErr: true,
	},
	"entry_modified_seconds.xml": {
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			entries := feed.GetItems()
			assert.Len(t, entries, 1)
			assert.Equal(t, "2002-12-31 19:20:30 +0100 +0100", feed.Entries[0].GetUpdatedDate().String())
		},
	},
	"entry_modified_utc.xml": {
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			entries := feed.GetItems()
			assert.Len(t, entries, 1)
			assert.Equal(t, "2002-12-31 19:20:30 +0000 UTC", feed.Entries[0].GetUpdatedDate().String())
		},
	},
	// TODO: this should fail but it doesn't (probably because types.Datetime excepts more formats than the Atom spec).
	"entry_modified_wrong_format.xml": {
		wantDecodeErr: true,
	},
	"entry_modified.xml": {
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			entries := feed.GetItems()
			assert.Len(t, entries, 1)
			assert.Equal(t, "2003-07-01 01:55:07 -0500 -0500", feed.Entries[0].GetUpdatedDate().String())
		},
	},
	"entry_summary_cdata.xml": {
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			entries := feed.GetItems()
			assert.Len(t, entries, 1)
			assert.Equal(t, "Valid summary", feed.Entries[0].GetDescription())
		},
	},
	// TODO: this should fail but doesn't
	"entry_summary_contains_html_cdata.xml": {
		wantInvalid: true,
	},
	"entry_summary_contains_html.xml": {
		wantInvalid: true,
	},
	"entry_summary_is_html.xml": {
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			entries := feed.GetItems()
			assert.Len(t, entries, 1)
			assert.Equal(t, "<b>Bold summary</b>", feed.Entries[0].GetDescription())
		},
	},
	// TODO: this fails as we sanitise the summary field and remove <code> blocks
	"entry_summary_is_inline.xml": {
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			entries := feed.GetItems()
			assert.Len(t, entries, 1)
			assert.Equal(t, "<code><p>foo</p></code>", feed.Entries[0].GetDescription())
		},
	},
	// TODO: this fails as we sanitise the summary field and remove <code> blocks
	"entry_summary_is_inline_2.xml": {
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			entries := feed.GetItems()
			assert.Len(t, entries, 1)
			assert.Equal(t, "<code><p>foo</p></code>", feed.Entries[0].GetDescription())
		},
	},
	// TODO: work out how to validate these
	// "entry_summary_missing.xml":
	// "entry_summary_multiple.xml":
	"entry_summary_no_html.xml": {
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			entries := feed.GetItems()
			assert.Len(t, entries, 1)
			assert.Equal(t, "Valid summary", feed.Entries[0].GetDescription())
		},
	},
	"entry_summary_not_escaped.xml": {
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			entries := feed.GetItems()
			assert.Len(t, entries, 1)
			assert.Equal(t, "<p>foo</p>", feed.Entries[0].GetDescription())
		},
	},
	"entry_summary_not_html_cdata.xml": {
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			entries := feed.GetItems()
			assert.Len(t, entries, 1)
			assert.Equal(t, "Valid summary", feed.Entries[0].GetDescription())
		},
	},
	"entry_summary_not_html.xml": {
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			entries := feed.GetItems()
			assert.Len(t, entries, 1)
			assert.Equal(t, "<a", feed.Entries[0].GetDescription())
		},
	},
	"entry_summary_not_inline_cdata.xml": {
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			entries := feed.GetItems()
			assert.Len(t, entries, 1)
			assert.Equal(
				t,
				"<p>This does not count as inline content, because it's in a CDATA block</p>",
				feed.Entries[0].GetDescription(),
			)
		},
	},
	"entry_summary_not_text_plain.xml": {
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			entries := feed.GetItems()
			assert.Len(t, entries, 1)
			assert.Equal(
				t,
				`So I was reading <a href="http://example.com/" rel="nofollow">example.com</a> the other day, it's really interesting.`,
				feed.Entries[0].GetDescription(),
			)
		},
	},
	"entry_summary_not_text_plain2.xml": {
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			entries := feed.GetItems()
			assert.Len(t, entries, 1)
			assert.Equal(
				t,
				`So I was reading <a href="http://example.com/" rel="nofollow">example.com</a> the other day, it's really interesting.`,
				feed.Entries[0].GetDescription(),
			)
		},
	},
	"entry_summary_not_text_plain3.xml": {
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			entries := feed.GetItems()
			assert.Len(t, entries, 1)
			assert.Equal(
				t,
				`So I was reading <a href="http://example.com/" rel="nofollow">example.com</a> the other day, it's really interesting.`,
				feed.Entries[0].GetDescription(),
			)
		},
	},
	// TODO: work out how to validate this
	// "entry_summary_type_blank.xml":
	"entry_summary_type_not_mime.xml": {
		wantInvalid: true,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			entries := feed.GetItems()
			assert.Len(t, entries, 1)
			failedValidations, err := getFailedValidations(validation.ValidateStruct(feed.Entries[0].Summary))
			require.NoError(t, err)
			assert.Contains(t, failedValidations["Summary.Type"], "mimetype")
		},
	},
	"entry_summary_type.xml": {
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			entries := feed.GetItems()
			assert.Len(t, entries, 1)
			assert.Equal(t, "text/html", feed.Entries[0].Summary.Type)
		},
	},
	"entry_summary_type2.xml": {
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			entries := feed.GetItems()
			assert.Len(t, entries, 1)
			assert.Equal(t, "application/xhtml+xml", feed.Entries[0].Summary.Type)
		},
	},
	"entry_summary_type3.xml": {
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			entries := feed.GetItems()
			assert.Len(t, entries, 1)
			assert.Equal(t, "image/jpeg", feed.Entries[0].Summary.Type)
		},
	},
	"entry_summary_type4.xml": {
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			entries := feed.GetItems()
			assert.Len(t, entries, 1)
			assert.Equal(t, "text/plain", feed.Entries[0].Summary.Type)
		},
	},
	"entry_summary.xml": {
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			entries := feed.GetItems()
			assert.Len(t, entries, 1)
			assert.Equal(t, "Valid summary", feed.Entries[0].GetDescription())
		},
	},
	"entry_title_cdata.xml": {
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			entries := feed.GetItems()
			assert.Len(t, entries, 1)
			assert.Equal(t, "Valid title", feed.Entries[0].GetTitle())
		},
	},
	// TODO: this should fail but doesn't
	"entry_title_contains_html_cdata.xml": {
		wantInvalid: true,
	},
	"entry_title_contains_html.xml": {
		wantInvalid: true,
	},
	"entry_title_is_html.xml": {
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			entries := feed.GetItems()
			assert.Len(t, entries, 1)
			assert.Equal(t, "<b>Bold title</b>", feed.Entries[0].GetTitle())
		},
	},
	// TODO: this fails as we sanitise the summary field and remove <code> blocks
	"entry_title_is_inline.xml": {
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			entries := feed.GetItems()
			assert.Len(t, entries, 1)
			assert.Equal(t, "<code><p>foo</p></code>", feed.Entries[0].GetTitle())
		},
	},
	// TODO: this fails as we sanitise the summary field and remove <code> blocks
	"entry_title_is_inline_2.xml": {
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			entries := feed.GetItems()
			assert.Len(t, entries, 1)
			assert.Equal(t, "<code><p>foo</p></code>", feed.Entries[0].GetTitle())
		},
	},
	// TODO: work out how to validate these
	// "entry_title_missing.xml":
	// "entry_title_multiple.xml":
	"entry_title_no_html.xml": {
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			entries := feed.GetItems()
			assert.Len(t, entries, 1)
			assert.Equal(t, "Valid title", feed.Entries[0].GetTitle())
		},
	},
	"entry_title_not_escaped.xml": {
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			entries := feed.GetItems()
			assert.Len(t, entries, 1)
			assert.Equal(t, "<p>foo</p>", feed.Entries[0].GetTitle())
		},
	},
	"entry_title_not_html_cdata.xml": {
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			entries := feed.GetItems()
			assert.Len(t, entries, 1)
			assert.Equal(t, "Valid title", feed.Entries[0].GetTitle())
		},
	},
	"entry_title_not_html.xml": {
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			entries := feed.GetItems()
			assert.Len(t, entries, 1)
			assert.Equal(t, "<a", feed.Entries[0].GetTitle())
		},
	},
	"entry_title_not_inline_cdata.xml": {
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			entries := feed.GetItems()
			assert.Len(t, entries, 1)
			assert.Equal(
				t,
				"<p>This does not count as inline content, because it's in a CDATA block</p>",
				feed.Entries[0].GetTitle(),
			)
		},
	},
	"entry_title_not_text_plain.xml": {
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			entries := feed.GetItems()
			assert.Len(t, entries, 1)
			assert.Equal(
				t,
				`So I was reading <a href="http://example.com/" rel="nofollow">example.com</a> the other day, it's really interesting.`,
				feed.Entries[0].GetTitle(),
			)
		},
	},
	"entry_title_not_text_plain2.xml": {
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			entries := feed.GetItems()
			assert.Len(t, entries, 1)
			assert.Equal(
				t,
				`So I was reading <a href="http://example.com/" rel="nofollow">example.com</a> the other day, it's really interesting.`,
				feed.Entries[0].GetTitle(),
			)
		},
	},
	// TODO: work out how to validate this
	// "entry_title_type_blank.xml":
	"entry_title_type_not_mime.xml": {
		wantInvalid: true,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			entries := feed.GetItems()
			assert.Len(t, entries, 1)
			failedValidations, err := getFailedValidations(validation.ValidateStruct(feed.Entries[0].Summary))
			require.NoError(t, err)
			assert.Contains(t, failedValidations["Title.Type"], "mimetype")
		},
	},
	"entry_title_type.xml": {
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			entries := feed.GetItems()
			assert.Len(t, entries, 1)
			assert.Equal(t, "text/html", feed.Entries[0].Title.Type)
		},
	},
	"entry_title_type2.xml": {
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			entries := feed.GetItems()
			assert.Len(t, entries, 1)
			assert.Equal(t, "application/xhtml+xml", feed.Entries[0].Title.Type)
		},
	},
	"entry_title_type3.xml": {
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			entries := feed.GetItems()
			assert.Len(t, entries, 1)
			assert.Equal(t, "image/jpeg", feed.Entries[0].Title.Type)
		},
	},
	"entry_title.xml": {
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			entries := feed.GetItems()
			assert.Len(t, entries, 1)
			assert.Equal(t, "Valid title", feed.Entries[0].GetTitle())
		},
	},
	// TODO: unknown elements are discarded, need this test?
	// "entry_unknown_element.xml":
	"feed_author_email_contains_plus.xml": {
		wantInvalid: false,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			assert.Len(t, feed.GetAuthors(), 1)
			assert.Equal(t, "Valid name (valid+folder@example.com)", feed.GetAuthors()[0])
		},
	},
	"feed_author_email_invalid.xml": {
		wantInvalid: true,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			assert.Len(t, feed.GetAuthors(), 1)
			failedValidations, err := getFailedValidations(feed.Authors[0].Validate())
			require.NoError(t, err)
			assert.Contains(t, failedValidations["PersonConstruct.Email.Value"], "email")
		},
	},
	"feed_author_email_overloaded.xml": {
		wantInvalid: true,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			assert.Len(t, feed.GetAuthors(), 1)
			failedValidations, err := getFailedValidations(feed.Authors[0].Validate())
			require.NoError(t, err)
			assert.Contains(t, failedValidations["PersonConstruct.Email.Value"], "email")
		},
	},
	"feed_author_email.xml": {
		wantInvalid: false,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			assert.Len(t, feed.GetAuthors(), 1)
			assert.Equal(t, "Valid name (valid@example.com)", feed.GetAuthors()[0])
		},
	},
	"feed_author_name.xml": {
		wantInvalid: false,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			assert.Len(t, feed.GetAuthors(), 1)
			assert.Equal(t, "Valid name", feed.GetAuthors()[0])
		},
	},
	"feed_author_name_cdata.xml": {
		wantInvalid: false,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			assert.Len(t, feed.GetAuthors(), 1)
			assert.Equal(t, "Valid name", feed.GetAuthors()[0])
		},
	},
	// // TODO: how to test name is NOT html encoded?
	// "feed_author_name_contains_html.xml": {
	// 	wantInvalid: true,
	// },
	// // TODO: how to test name is NOT html encoded?
	// "feed_author_name_contains_html_cdata.xml": {
	// 	wantInvalid: true,
	// },
	// TODO: might require custom unmarshal logic?
	"feed_author_name_multiple.xml": {
		wantInvalid: true,
	},
	"feed_author_unknown_element.xml": {
		wantInvalid: true,
	},
	"feed_author_url_contains_comma.xml": {
		wantInvalid: false,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			assert.Len(t, feed.GetAuthors(), 1)
			require.NoError(t, validation.ValidateStruct(feed.Authors[0].URI))
			assert.Equal(t, "http://www.wired.com/news/school/0,1383,54916,00.html", feed.Authors[0].URI.Value)
		},
	},
	"feed_author_url_ftp.xml": {
		wantInvalid: false,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			assert.Len(t, feed.GetAuthors(), 1)
			require.NoError(t, validation.ValidateStruct(feed.Authors[0].URI))
			assert.Equal(t, "ftp://example.com/", feed.Authors[0].URI.Value)
		},
	},
	"feed_author_url_http.xml": {
		wantInvalid: false,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			assert.Len(t, feed.GetAuthors(), 1)
			require.NoError(t, validation.ValidateStruct(feed.Authors[0].URI))
			assert.Equal(t, "http://example.com/", feed.Authors[0].URI.Value)
		},
	},
	"feed_contributor_email_contains_plus.xml": {
		wantInvalid: false,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			assert.Len(t, feed.GetContributors(), 1)
			assert.Equal(t, "Valid name (valid+folder@example.com)", feed.GetContributors()[0])
		},
	},
	"feed_contributor_email_invalid.xml": {
		wantInvalid: true,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			assert.Len(t, feed.GetContributors(), 1)
			failedValidations, err := getFailedValidations(feed.Contributors[0].Validate())
			require.NoError(t, err)
			assert.Contains(t, failedValidations["PersonConstruct.Email.Value"], "email")
		},
	},
	"feed_contributor_email_overloaded.xml": {
		wantInvalid: true,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			assert.Len(t, feed.GetContributors(), 1)
			failedValidations, err := getFailedValidations(feed.Contributors[0].Validate())
			require.NoError(t, err)
			assert.Contains(t, failedValidations["PersonConstruct.Email.Value"], "email")
		},
	},
	"feed_contributor_email.xml": {
		wantInvalid: false,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			assert.Len(t, feed.GetContributors(), 1)
			assert.Equal(t, "Valid name (valid@example.com)", feed.GetContributors()[0])
		},
	},
	"feed_contributor_name.xml": {
		wantInvalid: false,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			assert.Len(t, feed.GetContributors(), 1)
			assert.Equal(t, "Valid name", feed.GetContributors()[0])
		},
	},
	"feed_contributor_name_blank.xml": {
		wantInvalid: true,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			assert.Len(t, feed.GetContributors(), 1)
			failedValidations, err := getFailedValidations(feed.Contributors[0].Validate())
			require.NoError(t, err)
			assert.Contains(t, failedValidations["PersonConstruct.Name.Value"], "required")
		},
	},
	"feed_contributor_name_cdata.xml": {
		wantInvalid: false,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			assert.Len(t, feed.GetContributors(), 1)
			assert.Equal(t, "Valid name", feed.GetContributors()[0])
		},
	},
	// TODO: how to test name is NOT html encoded?
	"feed_contributor_name_contains_html.xml": {
		wantInvalid: true,
	},
	// TODO: how to test name is NOT html encoded?
	"feed_contributor_name_contains_html_cdata.xml": {
		wantInvalid: true,
	},
	"feed_contributor_name_missing.xml": {
		wantInvalid: true,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			assert.Len(t, feed.GetContributors(), 1)
			failedValidations, err := getFailedValidations(feed.Contributors[0].Validate())
			require.NoError(t, err)
			assert.Contains(t, failedValidations["PersonConstruct.Name.Value"], "required")
		},
	},
	// TODO: might require custom unmarshal logic?
	"feed_contributor_name_multiple.xml": {
		wantInvalid: true,
	},
	"feed_contributor_unknown_element.xml": {
		wantInvalid: true,
	},
	"feed_contributor_url_contains_comma.xml": {
		wantInvalid: false,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			assert.Len(t, feed.GetContributors(), 1)
			require.NoError(t, validation.ValidateStruct(feed.Contributors[0].URI))
			assert.Equal(t, "http://www.wired.com/news/school/0,1383,54916,00.html", feed.Contributors[0].URI.Value)
		},
	},
	"feed_contributor_url_ftp.xml": {
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			assert.Len(t, feed.GetContributors(), 1)
			require.NoError(t, validation.ValidateStruct(feed.Contributors[0].URI))
			assert.Equal(t, "ftp://example.com/", feed.Contributors[0].URI.Value)
		},
	},
	"feed_contributor_url_http.xml": {
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			assert.Len(t, feed.GetContributors(), 1)
			require.NoError(t, validation.ValidateStruct(feed.Contributors[0].URI))
			assert.Equal(t, "http://example.com/", feed.Contributors[0].URI.Value)
		},
	},
	// TODO: might require custom unmarshal logic?
	"feed_contributor_url_multiple.xml": {
		wantInvalid: true,
	},
	"feed_copyright_is_inline_2.xml": {
		wantDecodeErr: true,
	},
	"feed_copyright_is_inline.xml": {
		wantDecodeErr: true,
	},
	// TODO: is this test necessary?
	// "feed_copyright_missing.xml":
	"feed_generator_contains_comma.xml": {
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			assert.Equal(t, "http://www.wired.com/news/school/0,1383,54916,00.html", feed.Generator.URI)
		},
	},
	"feed_generator_name.xml": {
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			assert.Equal(t, "Pretty much any name is acceptable", feed.Generator.Value)
		},
	},
	"feed_generator_not_really_uri.xml": {
		wantInvalid: true,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			failedValidations, err := getFailedValidations(validation.ValidateStruct(feed.Generator))
			require.NoError(t, err)
			assert.Contains(t, failedValidations["Generator.URI"], "url")
		},
	},
	"feed_id_blank.xml": {
		wantInvalid: true,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			failedValidations, err := getFailedValidations(validation.ValidateStruct(feed.ID))
			require.NoError(t, err)
			assert.Contains(t, failedValidations["ID.Value"], "required")
		},
	},
	"feed_id_contains_comma.xml": {
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			assert.Equal(t, "http://www.wired.com/news/school/0,1383,54916,00.html", feed.ID.String())
		},
	},
	"feed_id_full_uri.xml": {
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			assert.Equal(t, "http://example.com/1", feed.ID.String())
		},
	},
	"feed_id_not_full_uri.xml": {
		wantInvalid: true,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			failedValidations, err := getFailedValidations(validation.ValidateStruct(feed.ID))
			require.NoError(t, err)
			assert.Contains(t, failedValidations["ID.Value"], "uri|urn_rfc2141")
		},
	},
	"feed_id_not_urn.xml": {
		wantInvalid: true,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			failedValidations, err := getFailedValidations(validation.ValidateStruct(feed.ID))
			require.NoError(t, err)
			assert.Contains(t, failedValidations["ID.Value"], "uri|urn_rfc2141")
		},
	},
	"feed_id_not_urn2.xml": {
		wantInvalid: true,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			failedValidations, err := getFailedValidations(validation.ValidateStruct(feed.ID))
			require.NoError(t, err)
			assert.Contains(t, failedValidations["ID.Value"], "uri|urn_rfc2141")
		},
	},
	"feed_id_urn.xml": {
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			assert.Equal(t, "urn:diveintomark-org:1", feed.ID.String())
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
			if data, err := os.ReadFile(testFile); err != nil {
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
				if err := feed.Validate(); (err != nil) != tt.suite.wantInvalid {
					t.Fatalf("Validate() error = %v, wantErr %v", err, tt.suite.wantInvalid)
					return
				}
			}
		})
	}
}
