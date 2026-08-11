// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: 	MIT

package feeds

import (
	"bytes"
	"os"
	"path/filepath"
	"slices"
	"testing"
	"time"

	"github.com/immanent-tech/go-syndication/atom"
	"github.com/immanent-tech/go-syndication/extensions"
	"github.com/immanent-tech/go-syndication/extensions/dc"
	"github.com/immanent-tech/go-syndication/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// atomTestSuite contains data for running a single test suite for the Atom format. The test suite can specify whether
// decoding should fail, validation should fail and/or define a custom test helper to run custom tests and validation of
// the atom object.
type atomTestSuite struct {
	wantInvalid   bool
	wantDecodeErr bool
	tests         func(t *testing.T, feed *atom.Feed)
}

var atomOtherTests = map[string]atomTestSuite{
	"brief-noerror.xml": {
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			assert.Equal(t, "Example Feed", feed.GetTitle())
			assert.Equal(t, "http://example.org/", feed.GetLink())
			assert.Equal(t, "2003-12-13 18:30:02 +0000 UTC", feed.GetUpdatedDate().String())
			assert.Equal(t, "John Doe", feed.GetAuthors()[0])
			assert.Equal(t, "urn:uuid:60a76c80-d399-11d9-b93C-0003939e0af6", feed.ID.String())
			assert.Len(t, feed.GetItems(), 1)
			item := feed.Entries[0]
			assert.Equal(t, "Atom-Powered Robots Run Amok", item.GetTitle())
			assert.Equal(t, "http://example.org/2003/12/13/atom03", item.GetLink())
			assert.Equal(t, "urn:uuid:1225c695-cfb8-4ebb-aaaa-80da344efa6a", item.GetID())
			assert.Equal(t, "2003-12-13 18:30:02 +0000 UTC", item.GetUpdatedDate().String())
			assert.Equal(t, "Some text.", item.Summary.String())
		},
	},
	"extensive-noerror.xml": {
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			assert.Equal(t, "dive into mark", feed.GetTitle())
			assert.Equal(t, atom.TextConstructTypeText, *feed.Title.Type)
			assert.Equal(
				t,
				"A <em>lot</em> of effort\n    went into making this effortless",
				feed.Subtitle.String(),
			)
			assert.Equal(t, atom.TextConstructTypeHtml, *feed.Subtitle.Type)
			assert.Equal(t, "2005-07-11 12:29:29 +0000 UTC", feed.GetUpdatedDate().String())
			assert.Equal(t, "tag:example.org,2003:3", feed.ID.String())
			assert.Len(t, feed.Links, 2)
			assert.Equal(t, "http://example.org/", feed.Links[0].String())
			assert.Equal(t, atom.LinkRelAlternate, *feed.Links[0].Rel)
			assert.Equal(t, "text/html", *feed.Links[0].Type)
			assert.Equal(t, "en", *feed.Links[0].HrefLang)
			assert.Equal(t, "http://example.org/feed.atom", feed.Links[1].String())
			assert.Equal(t, atom.LinkRelSelf, *feed.Links[1].Rel)
			assert.Equal(t, "application/atom+xml", *feed.Links[1].Type)
			assert.Equal(t, "Copyright (c) 2003, Mark Pilgrim", feed.Rights.String())
			assert.Equal(t, "Example Toolkit/1.0 (http://www.example.com/)", feed.Generator.String())
			assert.Len(t, feed.GetItems(), 1)
			item := feed.Entries[0]
			assert.Equal(t, "Atom draft-07 snapshot", item.GetTitle())
			assert.Len(t, item.Links, 2)
			assert.Equal(t, "http://example.org/2005/04/02/atom", item.Links[0].String())
			assert.Equal(t, atom.LinkRelAlternate, *item.Links[0].Rel)
			assert.Equal(t, "text/html", *item.Links[0].Type)
			assert.Equal(t, "http://example.org/audio/ph34r_my_podcast.mp3", item.Links[1].String())
			assert.Equal(t, atom.LinkRelEnclosure, *item.Links[1].Rel)
			assert.Equal(t, "audio/mpeg", *item.Links[1].Type)
			assert.Equal(t, 1337, *item.Links[1].Length)
			assert.Equal(t, "tag:example.org,2003:3.2397", item.GetID())
			assert.Equal(t, "2005-07-11 12:29:29 +0000 UTC", item.GetUpdatedDate().String())
			assert.Equal(t, "2003-12-13 08:29:29 -0400 -0400", item.GetPublishedDate().String())
			assert.Len(t, item.GetAuthors(), 1)
			assert.Equal(t, "Mark Pilgrim (f8dy@example.com) http://example.org/", item.GetAuthors()[0])
			assert.Len(t, item.GetContributors(), 2)
			assert.Equal(t, "Sam Ruby", item.GetContributors()[0])
			assert.Equal(t, "Joe Gregorio", item.GetContributors()[1])
			assert.Equal(t, "<p><i>[Update: The Atom draft is finished.]</i></p>", *item.GetContent())
			assert.Equal(t, atom.ContentTypeXhtml, *item.Content.Type)
			assert.Equal(t, "en", *item.Content.Lang)
			assert.Equal(t, "http://diveintomark.org/", *item.Content.Base)
		},
	},
	"extensive-nowarn.xml": {
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			t.Skip("same as extensive-noerror.xml")
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
			failedValidations, err := getFailedValidations(validation.ValidateStruct(feed.Entries[0].Authors[0]))
			require.NoError(t, err)
			assert.Contains(t, failedValidations["PersonConstruct.Email"], "email")
		},
	},
	"entry_author_email_overloaded.xml": {
		wantInvalid: true,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			entries := feed.GetItems()
			assert.Len(t, entries, 1)
			failedValidations, err := getFailedValidations(validation.ValidateStruct(feed.Entries[0].Authors[0]))
			require.NoError(t, err)
			assert.Contains(t, failedValidations["PersonConstruct.Email"], "email")
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
			failedValidations, err := getFailedValidations(validation.ValidateStruct(feed.Authors[0]))
			require.NoError(t, err)
			assert.Contains(t, failedValidations["PersonConstruct.Name"], "required")
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
	"entry_author_name_contains_html.xml": {
		wantInvalid: true,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			failedValidations, err := getFailedValidations(validation.ValidateStruct(feed))
			require.NoError(t, err)
			assert.Contains(t, failedValidations["Feed.Authors"], "gt")
		},
	},
	"entry_author_name_contains_html_cdata.xml": {
		wantInvalid: true,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			failedValidations, err := getFailedValidations(validation.ValidateStruct(feed))
			require.NoError(t, err)
			assert.Contains(t, failedValidations["Feed.Authors"], "gt")
		},
	},
	"entry_author_name_missing.xml": {
		wantInvalid: true,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			entries := feed.GetItems()
			assert.Len(t, entries, 1)
			failedValidations, err := getFailedValidations(validation.ValidateStruct(feed.Entries[0].Authors[0]))
			require.NoError(t, err)
			assert.Contains(t, failedValidations["PersonConstruct.Name"], "required")
		},
	},
	"entry_author_name_multiple.xml": {
		tests: func(t *testing.T, _ *atom.Feed) {
			t.Helper()
			t.Skip("Go's XML parser won't allow multiple.")
		},
	},
	"entry_author_unknown_element.xml": {
		tests: func(t *testing.T, _ *atom.Feed) {
			t.Helper()
			t.Skip("Unknown elements are ignored by Go's parser.")
		},
	},
	"entry_author_url_contains_comma.xml": {
		wantInvalid: false,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			entries := feed.GetItems()
			assert.Len(t, entries, 1)
			assert.Equal(
				t,
				"http://www.wired.com/news/school/0,1383,54916,00.html",
				*feed.Entries[0].Authors[0].URI,
			)
		},
	},
	"entry_author_url_ftp.xml": {
		wantInvalid: false,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			entries := feed.GetItems()
			assert.Len(t, entries, 1)
			assert.Equal(t, "ftp://example.com/", *feed.Entries[0].Authors[0].URI)
		},
	},
	"entry_author_url_http.xml": {
		wantInvalid: false,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			entries := feed.GetItems()
			assert.Len(t, entries, 1)
			assert.Equal(t, "http://example.com/", *feed.Entries[0].Authors[0].URI)
		},
	},
	"entry_author_url_multiple.xml": {
		tests: func(t *testing.T, _ *atom.Feed) {
			t.Helper()
			t.Skip("Go's XML parser won't allow multiple.")
		},
	},
	"entry_content_is_html.xml": {
		wantInvalid: false,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			entries := feed.GetItems()
			assert.Len(t, entries, 1)
			assert.NotNil(t, feed.Entries[0].Content)
			assert.Equal(t, "\n  <br>\n", feed.Entries[0].Content.String())
		},
	},
	"entry_content_is_inline.xml": {
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			entries := feed.GetItems()
			assert.Len(t, entries, 1)
			assert.NotNil(t, feed.Entries[0].Content)
			assert.Equal(t, "<code>&lt;p&gt;foo&lt;/p&gt;</code>", feed.Entries[0].Content.String())
		},
	},
	"entry_content_is_inline_2.xml": {
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			entries := feed.GetItems()
			assert.Len(t, entries, 1)
			assert.NotNil(t, feed.Entries[0].Content)
			assert.Equal(t, "<code>&lt;p&gt;foo&lt;/p&gt;</code>", feed.Entries[0].Content.String())
		},
	},
	"entry_content_type_blank.xml": {
		wantInvalid: true,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			require.Error(t, feed.Entries[0].Content.Validate())
		},
	},
	"entry_content_type_not_mime.xml": {
		tests: func(t *testing.T, _ *atom.Feed) {
			t.Helper()
			t.Skip("requires additional logic to verify a valid registered mimetype.")
		},
	},
	"entry_content_not_escaped.xml": {
		wantInvalid: true,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			require.Error(t, feed.Entries[0].Content.Validate())
		},
	},
	"entry_content_not_html.xml": {
		wantInvalid: true,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			require.Error(t, feed.Entries[0].Content.Validate())
		},
	},
	"entry_content_not_inline.xml": {
		wantInvalid: true,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			require.Error(t, feed.Entries[0].Content.Validate())
		},
	},
	"entry_content_not_inline_cdata.xml": {
		wantInvalid: true,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			require.Error(t, feed.Entries[0].Content.Validate())
		},
	},
	"entry_content_not_text_plain.xml": {
		wantInvalid: true,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			require.Error(t, feed.Entries[0].Content.Validate())
		},
	},
	"entry_content_not_text_plain_2.xml": {
		wantInvalid: false,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			require.Error(t, feed.Entries[0].Content.Validate())
		},
	},
	"entry_content_type.xml": {
		wantInvalid: false,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			entries := feed.GetItems()
			assert.Len(t, entries, 1)
			content := feed.Entries[0].Content
			assert.Equal(t, "text/html", string(*content.Type))
		},
	},
	"entry_content_type2.xml": {
		wantInvalid: false,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			entries := feed.GetItems()
			assert.Len(t, entries, 1)
			content := feed.Entries[0].Content
			assert.Equal(t, "application/xhtml+xml", string(*content.Type))
		},
	},
	"entry_content_type3.xml": {
		wantInvalid: false,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			entries := feed.GetItems()
			assert.Len(t, entries, 1)
			content := feed.Entries[0].Content
			assert.Equal(t, "image/jpeg", string(*content.Type))
		},
	},
	"entry_content_type4.xml": {
		wantInvalid: false,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			entries := feed.GetItems()
			assert.Len(t, entries, 1)
			content := feed.Entries[0].Content
			assert.Equal(t, "text/plain", string(*content.Type))
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
			failedValidations, err := getFailedValidations(validation.ValidateStruct(feed.Entries[0].Contributors[0]))
			require.NoError(t, err)
			assert.Contains(t, failedValidations["PersonConstruct.Email"], "email")
		},
	},
	"entry_contributor_email_overloaded.xml": {
		wantInvalid: true,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			entries := feed.GetItems()
			assert.Len(t, entries, 1)
			failedValidations, err := getFailedValidations(validation.ValidateStruct(feed.Entries[0].Contributors[0]))
			require.NoError(t, err)
			assert.Contains(t, failedValidations["PersonConstruct.Email"], "email")
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
			failedValidations, err := getFailedValidations(validation.ValidateStruct(feed.Entries[0].Contributors[0]))
			require.NoError(t, err)
			assert.Contains(t, failedValidations["PersonConstruct.Name"], "required")
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
	"entry_contributor_name_contains_html.xml": {
		wantInvalid: true,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			assert.Len(t, feed.Entries[0].GetContributors(), 1)
			assert.Error(t, feed.Entries[0].Contributors[0].Validate())
		},
	},
	"entry_contributor_name_contains_html_cdata.xml": {
		wantInvalid: true,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			assert.Len(t, feed.Entries[0].GetContributors(), 1)
			assert.Error(t, feed.Entries[0].Contributors[0].Validate())
		},
	},
	"entry_contributor_name_missing.xml": {
		wantInvalid: true,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			entries := feed.GetItems()
			assert.Len(t, entries, 1)
			failedValidations, err := getFailedValidations(validation.ValidateStruct(feed.Entries[0].Contributors[0]))
			require.NoError(t, err)
			assert.Contains(t, failedValidations["PersonConstruct.Name"], "required")
		},
	},
	"entry_contributor_name_multiple.xml": {
		tests: func(t *testing.T, _ *atom.Feed) {
			t.Helper()
			t.Skip("Go's XML parser won't allow multiple.")
		},
	},
	"entry_contributor_unknown_element.xml": {
		tests: func(t *testing.T, _ *atom.Feed) {
			t.Helper()
			t.Skip("Unknown elements are ignored by Go's parser.")
		},
	},
	"entry_contributor_url_contains_comma.xml": {
		wantInvalid: false,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			entries := feed.GetItems()
			assert.Len(t, entries, 1)
			assert.Equal(
				t,
				"http://www.wired.com/news/school/0,1383,54916,00.html",
				*feed.Entries[0].Contributors[0].URI,
			)
		},
	},
	"entry_contributor_url_ftp.xml": {
		wantInvalid: false,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			entries := feed.GetItems()
			assert.Len(t, entries, 1)
			assert.Equal(t, "ftp://example.com/", *feed.Entries[0].Contributors[0].URI)
		},
	},
	"entry_contributor_url_http.xml": {
		wantInvalid: false,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			entries := feed.GetItems()
			assert.Len(t, entries, 1)
			assert.Equal(t, "http://example.com/", *feed.Entries[0].Contributors[0].URI)
		},
	},
	"entry_contributor_url_multiple.xml": {
		tests: func(t *testing.T, _ *atom.Feed) {
			t.Helper()
			t.Skip("Go's XML parser won't allow multiple.")
		},
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
	"entry_id_duplicate_value.xml": {
		wantInvalid: true,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			entries := feed.GetItems()
			assert.Len(t, entries, 2)
			assert.Error(t, feed.Validate())
		},
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
	"entry_id_multiple.xml": {
		tests: func(t *testing.T, _ *atom.Feed) {
			t.Helper()
			t.Skip("Go's XML parser won't allow multiple.")
		},
	},
	"entry_id_not_full_uri.xml": {
		wantInvalid: true,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			entries := feed.GetItems()
			assert.Len(t, entries, 1)
			failedValidations, err := getFailedValidations(validation.ValidateStruct(feed.Entries[0].ID))
			require.NoError(t, err)
			assert.Contains(t, failedValidations["ID.Value"], "absolute_uri|urn_rfc2141|uuid")
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
			assert.Contains(t, failedValidations["ID.Value"], "absolute_uri|urn_rfc2141|uuid")
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
			assert.Contains(t, failedValidations["ID.Value"], "absolute_uri|urn_rfc2141|uuid")
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
			assert.Contains(t, failedValidations["ID.Value"], "absolute_uri|urn_rfc2141|uuid")
		},
	},
	"entry_id_not_uuid.xml": {
		tests: func(t *testing.T, _ *atom.Feed) {
			t.Helper()
			t.Skip("requires additional logic")
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
			assert.Contains(t, failedValidations["ID.Value"], "absolute_uri|urn_rfc2141|uuid")
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
	"entry_id_tag_2.xml": {
		wantInvalid: false,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			entries := feed.GetItems()
			assert.Len(t, entries, 1)
			assert.Equal(t, "tag:diveintomark.org,2003-11:blog-14:post-19", feed.Entries[0].GetID())
		},
	},
	"entry_id_tag_3.xml": {
		wantInvalid: false,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			entries := feed.GetItems()
			assert.Len(t, entries, 1)
			assert.Equal(t, "tag:diveintomark.org,2003-11-30:blog-14:post-19", feed.Entries[0].GetID())
		},
	},
	"entry_id_tag_4.xml": {
		wantInvalid: false,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			entries := feed.GetItems()
			assert.Len(t, entries, 1)
			assert.Equal(t, "tag:diveintomark.org,2003-11-30:19", feed.Entries[0].GetID())
		},
	},
	"entry_id_tag_5.xml": {
		wantInvalid: false,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			entries := feed.GetItems()
			assert.Len(t, entries, 1)
			assert.Equal(t, "tag:diveintomark.org,2003-11-30:foo", feed.Entries[0].GetID())
		},
	},
	"entry_id_tag_6.xml": {
		wantInvalid: false,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			entries := feed.GetItems()
			assert.Len(t, entries, 1)
			assert.Equal(t, "tag:me@example.com,2003-11-30:foo", feed.Entries[0].GetID())
		},
	},
	"entry_id_tag_7.xml": {
		wantInvalid: false,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			entries := feed.GetItems()
			assert.Len(t, entries, 1)
			assert.Equal(
				t,
				"tag:raelity.org,2003:/computers/internet/weblogs/blosxom/plugins/atomfeed",
				feed.Entries[0].GetID(),
			)
		},
	},
	"entry_id_tag_8.xml": {
		wantInvalid: false,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			entries := feed.GetItems()
			assert.Len(t, entries, 1)
			assert.Equal(t, "tag:example.org,2004:/2004_07.html#someitemanchor", feed.Entries[0].GetID())
		},
	},
	"entry_id_tag_authority_contains_comma.xml": {
		tests: func(t *testing.T, _ *atom.Feed) {
			t.Helper()
			t.Skip("requires additional logic")
		},
	},
	"entry_id_tag_authority_contains_digit.xml": {
		tests: func(t *testing.T, _ *atom.Feed) {
			t.Helper()
			t.Skip("requires additional logic")
		},
	},
	"entry_id_tag_authority_contains_hyphen.xml": {
		tests: func(t *testing.T, _ *atom.Feed) {
			t.Helper()
			t.Skip("requires additional logic")
		},
	},
	"entry_id_tag_authority_contains_space.xml": {
		tests: func(t *testing.T, _ *atom.Feed) {
			t.Helper()
			t.Skip("requires additional logic")
		},
	},
	"entry_id_tag_authority_contains_wacky_chars.xml": {
		tests: func(t *testing.T, _ *atom.Feed) {
			t.Helper()
			t.Skip("requires additional logic")
		},
	},
	"entry_id_tag_date_1_digit_month.xml": {
		tests: func(t *testing.T, _ *atom.Feed) {
			t.Helper()
			t.Skip("requires additional logic")
		},
	},
	"entry_id_tag_date_2_digit_year.xml": {
		tests: func(t *testing.T, _ *atom.Feed) {
			t.Helper()
			t.Skip("requires additional logic")
		},
	},
	"entry_id_tag_date_2_digit_year_2.xml": {
		tests: func(t *testing.T, _ *atom.Feed) {
			t.Helper()
			t.Skip("requires additional logic")
		},
	},
	"entry_id_tag_date_2_digit_year_3.xml": {
		tests: func(t *testing.T, _ *atom.Feed) {
			t.Helper()
			t.Skip("requires additional logic")
		},
	},
	"entry_id_tag_date_contains_space.xml": {
		tests: func(t *testing.T, _ *atom.Feed) {
			t.Helper()
			t.Skip("requires additional logic")
		},
	},
	"entry_id_tag_date_contains_space_2.xml": {
		tests: func(t *testing.T, _ *atom.Feed) {
			t.Helper()
			t.Skip("requires additional logic")
		},
	},
	"entry_id_tag_date_cutoff.xml": {
		tests: func(t *testing.T, _ *atom.Feed) {
			t.Helper()
			t.Skip("requires additional logic")
		},
	},
	"entry_id_tag_date_cutoff_2.xml": {
		tests: func(t *testing.T, _ *atom.Feed) {
			t.Helper()
			t.Skip("requires additional logic")
		},
	},
	"entry_id_tag_date_cutoff_3.xml": {
		tests: func(t *testing.T, _ *atom.Feed) {
			t.Helper()
			t.Skip("requires additional logic")
		},
	},
	"entry_id_tag_date_cutoff_4.xml": {
		tests: func(t *testing.T, _ *atom.Feed) {
			t.Helper()
			t.Skip("requires additional logic")
		},
	},
	"entry_id_tag_date_missing_year.xml": {
		tests: func(t *testing.T, _ *atom.Feed) {
			t.Helper()
			t.Skip("requires additional logic")
		},
	},
	"entry_id_tag_date_no_hyphens.xml": {
		tests: func(t *testing.T, _ *atom.Feed) {
			t.Helper()
			t.Skip("requires additional logic")
		},
	},
	"entry_id_tag_date_no_hyphens_2.xml": {
		tests: func(t *testing.T, _ *atom.Feed) {
			t.Helper()
			t.Skip("requires additional logic")
		},
	},
	"entry_id_tag_date_too_specific.xml": {
		tests: func(t *testing.T, _ *atom.Feed) {
			t.Helper()
			t.Skip("requires additional logic")
		},
	},
	"entry_id_tag_no_date.xml": {
		tests: func(t *testing.T, _ *atom.Feed) {
			t.Helper()
			t.Skip("requires additional logic")
		},
	},
	"entry_id_tag_specific_contains_space.xml": {
		tests: func(t *testing.T, _ *atom.Feed) {
			t.Helper()
			t.Skip("requires additional logic")
		},
	},
	"entry_id_tag_specific_contains_space_2.xml": {
		tests: func(t *testing.T, _ *atom.Feed) {
			t.Helper()
			t.Skip("requires additional logic")
		},
	},
	"entry_id_tag_specific_contains_wacky_chars.xml": {
		tests: func(t *testing.T, _ *atom.Feed) {
			t.Helper()
			t.Skip("requires additional logic")
		},
	},
	"entry_id_tag_specific_contains_wacky_chars_2.xml": {
		tests: func(t *testing.T, _ *atom.Feed) {
			t.Helper()
			t.Skip("requires additional logic")
		},
	},
	"entry_id_tag_specific_contains_wacky_chars_3.xml": {
		tests: func(t *testing.T, _ *atom.Feed) {
			t.Helper()
			t.Skip("requires additional logic")
		},
	},
	"entry_id_urn_multiple_colons.xml": {
		tests: func(t *testing.T, _ *atom.Feed) {
			t.Helper()
			t.Skip("requires additional logic")
		},
	},
	"entry_id_urn_nid_contains_period.xml": {
		tests: func(t *testing.T, _ *atom.Feed) {
			t.Helper()
			t.Skip("requires additional logic")
		},
	},
	"entry_id_urn_nid_contains_plus.xml": {
		tests: func(t *testing.T, _ *atom.Feed) {
			t.Helper()
			t.Skip("requires additional logic")
		},
	},
	"entry_id_urn_nid_contains_slash.xml": {
		tests: func(t *testing.T, _ *atom.Feed) {
			t.Helper()
			t.Skip("requires additional logic")
		},
	},
	"entry_id_urn_nid_starts_with_hyphen.xml": {
		tests: func(t *testing.T, _ *atom.Feed) {
			t.Helper()
			t.Skip("requires additional logic")
		},
	},
	"entry_id_urn_nss_contains_letters.xml": {
		tests: func(t *testing.T, _ *atom.Feed) {
			t.Helper()
			t.Skip("requires additional logic")
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
	"entry_issued_hours_minutes.xml": {
		wantDecodeErr: true,
	},
	"entry_issued_multiple.xml": {
		tests: func(t *testing.T, _ *atom.Feed) {
			t.Helper()
			t.Skip("Go's XML parser won't allow multiple.")
		},
	},
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
	"entry_link_multiple.xml": {
		tests: func(t *testing.T, _ *atom.Feed) {
			t.Helper()
			t.Skip("Go's XML parser won't allow multiple.")
		},
	},
	"entry_link_multiple2.xml": {
		tests: func(t *testing.T, _ *atom.Feed) {
			t.Helper()
			t.Skip("Go's XML parser won't allow multiple.")
		},
	},
	"entry_link_multiple3.xml": {
		tests: func(t *testing.T, _ *atom.Feed) {
			t.Helper()
			t.Skip("Go's XML parser won't allow multiple.")
		},
	},
	"entry_link_multiple4.xml": {
		tests: func(t *testing.T, _ *atom.Feed) {
			t.Helper()
			t.Skip("Go's XML parser won't allow multiple.")
		},
	},
	"entry_link_multiple5.xml": {
		tests: func(t *testing.T, _ *atom.Feed) {
			t.Helper()
			t.Skip("Go's XML parser won't allow multiple.")
		},
	},
	"entry_link_multiple6.xml": {
		tests: func(t *testing.T, _ *atom.Feed) {
			t.Helper()
			t.Skip("Go's XML parser won't allow multiple.")
		},
	},
	"entry_link_not_multiple.xml": {
		tests: func(t *testing.T, _ *atom.Feed) {
			t.Helper()
			t.Skip("Go's XML parser won't allow multiple.")
		},
	},
	"entry_link_not_multiple2.xml": {
		tests: func(t *testing.T, _ *atom.Feed) {
			t.Helper()
			t.Skip("Go's XML parser won't allow multiple.")
		},
	},
	"entry_link_not_multiple3.xml": {
		tests: func(t *testing.T, _ *atom.Feed) {
			t.Helper()
			t.Skip("Go's XML parser won't allow multiple.")
		},
	},
	"entry_link_not_empty.xml": {
		tests: func(t *testing.T, _ *atom.Feed) {
			t.Helper()
			t.Skip("spec mismatch")
		},
	},
	"entry_link_rel_alternate.xml": {
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			entries := feed.GetItems()
			assert.Len(t, entries, 1)
			assert.Equal(t, atom.LinkRelAlternate, *feed.Entries[0].Links[0].Rel)
		},
	},
	"entry_link_rel_blank.xml": {
		wantInvalid: true,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			require.Error(t, feed.Entries[0].Links[0].Validate())
		},
	},
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
			assert.Equal(t, atom.LinkRelRelated, *feed.Entries[0].Links[0].Rel)
		},
	},
	"entry_link_rel_via.xml": {
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			entries := feed.GetItems()
			assert.Len(t, entries, 1)
			assert.Equal(t, atom.LinkRelVia, *feed.Entries[0].Links[0].Rel)
		},
	},
	"entry_link_title_blank.xml": {
		wantInvalid: true,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			entries := feed.GetItems()
			assert.Len(t, entries, 1)
			assert.Error(t, feed.Entries[0].Links[0].Validate())
		},
	},
	"entry_link_title.xml": {
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			entries := feed.GetItems()
			assert.Len(t, entries, 1)
			assert.Equal(t, "pretty much anything is OK here", *feed.Entries[0].Links[0].Title)
		},
	},
	"entry_link_type_blank.xml": {
		wantInvalid: true,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			entries := feed.GetItems()
			assert.Len(t, entries, 1)
			assert.Error(t, feed.Entries[0].Links[0].Validate())
		},
	},
	"entry_link_type_not_mime.xml": {
		wantInvalid: true,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			entries := feed.GetItems()
			assert.Len(t, entries, 1)
			assert.Error(t, feed.Entries[0].Links[0].Validate())
		},
	},
	"entry_link_type.xml": {
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			entries := feed.GetItems()
			assert.Len(t, entries, 1)
			assert.Equal(t, "text/html", *feed.Entries[0].Links[0].Type)
		},
	},
	"entry_link_type2.xml": {
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			entries := feed.GetItems()
			assert.Len(t, entries, 1)
			assert.Equal(t, "application/xhtml+xml", *feed.Entries[0].Links[0].Type)
		},
	},
	"entry_link_type3.xml": {
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			entries := feed.GetItems()
			assert.Len(t, entries, 1)
			assert.Equal(t, "image/jpeg", *feed.Entries[0].Links[0].Type)
		},
	},
	"entry_link_type4.xml": {
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			entries := feed.GetItems()
			assert.Len(t, entries, 1)
			assert.Equal(t, "text/plain", *feed.Entries[0].Links[0].Type)
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
	"entry_modified_hours_minutes.xml": {
		wantDecodeErr: true,
	},
	"entry_modified_multiple.xml": {
		tests: func(t *testing.T, _ *atom.Feed) {
			t.Helper()
			t.Skip("Go's XML parser won't allow multiple.")
		},
	},
	"entry_modified_missing.xml": {
		wantInvalid: true,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			entries := feed.GetItems()
			assert.Len(t, entries, 1)
			failedValidations, err := getFailedValidations(validation.ValidateStruct(feed.Entries[0]))
			require.NoError(t, err)
			assert.Contains(t, failedValidations["Entry.Updated"], "required")
		},
	},
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
	"entry_summary_contains_html_cdata.xml": {
		wantInvalid: true,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			entries := feed.GetItems()
			assert.Len(t, entries, 1)
			assert.Error(t, validation.ValidateStruct(feed.Entries[0].Summary.Validate()))
		},
	},
	"entry_summary_contains_html.xml": {
		wantInvalid: true,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			entries := feed.GetItems()
			assert.Len(t, entries, 1)
			assert.Error(t, validation.ValidateStruct(feed.Entries[0].Summary.Validate()))
		},
	},
	"entry_summary_is_html.xml": {
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			entries := feed.GetItems()
			assert.Len(t, entries, 1)
			assert.Equal(t, "<b>Bold summary</b>", feed.Entries[0].Summary.String())
		},
	},
	"entry_summary_is_inline.xml": {
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			entries := feed.GetItems()
			assert.Len(t, entries, 1)
			assert.Equal(t, "<code>&lt;p&gt;foo&lt;/p&gt;</code>", feed.Entries[0].Summary.String())
		},
	},
	"entry_summary_is_inline_2.xml": {
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			entries := feed.GetItems()
			assert.Len(t, entries, 1)
			assert.Equal(t, "<code>&lt;p&gt;foo&lt;/p&gt;</code>", feed.Entries[0].Summary.String())
		},
	},
	"entry_summary_no_html.xml": {
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			entries := feed.GetItems()
			assert.Len(t, entries, 1)
			assert.Equal(t, "Valid summary", feed.Entries[0].Summary.String())
		},
	},
	"entry_summary_not_escaped.xml": {
		wantInvalid: true,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			assert.Error(t, feed.Entries[0].Summary.Validate())
		},
	},
	"entry_summary_not_html_cdata.xml": {
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			entries := feed.GetItems()
			assert.Len(t, entries, 1)
			assert.Equal(t, "Valid summary", feed.Entries[0].Summary.String())
		},
	},
	"entry_summary_not_html.xml": {
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			entries := feed.GetItems()
			assert.Len(t, entries, 1)
			assert.Equal(t, "<a", feed.Entries[0].Summary.String())
		},
	},
	"entry_summary_not_inline_cdata.xml": {
		wantInvalid: true,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			assert.Error(t, feed.Entries[0].Summary.Validate())
		},
	},
	"entry_summary_not_text_plain.xml": {
		wantInvalid: true,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			assert.Error(t, feed.Entries[0].Summary.Validate())
		},
	},
	"entry_summary_not_text_plain2.xml": {
		wantInvalid: true,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			assert.Error(t, feed.Entries[0].Summary.Validate())
		},
	},
	"entry_summary_not_text_plain3.xml": {
		wantInvalid: true,
		tests: func(t *testing.T, _ *atom.Feed) {
			t.Helper()
			t.Skip("need additional logic to check for html inside strings")
		},
	},
	"entry_summary_type_not_mime.xml": {
		wantInvalid: true,
		tests: func(t *testing.T, _ *atom.Feed) {
			t.Helper()
			t.Skip("need additional logic to validate registered mimetypes")
		},
	},
	"entry_summary_type.xml": {
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			entries := feed.GetItems()
			assert.Len(t, entries, 1)
			assert.Equal(t, "text/html", string(*feed.Entries[0].Summary.Type))
		},
	},
	"entry_summary_type2.xml": {
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			entries := feed.GetItems()
			assert.Len(t, entries, 1)
			assert.Equal(t, "application/xhtml+xml", string(*feed.Entries[0].Summary.Type))
		},
	},
	"entry_summary_type3.xml": {
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			entries := feed.GetItems()
			assert.Len(t, entries, 1)
			assert.Equal(t, "image/jpeg", string(*feed.Entries[0].Summary.Type))
		},
	},
	"entry_summary_type4.xml": {
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			entries := feed.GetItems()
			assert.Len(t, entries, 1)
			assert.Equal(t, "text/plain", string(*feed.Entries[0].Summary.Type))
		},
	},
	"entry_summary_type_blank.xml": {
		wantInvalid: true,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			assert.Error(t, feed.Entries[0].Summary.Validate())
		},
	},
	"entry_summary.xml": {
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			entries := feed.GetItems()
			assert.Len(t, entries, 1)
			assert.Equal(t, "Valid summary", feed.Entries[0].Summary.String())
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
	"entry_summary_missing.xml": {},
	"entry_summary_multiple.xml": {
		tests: func(t *testing.T, _ *atom.Feed) {
			t.Helper()
			t.Skip("Go's XML parser ignores multiple values.")
		},
	},
	"entry_title_contains_html_cdata.xml": {
		wantInvalid: true,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			entries := feed.GetItems()
			assert.Len(t, entries, 1)
			assert.Error(t, validation.ValidateStruct(feed.Entries[0].Title.Validate()))
		},
	},
	"entry_title_contains_html.xml": {
		wantInvalid: true,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			entries := feed.GetItems()
			assert.Len(t, entries, 1)
			assert.Error(t, validation.ValidateStruct(feed.Entries[0].Title.Validate()))
		},
	},
	"entry_title_is_html.xml": {
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			entries := feed.GetItems()
			assert.Len(t, entries, 1)
			assert.Equal(t, "<b>Bold title</b>", feed.Entries[0].GetTitle())
		},
	},
	"entry_title_is_inline.xml": {
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			entries := feed.GetItems()
			assert.Len(t, entries, 1)
			assert.Equal(t, "<code>&lt;p&gt;foo&lt;/p&gt;</code>", feed.Entries[0].GetTitle())
		},
	},
	"entry_title_is_inline_2.xml": {
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			entries := feed.GetItems()
			assert.Len(t, entries, 1)
			assert.Equal(t, "<code>&lt;p&gt;foo&lt;/p&gt;</code>", feed.Entries[0].GetTitle())
		},
	},
	"entry_title_missing.xml": {
		wantInvalid: true,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			entries := feed.GetItems()
			assert.Len(t, entries, 1)
			failedValidations, err := getFailedValidations(validation.ValidateStruct(feed.Entries[0]))
			require.NoError(t, err)
			assert.Contains(t, failedValidations["Entry.Title"], "required")
		},
	},
	"entry_title_multiple.xml": {
		tests: func(t *testing.T, _ *atom.Feed) {
			t.Helper()
			t.Skip("Go's XML parser ignores multiple values.")
		},
	},
	"entry_title_no_html.xml": {
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			entries := feed.GetItems()
			assert.Len(t, entries, 1)
			assert.Equal(t, "Valid title", feed.Entries[0].GetTitle())
		},
	},
	"entry_title_not_escaped.xml": {
		wantInvalid: true,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			assert.Len(t, feed.GetItems(), 1)
			assert.Error(t, feed.Entries[0].Title.Validate())
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
		wantInvalid: true,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			assert.Error(t, feed.Entries[0].Title.Validate())
		},
	},
	"entry_title_not_inline_cdata.xml": {
		wantInvalid: true,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			assert.Error(t, feed.Entries[0].Title.Validate())
		},
	},
	"entry_title_not_text_plain.xml": {
		wantInvalid: true,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			assert.Error(t, feed.Entries[0].Title.Validate())
		},
	},
	"entry_title_not_text_plain2.xml": {
		wantInvalid: true,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			assert.Error(t, feed.Entries[0].Title.Validate())
		},
	},
	"entry_title_type_blank.xml": {
		wantInvalid: true,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			assert.Error(t, feed.Entries[0].Title.Validate())
		},
	},
	"entry_title_type_not_mime.xml": {
		wantInvalid: true,
		tests: func(t *testing.T, _ *atom.Feed) {
			t.Helper()
			t.Skip("need additional logic to validate registered mimetypes")
		},
	},
	"entry_title_type.xml": {
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			entries := feed.GetItems()
			assert.Len(t, entries, 1)
			assert.Equal(t, "text/html", string(*feed.Entries[0].Title.Type))
		},
	},
	"entry_title_type2.xml": {
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			entries := feed.GetItems()
			assert.Len(t, entries, 1)
			assert.Equal(t, "application/xhtml+xml", string(*feed.Entries[0].Title.Type))
		},
	},
	"entry_title_type3.xml": {
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			entries := feed.GetItems()
			assert.Len(t, entries, 1)
			assert.Equal(t, "image/jpeg", string(*feed.Entries[0].Title.Type))
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
	"entry_unknown_element.xml": {},
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
			failedValidations, err := getFailedValidations(validation.ValidateStruct(feed.Authors[0]))
			require.NoError(t, err)
			assert.Contains(t, failedValidations["PersonConstruct.Email"], "email")
		},
	},
	"feed_author_email_overloaded.xml": {
		wantInvalid: true,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			assert.Len(t, feed.GetAuthors(), 1)
			failedValidations, err := getFailedValidations(validation.ValidateStruct(feed.Authors[0]))
			require.NoError(t, err)
			assert.Contains(t, failedValidations["PersonConstruct.Email"], "email")
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
	"feed_author_name_contains_html.xml": {
		wantInvalid: true,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			assert.Error(t, feed.Authors[0].Validate())
		},
	},
	"feed_author_name_contains_html_cdata.xml": {
		wantInvalid: true,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			assert.Error(t, feed.Authors[0].Validate())
		},
	},
	"feed_author_name_multiple.xml": {
		tests: func(t *testing.T, _ *atom.Feed) {
			t.Helper()
			t.Skip("Go's XML parser ignores multiple values.")
		},
	},
	"feed_author_unknown_element.xml": {},
	"feed_author_url_contains_comma.xml": {
		wantInvalid: false,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			assert.Len(t, feed.GetAuthors(), 1)
			assert.Equal(t, "http://www.wired.com/news/school/0,1383,54916,00.html", *feed.Authors[0].URI)
		},
	},
	"feed_author_url_ftp.xml": {
		wantInvalid: false,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			assert.Len(t, feed.GetAuthors(), 1)
			assert.Equal(t, "ftp://example.com/", *feed.Authors[0].URI)
		},
	},
	"feed_author_url_http.xml": {
		wantInvalid: false,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			assert.Len(t, feed.GetAuthors(), 1)
			assert.Equal(t, "http://example.com/", *feed.Authors[0].URI)
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
			failedValidations, err := getFailedValidations(validation.ValidateStruct(feed.Contributors[0]))
			require.NoError(t, err)
			assert.Contains(t, failedValidations["PersonConstruct.Email"], "email")
		},
	},
	"feed_contributor_email_overloaded.xml": {
		wantInvalid: true,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			assert.Len(t, feed.GetContributors(), 1)
			failedValidations, err := getFailedValidations(validation.ValidateStruct(feed.Contributors[0]))
			require.NoError(t, err)
			assert.Contains(t, failedValidations["PersonConstruct.Email"], "email")
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
			failedValidations, err := getFailedValidations(validation.ValidateStruct(feed.Contributors[0]))
			require.NoError(t, err)
			assert.Contains(t, failedValidations["PersonConstruct.Name"], "required")
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
	"feed_contributor_name_contains_html.xml": {
		wantInvalid: true,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			assert.Error(t, feed.Contributors[0].Validate())
		},
	},
	"feed_contributor_name_contains_html_cdata.xml": {
		wantInvalid: true,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			assert.Error(t, feed.Contributors[0].Validate())
		},
	},
	"feed_contributor_name_missing.xml": {
		wantInvalid: true,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			assert.Len(t, feed.GetContributors(), 1)
			failedValidations, err := getFailedValidations(validation.ValidateStruct(feed.Contributors[0]))
			require.NoError(t, err)
			assert.Contains(t, failedValidations["PersonConstruct.Name"], "required")
		},
	},
	"feed_contributor_name_multiple.xml": {
		tests: func(t *testing.T, _ *atom.Feed) {
			t.Helper()
			t.Skip("Go's XML parser ignores multiple elements.")
		},
	},
	"feed_contributor_unknown_element.xml": {},
	"feed_contributor_url_contains_comma.xml": {
		wantInvalid: false,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			assert.Len(t, feed.GetContributors(), 1)
			assert.Equal(t, "http://www.wired.com/news/school/0,1383,54916,00.html", *feed.Contributors[0].URI)
		},
	},
	"feed_contributor_url_ftp.xml": {
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			assert.Len(t, feed.GetContributors(), 1)
			assert.Equal(t, "ftp://example.com/", *feed.Contributors[0].URI)
		},
	},
	"feed_contributor_url_http.xml": {
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			assert.Len(t, feed.GetContributors(), 1)
			// require.NoError(t, validation.ValidateStruct(feed.Contributors[0]))
			assert.Equal(t, "http://example.com/", *feed.Contributors[0].URI)
		},
	},
	"feed_contributor_url_multiple.xml": {
		tests: func(t *testing.T, _ *atom.Feed) {
			t.Helper()
			t.Skip("Go's XML parser ignores multiple elements.")
		},
	},
	"feed_copyright_is_inline_2.xml": {
		wantDecodeErr: true,
	},
	"feed_copyright_is_inline.xml": {
		wantDecodeErr: true,
	},
	"feed_generator_contains_comma.xml": {
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			assert.Equal(t, "http://www.wired.com/news/school/0,1383,54916,00.html", *feed.Generator.URI)
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
	"feed_id_multiple.xml": {
		tests: func(t *testing.T, _ *atom.Feed) {
			t.Helper()
			t.Skip("Go's XML parser ignores multiple values.")
		},
	},
	"feed_id_not_full_uri.xml": {
		wantInvalid: true,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			failedValidations, err := getFailedValidations(validation.ValidateStruct(feed.ID))
			require.NoError(t, err)
			assert.Contains(t, failedValidations["ID.Value"], "absolute_uri|urn_rfc2141|uuid")
		},
	},
	"feed_id_not_urn.xml": {
		wantInvalid: true,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			failedValidations, err := getFailedValidations(validation.ValidateStruct(feed.ID))
			require.NoError(t, err)
			assert.Contains(t, failedValidations["ID.Value"], "absolute_uri|urn_rfc2141|uuid")
		},
	},
	"feed_id_not_urn2.xml": {
		wantInvalid: true,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			failedValidations, err := getFailedValidations(validation.ValidateStruct(feed.ID))
			require.NoError(t, err)
			assert.Contains(t, failedValidations["ID.Value"], "absolute_uri|urn_rfc2141|uuid")
		},
	},
	"feed_id_tag.xml": {
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			assert.Equal(t, "tag:diveintomark.org,2003:blog-14:post-19", feed.ID.String())
		},
	},
	"feed_id_tag_2.xml": {
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			assert.Equal(t, "tag:diveintomark.org,2003-11:blog-14:post-19", feed.ID.String())
		},
	},
	"feed_id_tag_3.xml": {
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			assert.Equal(t, "tag:diveintomark.org,2003-11-30:blog-14:post-19", feed.ID.String())
		},
	},
	"feed_id_tag_4.xml": {
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			assert.Equal(t, "tag:diveintomark.org,2003-11-30:19", feed.ID.String())
		},
	},
	"feed_id_tag_5.xml": {
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			assert.Equal(t, "tag:diveintomark.org,2003-11-30:foo", feed.ID.String())
		},
	},
	"feed_id_tag_6.xml": {
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			assert.Equal(t, "tag:me@example.com,2003-11-30:foo", feed.ID.String())
		},
	},
	"feed_id_tag_7.xml": {
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			assert.Equal(t, "tag:raelity.org,2003:/", feed.ID.String())
		},
	},
	"feed_id_tag_authority_contains_comma.xml": {
		tests: func(t *testing.T, _ *atom.Feed) {
			t.Helper()
			t.Skip("requires additional logic")
		},
	},
	"feed_id_tag_authority_contains_digit.xml": {
		tests: func(t *testing.T, _ *atom.Feed) {
			t.Helper()
			t.Skip("requires additional logic")
		},
	},
	"feed_id_tag_authority_contains_hyphen.xml": {
		tests: func(t *testing.T, _ *atom.Feed) {
			t.Helper()
			t.Skip("requires additional logic")
		},
	},
	"feed_id_tag_authority_contains_space.xml": {
		tests: func(t *testing.T, _ *atom.Feed) {
			t.Helper()
			t.Skip("requires additional logic")
		},
	},
	"feed_id_tag_authority_contains_wacky_chars.xml": {
		tests: func(t *testing.T, _ *atom.Feed) {
			t.Helper()
			t.Skip("requires additional logic")
		},
	},
	"feed_id_tag_date_1_digit_month.xml": {
		tests: func(t *testing.T, _ *atom.Feed) {
			t.Helper()
			t.Skip("requires additional logic")
		},
	},
	"feed_id_tag_date_2_digit_year.xml": {
		tests: func(t *testing.T, _ *atom.Feed) {
			t.Helper()
			t.Skip("requires additional logic")
		},
	},
	"feed_id_tag_date_2_digit_year_2.xml": {
		tests: func(t *testing.T, _ *atom.Feed) {
			t.Helper()
			t.Skip("requires additional logic")
		},
	},
	"feed_id_tag_date_2_digit_year_3.xml": {
		tests: func(t *testing.T, _ *atom.Feed) {
			t.Helper()
			t.Skip("requires additional logic")
		},
	},
	"feed_id_tag_date_contains_space.xml": {
		tests: func(t *testing.T, _ *atom.Feed) {
			t.Helper()
			t.Skip("requires additional logic")
		},
	},
	"feed_id_tag_date_contains_space_2.xml": {
		tests: func(t *testing.T, _ *atom.Feed) {
			t.Helper()
			t.Skip("requires additional logic")
		},
	},
	"feed_id_tag_date_contains_space_3.xml": {
		tests: func(t *testing.T, _ *atom.Feed) {
			t.Helper()
			t.Skip("requires additional logic")
		},
	},
	"feed_id_tag_date_cutoff.xml": {
		tests: func(t *testing.T, _ *atom.Feed) {
			t.Helper()
			t.Skip("requires additional logic")
		},
	},
	"feed_id_tag_date_cutoff_2.xml": {
		tests: func(t *testing.T, _ *atom.Feed) {
			t.Helper()
			t.Skip("requires additional logic")
		},
	},
	"feed_id_tag_date_cutoff_3.xml": {
		tests: func(t *testing.T, _ *atom.Feed) {
			t.Helper()
			t.Skip("requires additional logic")
		},
	},
	"feed_id_tag_date_cutoff_4.xml": {
		tests: func(t *testing.T, _ *atom.Feed) {
			t.Helper()
			t.Skip("requires additional logic")
		},
	},
	"feed_id_tag_date_missing_year.xml": {
		tests: func(t *testing.T, _ *atom.Feed) {
			t.Helper()
			t.Skip("requires additional logic")
		},
	},
	"feed_id_tag_date_no_hyphens.xml": {
		tests: func(t *testing.T, _ *atom.Feed) {
			t.Helper()
			t.Skip("requires additional logic")
		},
	},
	"feed_id_tag_date_no_hyphens_2.xml": {
		tests: func(t *testing.T, _ *atom.Feed) {
			t.Helper()
			t.Skip("requires additional logic")
		},
	},
	"feed_id_tag_date_too_specific.xml": {
		tests: func(t *testing.T, _ *atom.Feed) {
			t.Helper()
			t.Skip("requires additional logic")
		},
	},
	"feed_id_tag_no_date.xml": {
		tests: func(t *testing.T, _ *atom.Feed) {
			t.Helper()
			t.Skip("requires additional logic")
		},
	},
	"feed_id_tag_specific_contains_space.xml": {
		tests: func(t *testing.T, _ *atom.Feed) {
			t.Helper()
			t.Skip("requires additional logic")
		},
	},
	"feed_id_tag_specific_contains_space_2.xml": {
		tests: func(t *testing.T, _ *atom.Feed) {
			t.Helper()
			t.Skip("requires additional logic")
		},
	},
	"feed_id_tag_specific_contains_wacky_chars.xml": {
		tests: func(t *testing.T, _ *atom.Feed) {
			t.Helper()
			t.Skip("requires additional logic")
		},
	},
	"feed_id_tag_specific_contains_wacky_chars_2.xml": {
		tests: func(t *testing.T, _ *atom.Feed) {
			t.Helper()
			t.Skip("requires additional logic")
		},
	},
	"feed_id_tag_specific_contains_wacky_chars_3.xml": {
		tests: func(t *testing.T, _ *atom.Feed) {
			t.Helper()
			t.Skip("requires additional logic")
		},
	},
	"feed_id_urn.xml": {
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			assert.Equal(t, "urn:diveintomark-org:1", feed.ID.String())
		},
	},
	"feed_id_urn_multiple_colons.xml": {
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			assert.Equal(t, "urn:diveintomark-org:20030729:1", feed.ID.String())
		},
	},
	"feed_id_urn_nid_contains_period.xml": {
		tests: func(t *testing.T, _ *atom.Feed) {
			t.Helper()
			t.Skip("requires additional logic")
		},
	},
	"feed_id_urn_nid_contains_plus.xml": {
		tests: func(t *testing.T, _ *atom.Feed) {
			t.Helper()
			t.Skip("requires additional logic")
		},
	},
	"feed_id_urn_nid_contains_slash.xml": {
		tests: func(t *testing.T, _ *atom.Feed) {
			t.Helper()
			t.Skip("requires additional logic")
		},
	},
	"feed_id_urn_nid_starts_with_hyphen.xml": {
		tests: func(t *testing.T, _ *atom.Feed) {
			t.Helper()
			t.Skip("requires additional logic")
		},
	},
	"feed_id_urn_nss_contains_letters.xml": {
		tests: func(t *testing.T, _ *atom.Feed) {
			t.Helper()
			t.Skip("requires additional logic")
		},
	},
	"feed_id_urn_upper.xml": {
		tests: func(t *testing.T, _ *atom.Feed) {
			t.Helper()
			t.Skip("requires additional logic")
		},
	},
	"feed_info_is_inline.xml": {
		tests: func(t *testing.T, _ *atom.Feed) {
			t.Helper()
			t.Skip("not in spec?")
		},
	},
	"feed_info_is_inline_2.xml": {
		tests: func(t *testing.T, _ *atom.Feed) {
			t.Helper()
			t.Skip("not in spec?")
		},
	},
	"feed_info_missing.xml": {
		tests: func(t *testing.T, _ *atom.Feed) {
			t.Helper()
			t.Skip("not in spec?")
		},
	},
	"feed_info_no_html.xml": {
		tests: func(t *testing.T, _ *atom.Feed) {
			t.Helper()
			t.Skip("not in spec?")
		},
	},
	"feed_info_no_html_cdata.xml": {
		tests: func(t *testing.T, _ *atom.Feed) {
			t.Helper()
			t.Skip("not in spec?")
		},
	},
	"feed_link_contains_comma.xml": {
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			assert.Equal(t, "http://www.wired.com/news/school/0,1383,54916,00.html", feed.GetLink())
		},
	},
	"feed_link_ftp.xml": {
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			assert.Equal(t, "ftp://example.com/", feed.GetLink())
		},
	},
	"feed_link_href_missing.xml": {
		wantInvalid: true,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			failedValidations, err := getFailedValidations(validation.ValidateStruct(feed))
			require.NoError(t, err)
			assert.Contains(t, failedValidations["Feed.Links[0]"], "validateFn")
		},
	},
	"feed_link_http.xml": {
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			assert.Equal(t, "http://example.com/", feed.GetLink())
		},
	},
	"feed_link_mailto.xml": {
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			assert.Equal(t, "mailto:me@example.com", feed.GetLink())
		},
	},
	"feed_link_multiple.xml": {
		tests: func(t *testing.T, _ *atom.Feed) {
			t.Helper()
			t.Skip("multiple links are ignored by parser.")
		},
	},
	"feed_link_multiple2.xml": {
		tests: func(t *testing.T, _ *atom.Feed) {
			t.Helper()
			t.Skip("multiple links are ignored by parser.")
		},
	},
	"feed_link_not_empty.xml": {
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			assert.Equal(t, "http://example.com/", feed.GetLink())
		},
	},
	"feed_link_not_multiple.xml": {
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			assert.Equal(t, "http://example.com/", feed.Links[0].String())
			assert.Equal(t, atom.LinkRelAlternate, *feed.Links[0].Rel)
			assert.Equal(t, "text/html", *feed.Links[0].Type)
			assert.Equal(t, "http://example.com/index.atom", feed.Links[1].String())
			assert.Equal(t, atom.LinkRel("start"), *feed.Links[1].Rel)
			assert.Equal(t, "application/atom+xml", *feed.Links[1].Type)
		},
	},
	"feed_link_not_multiple2.xml": {
		wantInvalid: true,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			assert.Error(t, feed.Validate())
		},
	},
	"feed_link_not_multiple3.xml": {
		wantInvalid: true,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			assert.Error(t, feed.Validate())
		},
	},
	"feed_link_rel_alternate.xml": {
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			assert.Equal(t, atom.LinkRelAlternate, *feed.Links[0].Rel)
		},
	},
	"feed_link_rel_blank.xml": {
		wantInvalid: true,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			assert.Error(t, feed.Validate())
		},
	},
	"feed_link_rel_invalid.xml": {
		tests: func(t *testing.T, _ *atom.Feed) {
			t.Helper()
			t.Skip("requires additional validation logic.")
		},
	},
	"feed_link_rel_invalid2.xml": {
		tests: func(t *testing.T, _ *atom.Feed) {
			t.Helper()
			t.Skip("requires additional validation logic.")
		},
	},
	"feed_link_title.xml": {
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			assert.Equal(t, "pretty much anything is OK here", *feed.Links[0].Title)
		},
	},
	"feed_link_title_blank.xml": {
		wantInvalid: true,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			assert.Error(t, feed.Validate())
		},
	},
	"feed_link_type.xml": {
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			assert.Equal(t, "text/html", *feed.Links[0].Type)
		},
	},
	"feed_link_type2.xml": {
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			assert.Equal(t, "application/xhtml+xml", *feed.Links[0].Type)
		},
	},
	"feed_link_type3.xml": {
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			assert.Equal(t, "image/jpeg", *feed.Links[0].Type)
		},
	},
	"feed_link_type4.xml": {
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			assert.Equal(t, "text/plain", *feed.Links[0].Type)
		},
	},
	"feed_link_type_blank.xml": {
		wantInvalid: true,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			assert.Error(t, feed.Validate())
		},
	},
	"feed_link_type_not_mime.xml": {
		tests: func(t *testing.T, _ *atom.Feed) {
			t.Helper()
			t.Skip("requires additional logic")
		},
	},
	"feed_missing.xml": {
		tests: func(t *testing.T, _ *atom.Feed) {
			t.Helper()
			t.Skip("parser is lenient and won't fail")
		},
	},
	"feed_missing2.xml": {
		tests: func(t *testing.T, _ *atom.Feed) {
			t.Helper()
			t.Skip("parser is lenient and won't fail")
		},
	},
	"feed_modified.xml": {
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			assert.Equal(t, "2003-07-01 01:55:07 -0500 -0500", feed.GetUpdatedDate().String())
		},
	},
	"feed_modified_bad_day.xml": {
		wantDecodeErr: true,
	},
	"feed_modified_bad_day2.xml": {
		wantDecodeErr: true,
	},
	"feed_modified_bad_hours.xml": {
		wantDecodeErr: true,
	},
	"feed_modified_bad_minutes.xml": {
		wantDecodeErr: true,
	},
	"feed_modified_bad_month.xml": {
		wantDecodeErr: true,
	},
	"feed_modified_bad_seconds.xml": {
		wantDecodeErr: true,
	},
	"feed_modified_extra_spaces.xml": {
		wantDecodeErr: true,
	},
	"feed_modified_extra_spaces2.xml": {
		wantDecodeErr: true,
	},
	"feed_modified_extra_spaces3.xml": {
		wantDecodeErr: true,
	},
	"feed_modified_extra_spaces4.xml": {
		wantDecodeErr: true,
	},
	"feed_modified_extra_spaces5.xml": {
		wantDecodeErr: true,
	},
	"feed_modified_fractional_second.xml": {
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			assert.Equal(t, "2002-12-31 19:20:30.45 +0100 +0100", feed.GetUpdatedDate().String())
		},
	},
	"feed_modified_hours_minutes.xml": {
		wantDecodeErr: true,
	},
	"feed_modified_missing.xml": {},
	"feed_modified_multiple.xml": {
		tests: func(t *testing.T, _ *atom.Feed) {
			t.Helper()
			t.Skip("multiple values are ignored by parser.")
		},
	},
	"feed_modified_no_colons.xml": {
		wantDecodeErr: true,
	},
	"feed_modified_no_hyphens.xml": {
		wantDecodeErr: true,
	},
	"feed_modified_no_t.xml": {
		wantDecodeErr: true,
	},
	"feed_modified_no_timezone_colon.xml": {
		wantDecodeErr: true,
	},
	"feed_modified_no_year.xml": {
		wantDecodeErr: true,
	},
	"feed_modified_seconds.xml": {
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			assert.Equal(t, "2002-12-31 19:20:30 +0100 +0100", feed.GetUpdatedDate().String())
		},
	},
	"feed_modified_utc.xml": {
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			assert.Equal(t, "2002-12-31 19:20:30 +0000 UTC", feed.GetUpdatedDate().String())
		},
	},
	"feed_modified_wrong_format.xml": {
		wantDecodeErr: true,
	},
	"feed_namespace_01.xml": {
		tests: func(t *testing.T, _ *atom.Feed) {
			t.Helper()
			t.Skip("out of scope.")
		},
	},
	"feed_namespace_invalid.xml": {
		tests: func(t *testing.T, _ *atom.Feed) {
			t.Helper()
			t.Skip("namespaces not validated.")
		},
	},
	"feed_namespace_missing.xml": {
		tests: func(t *testing.T, _ *atom.Feed) {
			t.Helper()
			t.Skip("namespaces not validated.")
		},
	},
	"feed_namespace_missing_dc.xml": {
		tests: func(t *testing.T, _ *atom.Feed) {
			t.Helper()
			t.Skip("namespaces not validated.")
		},
	},
	"feed_tagline.xml": {
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			assert.Equal(t, "Valid tagline", feed.Subtitle.String())
		},
	},
	"feed_tagline_cdata.xml": {
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			assert.Equal(t, "Valid tagline", feed.Subtitle.String())
		},
	},
	"feed_tagline_is_inline.xml": {
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			assert.Equal(t, "<code>&lt;p&gt;foo&lt;/p&gt;</code>", feed.Subtitle.String())
		},
	},
	"feed_tagline_is_inline_2.xml": {
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			assert.Equal(t, "<code>&lt;p&gt;foo&lt;/p&gt;</code>", feed.Subtitle.String())
		},
	},
	"feed_title_contains_html.xml": {
		wantInvalid: true,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			assert.Error(t, feed.Title.Validate())
		},
	},
	"feed_title_contains_html_cdata.xml": {
		wantInvalid: true,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			assert.Error(t, feed.Title.Validate())
		},
	},
	"feed_title_is_html.xml": {
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			assert.Equal(t, "<b>Bold title</b>", feed.GetTitle())
		},
	},
	"feed_title_is_inline.xml": {
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			assert.Equal(t, "<code>&lt;p&gt;foo&lt;/p&gt;</code>", feed.GetTitle())
		},
	},
	"feed_title_is_inline_2.xml": {
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			assert.Equal(t, "<code>&lt;p&gt;foo&lt;/p&gt;</code>", feed.GetTitle())
		},
	},
	"feed_title_missing.xml": {
		wantInvalid: true,
		tests: func(t *testing.T, feed *atom.Feed) {
			failedValidations, err := getFailedValidations(validation.ValidateStruct(feed))
			require.NoError(t, err)
			assert.Contains(t, failedValidations["Feed.Title"], "required")
		},
	},
	"feed_title_multiple.xml": {
		tests: func(t *testing.T, _ *atom.Feed) {
			t.Helper()
			t.Skip("multiple values are ignored by parser.")
		},
	},
	"feed_title_no_html.xml": {
		tests: func(t *testing.T, feed *atom.Feed) {
			assert.Equal(t, "Valid title", feed.GetTitle())
		},
	},
	"feed_title_no_html_cdata.xml": {
		tests: func(t *testing.T, feed *atom.Feed) {
			assert.Equal(t, "Valid title", feed.GetTitle())
		},
	},
	"feed_title_not_escaped.xml": {
		wantInvalid: true,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			assert.Error(t, feed.Title.Validate())
		},
	},
	"feed_title_not_html.xml": {
		wantInvalid: true,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			assert.Error(t, feed.Title.Validate())
		},
	},
	"feed_title_not_inline.xml": {
		wantInvalid: true,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			assert.Error(t, feed.Title.Validate())
		},
	},
	"feed_title_not_inline_cdata.xml": {
		wantInvalid: true,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			assert.Error(t, feed.Title.Validate())
		},
	},
	"feed_title_not_text_plain.xml": {
		wantInvalid: true,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			assert.Error(t, feed.Title.Validate())
		},
	},
	"feed_title_not_text_plain2.xml": {
		wantInvalid: true,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			assert.Error(t, feed.Title.Validate())
		},
	},
	"feed_title_type.xml": {
		wantInvalid: true,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			assert.Equal(t, atom.TextConstructType("text/html"), *feed.Title.Type)
		},
	},
	"feed_title_type2.xml": {
		wantInvalid: true,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			assert.Equal(t, atom.TextConstructType("application/xhtml+xml"), *feed.Title.Type)
		},
	},
	"feed_title_type3.xml": {
		wantInvalid: true,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			assert.Equal(t, atom.TextConstructType("image/jpeg"), *feed.Title.Type)
		},
	},
	"feed_title_type_blank.xml": {
		wantInvalid: true,
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			assert.Error(t, feed.Title.Validate())
		},
	},
	"feed_title_type_not_mime.xml": {
		tests: func(t *testing.T, _ *atom.Feed) {
			t.Helper()
			t.Skip("requires additional logic")
		},
	},
	"feed_unknown_element.xml": {
		tests: func(t *testing.T, _ *atom.Feed) {
			t.Helper()
			t.Skip("unknown elements are ignored by parser.")
		},
	},
	"feed_unknown_element_core_namespace.xml": {
		tests: func(t *testing.T, _ *atom.Feed) {
			t.Helper()
			t.Skip("unknown elements are ignored by parser.")
		},
	},
	"feed_unknown_element_pubdate.xml": {
		tests: func(t *testing.T, _ *atom.Feed) {
			t.Helper()
			t.Skip("unknown elements are ignored by parser.")
		},
	},
	"feed_version_01.xml": {
		tests: func(t *testing.T, _ *atom.Feed) {
			t.Helper()
			t.Skip("out of scope.")
		},
	},
	"feed_version_02.xml": {
		tests: func(t *testing.T, _ *atom.Feed) {
			t.Helper()
			t.Skip("out of scope.")
		},
	},
	"feed_version_021.xml": {
		tests: func(t *testing.T, _ *atom.Feed) {
			t.Helper()
			t.Skip("out of scope.")
		},
	},
	"feed_xml_id_attribute.xml": {},
	"invalid_xhtml_namespace.xml": {
		tests: func(t *testing.T, _ *atom.Feed) {
			t.Helper()
			t.Skip("requires additional logic")
		},
	},
	"invalid_xml.xml": {
		wantDecodeErr: true,
	},
	"unknown_element_in_known_namespace.xml": {
		tests: func(t *testing.T, _ *atom.Feed) {
			t.Helper()
			t.Skip("unknown elements are ignored")
		},
	},
	"unknown_namespace.xml": {
		tests: func(t *testing.T, _ *atom.Feed) {
			t.Helper()
			t.Skip("unknown namespaces are ignored")
		},
	},
	"valid_dc_all.xml": {
		tests: func(t *testing.T, feed *atom.Feed) {
			t.Helper()
			assert.True(t, slices.ContainsFunc(feed.Namespaces, func(e extensions.Namespace) bool {
				return e.Prefix == "dc"
			}))
			assert.Equal(t, "dive into mark", feed.Title.String())
			assert.Len(t, feed.Links, 1)
			assert.Equal(t, atom.LinkRelAlternate, *feed.Links[0].Rel)
			assert.Equal(t, "text/html", *feed.Links[0].Type)
			assert.Equal(t, "http://diveintomark.org/", feed.Links[0].String())
			assert.Equal(t, "urn:uuid:60a76c80-d399-11d9-b93C-0003939e0af6", feed.ID.String())
			assert.Equal(t, "2003-12-13 18:30:02 +0000 UTC", feed.GetUpdatedDate().String())
			assert.Len(t, feed.Authors, 1)
			assert.Equal(t, "Mark Pilgrim", feed.Authors[0].Name)
			assert.Len(t, feed.DcTitle, 1)
			assert.Equal(t, "Foo en DC", feed.DcTitle[0])
			assert.Len(t, feed.DcCreator, 1)
			assert.Equal(t, "Bob Jones", feed.DcCreator[0])
			assert.Len(t, feed.DcSubject, 1)
			assert.Equal(t, "Examples", feed.DcSubject[0])
			assert.Len(t, feed.DcDescription, 1)
			assert.Equal(t, "Bar with extra DCness", feed.DcDescription[0])
			assert.Len(t, feed.DcPublisher, 1)
			assert.Equal(t, "Example Inc.", feed.DcPublisher[0])
			assert.Len(t, feed.DcContributor, 1)
			assert.Equal(t, "Bill Smith", feed.DcContributor[0])
			assert.Len(t, feed.DcDate, 1)
			assert.Equal(t, "2005-07-04", feed.DcDate[0].Value.Format(time.DateOnly))
			assert.Len(t, feed.DcType, 1)
			assert.Equal(t, dc.Text, feed.DcType[0])
			assert.Len(t, feed.DcFormat, 1)
			assert.Equal(t, "text/html", feed.DcFormat[0])
			assert.Len(t, feed.DcIdentifier, 1)
			assert.Equal(t, "http://example.org/", feed.DcIdentifier[0])
			assert.Len(t, feed.DcLanguage, 1)
			assert.Equal(t, "en-US", feed.DcLanguage[0])
			assert.Len(t, feed.DcRelation, 1)
			assert.Equal(t, "http://example.net/", feed.DcRelation[0])
			assert.Len(t, feed.DcCoverage, 1)
			assert.Equal(t, "Earth", feed.DcCoverage[0])
			assert.Len(t, feed.DcRights, 1)
			assert.Equal(t, "Copyright 2005 Example Inc.", feed.DcRights[0])
			assert.Len(t, feed.Entries, 1)
			entry := feed.Entries[0]
			assert.Equal(t, "Atom 0.3 snapshot", entry.Title.String())
			assert.Len(t, entry.Links, 1)
			assert.Equal(t, atom.LinkRelAlternate, *entry.Links[0].Rel)
			assert.Equal(t, "text/html", *entry.Links[0].Type)
			assert.Equal(t, "http://diveintomark.org/2003/12/13/atom03", entry.Links[0].String())
			assert.Equal(t, "tag:diveintomark.org,2003:3.2397", entry.GetID())
			assert.Equal(t, "2003-12-13 08:29:29 -0400 -0400", entry.GetPublishedDate().String())
			assert.Equal(t, "2003-12-13 18:30:02 +0000 UTC", entry.GetUpdatedDate().String())
			assert.Len(t, entry.DcTitle, 1)
			assert.Equal(t, "Bar en DC", entry.DcTitle[0])
			assert.Len(t, entry.DcCreator, 1)
			assert.Equal(t, "Bob Jones", entry.DcCreator[0])
			assert.Len(t, entry.DcSubject, 1)
			assert.Equal(t, "Example Items", entry.DcSubject[0])
			assert.Len(t, entry.DcDescription, 1)
			assert.Equal(t, "Foo with extra DCness", entry.DcDescription[0])
			assert.Len(t, entry.DcPublisher, 1)
			assert.Equal(t, "Example Inc.", entry.DcPublisher[0])
			assert.Len(t, entry.DcContributor, 1)
			assert.Equal(t, "Bill Smith", entry.DcContributor[0])
			assert.Len(t, entry.DcDate, 1)
			assert.Equal(t, "2005-07-04", entry.DcDate[0].Value.Format(time.DateOnly))
			assert.Len(t, entry.DcType, 1)
			assert.Equal(t, dc.Text, entry.DcType[0])
			assert.Len(t, entry.DcFormat, 1)
			assert.Equal(t, "text/html", entry.DcFormat[0])
			assert.Len(t, entry.DcIdentifier, 1)
			assert.Equal(t, "http://example.org/1", entry.DcIdentifier[0])
			assert.Len(t, entry.DcSource, 1)
			assert.Equal(t, "http://example.com/1", entry.DcSource[0])
			assert.Len(t, entry.DcLanguage, 1)
			assert.Equal(t, "en-US", entry.DcLanguage[0])
			assert.Len(t, entry.DcRelation, 1)
			assert.Equal(t, "http://example.net/", entry.DcRelation[0])
			assert.Len(t, entry.DcCoverage, 1)
			assert.Equal(t, "Earth", entry.DcCoverage[0])
			assert.Len(t, entry.DcRights, 1)
			assert.Equal(t, "Copyright 2005 Example Inc.", entry.DcRights[0])
		},
	},
}

// atomTests are the groups of test cases to run.
var atomTests = map[string]map[string]atomTestSuite{
	"test/feedvalidator/testcases/atom/1.1":  atomOtherTests,
	"test/feedvalidator/testcases/atom/must": atomMustTests,
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
				t.Errorf("could not read file: %s: %s", name, err.Error())
				t.Fail()
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
			feed, err := Decode[*atom.Feed]("", bytes.NewReader(tt.args.data))
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
