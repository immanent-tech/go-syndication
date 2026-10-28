// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: 	MIT

package feeds

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync/atomic"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/immanent-tech/go-syndication/atom"
	"github.com/immanent-tech/go-syndication/extensions/media"
	"github.com/immanent-tech/go-syndication/rss"
	"github.com/immanent-tech/go-syndication/sanitization"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type rssTestSuite struct {
	wantInvalid   bool
	wantDecodeErr bool
	tests         func(t *testing.T, feed *rss.RSS)
}

var rssMustPass = map[string]rssTestSuite{
	// "admin_errorReportsTo.xml": {wantInvalid: false},
	// "admin_generatorAgent.xml": false,
	"atom_link2.xml": {
		wantInvalid: false,
		tests: func(t *testing.T, feed *rss.RSS) {
			assert := assert.New(t)
			idx := slices.IndexFunc(feed.Channel.AtomLinks, func(e rss.AtomLink) bool {
				return e.Rel != nil && *e.Rel == atom.LinkRelSelf
			})
			assert.GreaterOrEqual(idx, 0, "contains valid atom link")
			assert.Equal("http://www.rss-world.info/", *feed.Channel.AtomLinks[idx].UndefinedContent)
			assert.Equal("http://feeds.feedburner.com/rssworld/news", feed.Channel.AtomLinks[idx].Href)
		},
	},
	"atom_link.xml": {
		wantInvalid: false,
		tests: func(t *testing.T, feed *rss.RSS) {
			t.Helper()
			idx := slices.IndexFunc(feed.Channel.AtomLinks, func(e rss.AtomLink) bool {
				return e.Rel != nil && *e.Rel == atom.LinkRelSelf
			})
			assert.GreaterOrEqual(t, idx, 0)
			assert.Equal(t, "http://www.rss-world.info/", *feed.Channel.AtomLinks[idx].UndefinedContent)
			assert.Equal(t, "http://feeds.feedburner.com/rssworld/news", feed.Channel.AtomLinks[idx].Href)
		},
	},
	// TODO: implement blogChannel
	// "blogChannel_blink.xml":             false,
	// "blogChannel_blogRoll.xml":          false,
	// "blogChannel_changes.xml":           false,
	// "blogChannel_mySubscriptions.xml":   false,
	// "cp_server.xml":                     false,
	"dcdate_complete_date.xml": {
		wantInvalid: false,
		tests: func(t *testing.T, feed *rss.RSS) {
			t.Helper()
			assert.NotNil(t, feed.Channel.DcDate)
			date := feed.Channel.DcDate[0]
			assert.Equal(t, "2002-12-31", date.Value.Format(time.DateOnly))
		},
	},
	"dcdate_fractional_second.xml": {
		wantInvalid: false,
		tests: func(t *testing.T, feed *rss.RSS) {
			t.Helper()
			assert.NotNil(t, feed.Channel.DcDate)
			date := feed.Channel.DcDate[0]
			assert.Equal(
				t,
				"2002-12-31T19:20:30.45+01:00",
				date.Value.Format("2006-01-02T15:04:05.00Z07:00"),
			)
		},
	},
	"dcdate_hours_minutes.xml": {
		wantInvalid: false,
		tests: func(t *testing.T, feed *rss.RSS) {
			t.Helper()
			assert.NotNil(t, feed.Channel.DcDate)
			date := feed.Channel.DcDate[0]
			assert.Equal(t, "2002-12-31T19:20+01:00", date.Value.Format(time.DateOnly+"T"+"15:04-07:00"))
		},
	},
	// "dc_date_must_include_timezone.xml": {
	// 	wantInvalid: true,
	// },
	"dcdate_seconds.xml": {
		wantInvalid: false,
		tests: func(t *testing.T, feed *rss.RSS) {
			t.Helper()
			assert.NotNil(t, feed.Channel.DcDate)
			date := feed.Channel.DcDate[0]
			assert.Equal(t, "2002-12-31T19:20:30+01:00", date.Value.Format(time.RFC3339))
		},
	},
	"dc_date_with_just_day.xml": {
		wantInvalid: false,
		tests: func(t *testing.T, feed *rss.RSS) {
			t.Helper()
			assert.NotNil(t, feed.Channel.DcDate)
			date := feed.Channel.DcDate[0]
			assert.Equal(t, "2003-09-24", date.Value.Format(time.DateOnly))
		},
	},
	"dcdate.xml": {
		wantInvalid: false,
		tests: func(t *testing.T, feed *rss.RSS) {
			t.Helper()
			assert.NotNil(t, feed.Channel.DcDate)
			date := feed.Channel.DcDate[0]
			assert.Equal(t, "2002-12-31T01:15:07-05:00", date.Value.Format(time.RFC3339))
		},
	},
	"dcdate_year_and_month.xml": {
		wantInvalid: false,
		tests: func(t *testing.T, feed *rss.RSS) {
			t.Helper()
			assert.NotNil(t, feed.Channel.DcDate)
			date := feed.Channel.DcDate[0]
			assert.Equal(t, "2002-12", date.Value.Format("2006-01"))
		},
	},
	"dcdate_year_only.xml": {
		wantInvalid: false,
		tests: func(t *testing.T, feed *rss.RSS) {
			t.Helper()
			assert.NotNil(t, feed.Channel.DcDate)
			date := feed.Channel.DcDate[0]
			assert.Equal(t, "2002", date.Value.Format("2006"))
		},
	},
	"dclanguage_country_code.xml": {
		wantInvalid: false,
		tests: func(t *testing.T, feed *rss.RSS) {
			t.Helper()
			assert.Len(t, feed.Channel.DcLanguage, 1)
			assert.Equal(t, "en-us", feed.Channel.DcLanguage[0])
		},
	},
	"dclanguage.xml": {
		wantInvalid: false,
		tests: func(t *testing.T, feed *rss.RSS) {
			t.Helper()
			assert.Len(t, feed.Channel.DcLanguage, 1)
			assert.Equal(t, "en", feed.Channel.DcLanguage[0])
		},
	},
	// "doctype_not_entity.xml": {
	// 	wantInvalid: false,
	// },
	// "doctype_wrong_version.xml": {
	// 	wantInvalid: true,
	// 	// TODO: doctype parsing...
	// },
	// doctype.xml
	// ev_enddate.xml
	// ev_startdate.xml
	// foaf_name.xml
	// foaf_person.xml
	"ignorable_whitespace.xml": {
		wantInvalid: false,
		tests: func(t *testing.T, feed *rss.RSS) {
			t.Helper()

			assert.Equal(
				t,
				"http://example.com/mt/mt-comments.cgi?entryid=1",
				strings.TrimSpace(*feed.Channel.Items[0].Comments),
			)
		},
	},
	// TODO: implement blogChannel
	// "invalid_blogChannel_blink.xml":           true,
	// "invalid_blogChannel_blogRoll.xml":        true,
	// "invalid_blogChannel_mySubscriptions.xml": true,
	// "invalid_dcdate.xml":           true,
	// "invalid_dclanguage_blank.xml": true,
	// "invalid_dclanguage.xml":       true,
	// TODO: implement geo
	// "invalid_geo_geo_latitude.xml":            true,
	// "invalid_geo_geo_longitude.xml":           true,
	// "invalid_geo_geourl_latitude.xml":         true,
	// "invalid_geo_geourl_longitude.xml":        true,
	// "invalid_geo_icbm_latitude.xml":           true,
	// "invalid_geo_icbm_longitude.xml":          true,
	// "invalid_item_rdf_about.xml": true,
	// "invalid_namespace2.xml":     true,
	// "invalid_namespace.xml":      true,
	// "invalid_rdf_about.xml":      true,
	"invalid_rss_version.xml": {wantInvalid: true},
	// TODO: implement slash hit parade.
	// "invalid_slash_hit_parade.xml":            true,
	"invalid_sy_updateBase_blank.xml":         {wantDecodeErr: true},
	"invalid_sy_updateBase.xml":               {wantDecodeErr: true},
	"invalid_sy_updateFrequency_blank.xml":    {wantInvalid: true},
	"invalid_sy_updateFrequency_decimal.xml":  {wantDecodeErr: true},
	"invalid_sy_updateFrequency_negative.xml": {wantInvalid: true},
	"invalid_sy_updateFrequency_zero.xml":     {wantInvalid: true},
	// "invalid_sy_updatePeriod_blank.xml":       {wantInvalid: true},
	// "invalid_sy_updatePeriod.xml": {wantInvalid: true},
	// "invalid_xml.xml": {wantDecodeErr: true},
	// "l_permalink.xml":
	"missing_namespace2.xml":          {wantInvalid: true},
	"missing_namespace_attr_only.xml": {wantInvalid: true},
	// "missing_namespace.xml":           {wantInvalid: true},
	"missing_rss2.xml": {wantInvalid: true},
	"missing_rss.xml":  {wantInvalid: true},
	// multiple_admin_errorReportsTo.xml
	// multiple_admin_generatorAgent.xml
	// multiple_channel1.xml
	// multiple_dccreator.xml
	// multiple_dcdate.xml
	// multiple_dclanguage.xml
	// multiple_dcpublisher.xml
	// multiple_dcrights.xml
	// multiple_item_content_encoded.xml
	// multiple_item_dccreator.xml
	// multiple_item_dcdate.xml
	// multiple_item_dcsubject.xml
	// multiple_items.xml
	// multiple_sy_updateBase.xml
	// multiple_sy_updateFrequency.xml
	// multiple_sy_updatePeriod.xml
	// no_blink.xml
	// nodupl_undefined.xml
	// rdf_about.xml
	// rdf_Description.xml
	// rdfs_seeAlso2.xml
	// rdfs_seeAlso.xml
	// rdf_unknown.xml
	// rss10_contentItems.xml*
	// rss10_image2.xml
	// rss10_image.xml
	// rss10_invalid_namespace2.xml*
	// rss10_invalid_namespace.xml
	// rss10_item_in_channel.xml*
	// rss10_missing_item_link.xml*
	// rss10_missing_items.xml*
	// rss10_missing_item_title.xml*
	// rss10_missing_rdf_about_image.xml
	// rss10_mixedContent.xml
	// rss10_parseType.xml
	// rss10_rdfDescription.xml*
	// rss10_resources.xml*
	// rss10_spec_sample_noerror.xml*
	// rss10_spec_sample_nowarn.xml*
	// rss10_textinput2.xml
	// rss10_textinput.xml
	// rss10_title.xml
	// rss10_trackback_invalid_about.xml
	// rss10_trackback_invalid_ping.xml
	// rss10_trackback.xml*
	// rss10_unexpected_channel_language.xml
	// rss10_unexpected_image_width.xml
	// rss10_unexpected_item_pubDate.xml
	// harper:ignore
	"rss20_spec_sample_noerror.xml": {
		wantInvalid: false,
		tests: func(t *testing.T, feed *rss.RSS) {
			t.Helper()

			assert.Equal(t, "Scripting News", feed.Channel.GetTitle())
			assert.Equal(t, "http://www.scripting.com/", feed.Channel.GetLink())
			assert.Equal(t, "A weblog about scripting and stuff like that.", feed.Channel.GetDescription())
			assert.Equal(t, "en-us", *feed.Channel.GetLanguage())
			assert.Equal(t, "Copyright 1997-2002 Dave Winer", *feed.Channel.Copyright)
			assert.Equal(t, "Mon, 30 Sep 2002 11:00:00 +0000", feed.Channel.LastBuildDate.String())
			assert.Equal(t, "http://backend.userland.com/rss", *feed.Channel.Docs)
			assert.Equal(t, "Radio UserLand v8.0.5", *feed.Channel.Generator)
			assert.True(t, slices.Contains(feed.Channel.GetCategories(), "1765"))
			assert.Equal(t, "dave@userland.com", *feed.Channel.ManagingEditor)
			assert.Equal(t, "dave@userland.com", *feed.Channel.WebMaster)
			assert.Equal(t, 40, feed.Channel.TTL)
			assert.Len(t, feed.Channel.GetItems(), 9)
			// Check item contents.
			item := feed.Channel.Items[8]
			// assert.Equal(t,
			// 	sanitization.SanitizeString("&quot;rssflowersalignright&quot;With any luck we should have one or two more days of namespaces stuff here on Scripting News. It feels like it's winding down. Later in the week I'm going to a &lt;a href=&quot;http://harvardbusinessonline.hbsp.harvard.edu/b02/en/conferences/conf_detail.jhtml?id=s775stg&amp;pid=144XCF&quot;&gt;conference&lt;/a&gt; put on by the Harvard Business School. So that should change the topic a bit. The following week I'm off to Colorado for the &lt;a href=&quot;http://www.digitalidworld.com/conference/2002/index.php&quot;&gt;Digital ID World&lt;/a&gt; conference. We had to go through namespaces, and it turns out that weblogs are a great way to work around mail lists that are clogged with &lt;a href=&quot;http://www.userland.com/whatIsStopEnergy&quot;&gt;stop energy&lt;/a&gt;. I think we solved the problem, have reached a consensus, and will be ready to move forward shortly."),
			// 	feed.Channel.Items[0].GetDescription(),
			// )
			assert.Equal(t, "Sun, 29 Sep 2002 11:13:10 +0000", item.PubDate.String())
			assert.Equal(
				t,
				"http://scriptingnews.userland.com/backissues/2002/09/29#reallyEarlyMorningNocoffeeNotes",
				item.GUID.Value,
			)
			assert.Equal(
				t,
				"http://scriptingnews.userland.com/backissues/2002/09/29#reallyEarlyMorningNocoffeeNotes",
				item.GetLink(),
			)
			assert.Equal(t, "Really early morning no-coffee notes", item.GetTitle())
		},
	},
	// rss20_trackback_invalid_about.xml
	// rss20_trackback_invalid_ping.xml
	// rss20_trackback.xml*
	// rss91n_deprecated.xml
	// rss91n_entity.xml
	// rss91rab.xml
	// rss91u_entity.xml
	// slash_zero_comments.xml
	// "sy_updateBase.xml": {wantInvalid: false},
	// sy_updateFrequency.xml
	// sy_updatePeriod_daily.xml
	// sy_updatePeriod_hourly.xml
	// sy_updatePeriod_monthly.xml
	// sy_updatePeriod_weekly.xml
	// sy_updatePeriod_yearly.xml
	// thr_children.xml
	// ulcc_channel_url.xml
	// ulcc_item_url.xml
	// "unexpected_text.xml": {wantInvalid: true},
	// unknown_element2.xml
	// unknown_element_in_known_namespace.xml
	// unknown_element.xml
	// unknown_namespace.xml
	// unknown_root_element.xml
	// valid_ag_all.xml*
	// valid_all_rss2_attributes.xml
	// valid_dc_all2.xml
	// valid_dc_all.xml
	// valid_dcterms_all2.xml*
	// valid_dcterms_all.xml*
	// valid_ev_all.xml
	// valid_geo_all.xml*
	// valid_rss_090.xml
	// valid_slash_all.xml
	// valid_taxo_all.xml
	// xml_utf-8_bom_with_ascii_declaration.xml
	// xmlversion_10.xml
	// xmlversion_11.xml
}

var rss20 = map[string]rssTestSuite{
	"element-channel-image-description/image_no_description.xml": {
		wantInvalid: false,
		tests: func(t *testing.T, feed *rss.RSS) {
			t.Helper()

			assert.NotNil(t, feed.Channel.Image)
			assert.Equal(t, "Valid image", feed.Channel.Image.Title)
			assert.Equal(t, "http://purl.org/rss/2.0/", feed.Channel.Image.Link)
			assert.Equal(t, "http://example.com/image.jpg", feed.Channel.Image.URL)
		},
	},
	"element-channel/multiple_category.xml": {
		wantInvalid: false,
		tests: func(t *testing.T, feed *rss.RSS) {
			t.Helper()

			assert.Len(t, feed.Channel.Categories, 2)
			assert.Equal(t, "Weblogging", feed.Channel.Categories[0].String())
			assert.Equal(t, "Navel-gazing", feed.Channel.Categories[1].String())
		},
	},
	"element-channel/missing_channel_description.xml":               {wantInvalid: true},
	"element-channel/missing_channel_link.xml":                      {wantInvalid: true},
	"element-channel/missing_channel_title.xml":                     {wantInvalid: true},
	"element-channel-item/invalid_item_no_title_or_description.xml": {wantInvalid: true},
	"element-channel-link/invalid_link.xml":                         {wantInvalid: true},
	// "element-channel-link/invalid_link2.xml":                        {wantInvalid: true},
	"element-channel-link/link.xml": {
		wantInvalid: false,
		tests: func(t *testing.T, feed *rss.RSS) {
			t.Helper()

			assert.Equal(t, "http://purl.org/rss/2.0/", feed.Channel.GetLink())
		},
	},
	"element-channel-link/link_contains_comma.xml": {
		wantInvalid: false,
		tests: func(t *testing.T, feed *rss.RSS) {
			t.Helper()

			assert.Equal(t, "http://www.wired.com/news/school/0,1383,54916,00.html", feed.Channel.GetLink())
		},
	},
	"element-channel-link/link_ftp.xml": {
		wantInvalid: false,
		tests: func(t *testing.T, feed *rss.RSS) {
			t.Helper()

			assert.Equal(t, "ftp://purl.org/rss/2.0/", feed.Channel.GetLink())
		},
	},
	"element-rss/missing_channel.xml":           {wantInvalid: true},
	"element-rss/missing_version_attribute.xml": {wantInvalid: true},
}

var rssMedia = map[string]rssTestSuite{
	"example1.xml": {
		wantInvalid: false,
		tests: func(t *testing.T, feed *rss.RSS) {
			t.Helper()

			item := feed.Channel.Items[0]
			assert.Equal(t, "Story about something", item.GetTitle())
			assert.Equal(t, "http://www.foo.com/item1.htm", item.GetLink())
			assert.Equal(t, "http://www.foo.com/file.mov", item.Enclosure.URL)
			assert.Equal(t, "video/quicktime", item.Enclosure.Type)
			assert.Equal(t, 320000, item.Enclosure.Length)
		},
	},
	"example2.xml": {
		wantInvalid: false,
		tests: func(t *testing.T, feed *rss.RSS) {
			t.Helper()

			assert.Len(t, feed.Channel.Items, 1)
			item := feed.Channel.Items[0]
			assert.NotNil(t, item.MediaContent)
			assert.Equal(t, "Movie Title: Is this a good movie?", item.GetTitle())
			assert.Equal(t, "http://www.foo.com/trailer.mov", item.MediaContent.URL)
			assert.Equal(t, 12216320, *item.MediaContent.FileSize)
			assert.Equal(t, "video/quicktime", *item.MediaContent.Type)
			assert.Equal(t, media.Sample, *item.MediaContent.Expression)
			assert.Equal(t, "nonadult", item.MediaRating.Value)
		},
	},
	"example3.xml": {
		wantInvalid: false,
		tests: func(t *testing.T, feed *rss.RSS) {
			t.Helper()

			assert.Len(t, feed.Channel.Items, 1)
			item := feed.Channel.Items[0]
			assert.Equal(t, "The latest video from an artist", item.GetTitle())
			if assert.NotNil(t, item.MediaContent) {
				assert.Equal(t, "http://www.foo.com/movie.mov", item.MediaContent.URL)
				assert.Equal(t, 12216320, *item.MediaContent.FileSize)
				assert.Equal(t, "video/quicktime", *item.MediaContent.Type)
				assert.Equal(t, media.Full, *item.MediaContent.Expression)
				if assert.NotNil(t, item.MediaContent.MediaPlayer) {
					assert.Equal(t, "http://www.foo.com/player?id=1111", item.MediaContent.MediaPlayer.URL)
					assert.Equal(t, 200, *item.MediaContent.MediaPlayer.Height)
					assert.Equal(t, 400, *item.MediaContent.MediaPlayer.Width)
				} else {
					t.Fail()
				}
				assert.Len(t, item.MediaContent.MediaHashes, 1)
				assert.Equal(t, media.Md5, *item.MediaContent.MediaHashes[0].Algo)
				assert.Equal(t, "dfdec888b72151965a34b4b59031290a", item.MediaContent.MediaHashes[0].Value)
				assert.Len(t, item.MediaContent.MediaCredits, 2)
				assert.Equal(t, "producer", *item.MediaContent.MediaCredits[0].Role)
				assert.Equal(t, "producer's name", item.MediaContent.MediaCredits[0].Value)
				assert.Equal(t, "music/artist name/album/song", item.MediaContent.GetCategory())
				assert.Equal(t, "http://blah.com/scheme", *item.MediaContent.MediaCategory.Scheme)
				assert.Len(t, item.MediaContent.MediaTexts, 1)
				assert.Equal(
					t,
					"Oh, say, can you see, by the dawn's early light",
					item.MediaContent.MediaTexts[0].GetText(),
				)
				assert.Equal(t, "nonadult", item.MediaContent.MediaRating.Value)
			} else {
				t.Fail()
			}
		},
	},
	"valid_thumbnail_time_hms.xml": {
		wantInvalid: false,
		tests: func(t *testing.T, feed *rss.RSS) {
			t.Helper()

			assert.Len(t, feed.Channel.Items, 1)
			item := feed.Channel.Items[0]
			assert.Equal(t, "Movie Title: Is this a good movie?", item.GetTitle())
			assert.Equal(t, "http://www.foo.com/item1.htm", item.GetLink())
			assert.Equal(t, "http://www.foo.com/trailer.mov", item.MediaContent.URL)
			assert.Equal(t, 12216320, *item.MediaContent.FileSize)
			assert.Equal(t, "video/quicktime", *item.MediaContent.Type)
			assert.Equal(t, media.Sample, *item.MediaContent.Expression)
			assert.Len(t, item.MediaThumbnails, 1)
			assert.NotNil(t, item.GetImage())
			assert.Equal(t, "http://example.com/thumbnail", item.GetImage().GetURL())
			assert.Equal(t, "12:34:56", *item.MediaThumbnails[0].Time)
		},
	},
}

var rssSecFields = map[string]rssTestSuite{
	"bad_values.xml": {
		wantDecodeErr: true,
	},
	"hostile_titles.xml": {
		tests: func(t *testing.T, feed *rss.RSS) {
			t.Helper()
			if strings.Contains(feed.GetTitle(), "&lt;script&gt;window.PWNED=1") {
				t.Errorf("hostile title should appear as escaped text; page was:\n%s", trunc(feed.GetTitle()))
			}
			for item := range slices.Values(feed.GetItems()) {
				for _, bad := range []string{"<img src=x onerror", "<script>window.PWNED", "<svg/onload"} {
					if strings.Contains(item.GetTitle(), bad) {
						t.Errorf("raw hostile markup in title %q reached the page unescaped", bad)
					}
				}
			}
		},
	},
	"links_rss.xml": {
		wantInvalid: true,
		tests: func(t *testing.T, feed *rss.RSS) {
			t.Helper()
			for item := range slices.Values(feed.Channel.Items) {
				var (
					link          string
					enclosureLink string
				)
				link = item.GetLink()
				if item.Enclosure != nil {
					enclosureLink = item.Enclosure.GetURL()
				}
				switch {
				case strings.HasPrefix(item.GetTitle(), "BAD:"):
					if item.Enclosure != nil && enclosureLink != "" {
						t.Errorf(
							"%q: link=%q enclosure=%q",
							item.GetTitle(),
							item.Link,
							item.Enclosure.URL,
						)
					} else if item.Enclosure == nil && link != "" {
						t.Errorf(
							"%q: link=%q",
							item.GetTitle(),
							item.Link,
						)
					}
				}
			}
		},
	},
	"unicode.xml": {
		tests: func(t *testing.T, feed *rss.RSS) {
			t.Helper()
			for item := range slices.Values(feed.GetItems()) {
				if utf8.RuneCountInString(item.GetTitle()) > sanitization.DefaultMaxRunes {
					t.Errorf("title not capped: %d runes", utf8.RuneCountInString(item.GetTitle()))
				}
				for _, r := range item.GetTitle() {
					if (r >= 0x202A && r <= 0x202E) || (r >= 0x2066 && r <= 0x2069) || r == 0x200B || r == 0xFEFF ||
						r < 0x20 {
						t.Errorf("dangerous rune %U survived in %q", r, trunc(item.GetTitle()))
					}
				}
			}

		},
	},
}

var rssSecMalformed = map[string]rssTestSuite{
	"doctype_only.xml": {
		wantDecodeErr: true,
	},
	"empty.xml": {
		wantDecodeErr: true,
	},
	"invalid_utf8.xml": {
		wantDecodeErr: true,
	},
	"not_a_feed.html": {
		wantInvalid: true,
	},
	"nul_bytes.xml": {
		wantDecodeErr: true,
	},
	"truncated.xml": {
		wantDecodeErr: true,
	},
	"unclosed_tags.xml": {
		wantInvalid: true,
	},
	"utf16_declared.xml": {
		wantDecodeErr: true,
	},
	"wrong_encoding.xml": {
		wantDecodeErr: true,
	},
	"wrong_root.xml": {
		wantInvalid: true,
	},
}

var rssSecXSS = map[string]rssTestSuite{
	"01_script_and_handlers.xml": {
		tests: func(t *testing.T, feed *rss.RSS) {
			assert := assert.New(t)
			assert.Equal("<p>hello</p>", feed.Channel.Items[0].GetDescription(), "script tag")
			assert.Nil(feed.Channel.Items[1].GetContent(), "external script")
			assert.Equal(`<img src="x">`, feed.Channel.Items[2].GetDescription(), "img onerror")
			assert.Equal("<svg></svg>", *feed.Channel.Items[3].GetContent(), "svg onload")
			assert.Equal(`<details open="">x</details>`, feed.Channel.Items[4].GetDescription(), "details ontoggle")
			assert.Equal("text", *feed.Channel.Items[5].GetContent(), "body onload")
			assert.NotContains(feed.Channel.Items[6].GetDescription(), " onerror", "slash separators")
			assert.Nil(feed.Channel.Items[7].GetContent(), "unclosed tag")
			assert.NotContains(feed.Channel.Items[8].GetDescription(), "<script>", "split tags")
			assert.Nil(feed.Channel.Items[9].GetContent(), "mixed case and spaces")
			assert.NotContains(feed.Channel.Items[10].GetDescription(), "onclick", "onclick with spaces")
			assert.Nil(feed.Channel.Items[11].GetContent(), "autofocus input")
			assert.NotContains(feed.Channel.Items[12].GetDescription(), "onstart", "marquee onstart")
			assert.Nil(feed.Channel.Items[13].GetContent(), "video source onerror")
			assert.Equal(`<img src="%60x%60">`, feed.Channel.Items[14].GetDescription(), "backtick attribute")
			assert.Equal(
				`<img src="https://example.com/a.png">`,
				*feed.Channel.Items[15].GetContent(),
				"attribute breakout via alt",
			)
			assert.NotContains(feed.Channel.Items[16].GetDescription(), "onmouseover", "onmouseover on anchor")

		},
	},
	"02_obfuscated_urls.xml": {
		tests: func(t *testing.T, feed *rss.RSS) {
			t.Helper()
			assert := assert.New(t)
			assert.Equal("x", feed.Channel.Items[0].GetDescription(), "javascript plain")
			assert.Equal("x", *feed.Channel.Items[1].GetContent(), "javascript mixed case")
			assert.Equal("x", feed.Channel.Items[2].GetDescription(), "decimal entity")
			assert.Equal("x", *feed.Channel.Items[3].GetContent(), "hex entities")
			assert.Equal("x", feed.Channel.Items[4].GetDescription(), "embedded tab")
			assert.Equal("x", *feed.Channel.Items[5].GetContent(), "embedded newline")
			assert.Equal("x", feed.Channel.Items[6].GetDescription(), "leading space")
			assert.Equal("x", *feed.Channel.Items[7].GetContent(), "leading control character")
			assert.Equal("x", feed.Channel.Items[8].GetDescription(), "named colon entity")
			assert.Equal("x", *feed.Channel.Items[9].GetContent(), "data html")
			assert.Equal("x", feed.Channel.Items[10].GetDescription(), "vbscript")
			assert.Nil(feed.Channel.Items[11].GetContent(), "img javascript attribute")
			assert.Equal("x", feed.Channel.Items[12].GetDescription(), "file scheme")
			t.Log("Relative path and protocol tests skipped as sanitization does not handle them")
			// assert.Equal("x", *feed.Channel.Items[13].GetContent(), "protocol relative")
			// assert.Empty(feed.Channel.Items[14].GetDescription(), "relative path")
			assert.Equal(
				`<a href="ftp://example.com/x" rel="nofollow noreferrer">a</a>b`,
				*feed.Channel.Items[15].GetContent(),
				"ftp and gopher",
			)
			// Note the sanitization does not handle relative paths and protocols.
			// assert.Empty(feed.Channel.Items[16].GetDescription(), "userinfo confusion")
		},
	},
	"03_svg_mathml_mxss.xml": {
		tests: func(t *testing.T, feed *rss.RSS) {
			t.Helper()
			assert := assert.New(t)
			assert.Equal("<svg></svg>", feed.Channel.Items[0].GetDescription(), "svg script")
			assert.Equal("<svg>x</svg>", *feed.Channel.Items[1].GetContent(), "svg xlink javascript")
			assert.NotContains(feed.Channel.Items[2].GetDescription(), "onbegin", "svg animate onbegin")
			assert.Equal("<svg></svg>", *feed.Channel.Items[3].GetContent(), "svg use data")
			assert.NotContains(feed.Channel.Items[4].GetDescription(), "PWNED", "math mglyph style")
			assert.Equal("click", *feed.Channel.Items[5].GetContent(), "math href")
			assert.NotContains(feed.Channel.Items[6].GetDescription(), "PWNED", "math maction")
			assert.Nil(feed.Channel.Items[7].GetContent(), "form math mglyph")
			assert.NotContains(feed.Channel.Items[8].GetDescription(), "PWNED", "noscript attr closer")
			assert.NotContains(*feed.Channel.Items[9].GetContent(), "PWNED", "svg style attr closer")
			assert.NotContains(feed.Channel.Items[10].GetDescription(), "PWNED", "xmp attr closer")
			assert.NotContains(*feed.Channel.Items[11].GetContent(), "PWNED", "textarea attr closer")
			assert.NotContains(feed.Channel.Items[12].GetDescription(), "PWNED", "iframe attr closer")
			assert.NotContains(*feed.Channel.Items[13].GetContent(), "PWNED", "title attr closer")
			assert.NotContains(feed.Channel.Items[14].GetDescription(), "PWNED", "svg foreign object")
		},
	},
	"04_dangerous_elements.xml": {
		tests: func(t *testing.T, feed *rss.RSS) {
			t.Helper()
			assert := assert.New(t)
			assert.Empty(feed.Channel.Items[0].GetDescription(), "iframe src")
			assert.Nil(feed.Channel.Items[1].GetContent(), "iframe srcdoc")
			assert.Empty(feed.Channel.Items[2].GetDescription(), "object data")
			assert.Nil(feed.Channel.Items[3].GetContent(), "embed src")
			assert.NotContains("form", feed.Channel.Items[4].GetDescription(), "form action")
			assert.Equal("x", *feed.Channel.Items[5].GetContent(), "button formaction")
			assert.Equal(
				`<a href="/relative" rel="nofollow noreferrer">x</a>`,
				feed.Channel.Items[6].GetDescription(),
				"base href hijack",
			)
			assert.Nil(feed.Channel.Items[7].GetContent(), "meta refresh")
			assert.Empty(feed.Channel.Items[8].GetDescription(), "link stylesheet")
			assert.Nil(feed.Channel.Items[9].GetContent(), "style tag import")
			assert.NotContains(feed.Channel.Items[10].GetDescription(), "PWNED", "style attr javascript url")
			assert.NotContains(*feed.Channel.Items[11].GetContent(), "PWNED", "style attr expression")
			assert.NotContains(feed.Channel.Items[12].GetDescription(), "style", "fixed-position overlay")
			assert.NotContains(*feed.Channel.Items[13].GetContent(), "PWNED", "template element")
			assert.NotContains(feed.Channel.Items[14].GetDescription(), "PWNED", "img srcset")
			assert.Equal(
				`<img src="https://example.com/a.png">`,
				*feed.Channel.Items[15].GetContent(),
				"picture source",
			)
			assert.Empty(feed.Channel.Items[16].GetDescription(), "video poster")
			assert.Equal(
				`<a href="https://example.com/" rel="nofollow noreferrer">x</a>`,
				*feed.Channel.Items[17].GetContent(),
				"anchor ping tracking",
			)
			assert.Equal(
				`<a href="https://example.com/" rel="nofollow noreferrer">x</a>`,
				feed.Channel.Items[18].GetDescription(),
				"target blank no rel",
			)
			assert.Nil(feed.Channel.Items[19].GetContent(), "input type image")
			assert.Empty(feed.Channel.Items[20].GetDescription(), "applet and isindex")
		},
	},
}

var rssTests = map[string]map[string]rssTestSuite{
	"test/feedvalidator/testcases/rss/must":  rssMustPass,
	"test/feedvalidator/testcases/ext/media": rssMedia,
	"test/feedvalidator/testcases/rss20":     rss20,
	"test/feedsec/fields":                    rssSecFields,
	"test/feedsec/malformed":                 rssSecMalformed,
	"test/feedsec/xss":                       rssSecXSS,
}

func TestNewFeedFromBytesRSS(t *testing.T) {
	type args struct {
		data []byte
	}
	tests := []struct {
		name  string
		args  args
		want  *Feed
		suite rssTestSuite
	}{}
	for set, testSuites := range rssTests {
		for name, suite := range testSuites {
			testFile := filepath.Join(set, name)
			data := readFile(t, testFile)
			tests = append(tests, struct {
				name  string
				args  args
				want  *Feed
				suite rssTestSuite
			}{
				name:  "file:" + testFile,
				args:  args{data: data},
				suite: suite,
			})
		}
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			feed, err := Decode[*rss.RSS]("", bytes.NewReader(tt.args.data))
			if err != nil {
				t.Logf("decode error: %v", err)
			}
			if (err != nil) != tt.suite.wantDecodeErr {
				t.Errorf("Decode() error = %v, wantDecodeErr %v", err, tt.suite.wantDecodeErr)
				return
			}

			// Run test suites.
			if tt.suite.tests != nil {
				tt.suite.tests(t, feed)
			}
			// If wantErr, make sure that occurs.
			if tt.suite.wantInvalid {
				err := feed.Validate()
				if err != nil {
					t.Logf("validate error: %v", err)
				}
				if (err != nil) != tt.suite.wantInvalid {
					t.Errorf("Validate() error = %v, wantErr %v", err, tt.suite.wantInvalid)
					return
				}
			}
		})
	}
}

func TestBasicGeo(t *testing.T) {
	exampleFeed := rss.RSS{
		Version: "2.0",
		Channel: rss.Channel{
			Title: "Geotagged items",
			Link:  "http://example.org",
			Items: []rss.Item{
				{
					Title: "A photo taken in Copenhagen",
					Link:  "http://example.org/photo/1",
					Lat:   new(55.701),
					Lon:   new(12.552),
				},
			},
		},
	}
	exampleDoc := `
 <?xml version="1.0"?>
 <?xml-stylesheet href="/eqcenter/catalogs/rssxsl.php?feed=eqs7day-M5.xml" type="text/xsl"
                  media="screen"?>
 <rss version="2.0"
      xmlns:geo="http://www.w3.org/2003/01/geo/wgs84_pos#"
      xmlns:dc="http://purl.org/dc/elements/1.1/">
  <channel>
     <title>USGS M5+ Earthquakes</title>
     <description>Real-time, worldwide earthquake list for the past 7 days</description>
     <link>https://earthquake.usgs.gov/eqcenter/</link>
     <dc:publisher>U.S. Geological Survey</dc:publisher>
     <pubDate>Thu, 27 Dec 2007 23:56:15 PST</pubDate>
     <item>
       <pubDate>Fri, 28 Dec 2007 05:24:17 GMT</pubDate>
       <title>M 5.3, northern Sumatra, Indonesia</title>
       <description>December 28, 2007 05:24:17 GMT</description>
       <link>https://earthquake.usgs.gov/eqcenter/recenteqsww/Quakes/us2007llai.php</link>
       <geo:lat>5.5319</geo:lat>
       <geo:long>95.8972</geo:long>
     </item>
   </channel>
 </rss>
	`

	t.Run("example feed", func(t *testing.T) {
		_, err := Encode(exampleFeed)
		require.NoError(t, err, "no errors in example feed")
	})

	t.Run("example doc", func(t *testing.T) {
		doc, err := Decode[*rss.RSS]("", strings.NewReader(exampleDoc))
		require := require.New(t)
		require.NoError(err, "no errors in example feed")
		require.InEpsilon(5.5319, *doc.Channel.Items[0].Lat, 0.0001, "lat matches")
		require.InEpsilon(95.8972, *doc.Channel.Items[0].Lon, 0.0001, "long matches")
	})
}

func TestXXEAndEntityExpansion(t *testing.T) {
	const secret = "XXE_FILE_CANARY_9f3a7c"
	canaryFile := filepath.Join(t.TempDir(), "secret.txt")
	if err := os.WriteFile(canaryFile, []byte(secret), 0o600); err != nil {
		t.Fatal(err)
	}
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		_, _ = w.Write([]byte("<!ENTITY x 'XXE_HTTP_CANARY'>"))
	}))
	defer srv.Close()

	const testFiles = "test/feedsec/xml_attacks"

	for _, name := range []string{"xxe_file.xml", "xxe_ssrf.xml", "billion_laughs.xml"} {
		t.Run(name, func(t *testing.T) {
			raw := string(readFile(t, filepath.Join(testFiles, name)))
			raw = strings.ReplaceAll(raw, "__CANARY_FILE__", filepath.ToSlash(canaryFile))
			raw = strings.ReplaceAll(raw, "__CANARY_URL__", srv.URL)

			start := time.Now()
			feed, err := Decode[*rss.RSS]("", bytes.NewReader([]byte(raw)))
			if err != nil {
				t.Fatalf("decode error: %v", err)
			}

			if d := time.Since(start); d > 3*time.Second {
				t.Errorf("parsing took %v; possible entity-expansion DoS", d)
			}
			if feed == nil {
				return // rejecting the feed outright is an acceptable outcome
			}

			data, err := Encode(feed)
			if err != nil {
				t.Fatal(err)
			}
			if bytes.Contains(data, []byte(secret)) {
				t.Error("local file contents leaked into output (XXE)")
			}
			if bytes.Contains(data, []byte("XXE_HTTP_CANARY")) {
				t.Error("remote entity content leaked into output")
			}
			if len(data) > 1<<20 {
				t.Errorf("output is %d bytes; entities were expanded", len(data))
			}
		})
	}
	if n := hits.Load(); n != 0 {
		t.Errorf("parser made %d outbound request(s) while parsing hostile feeds", n)
	}
}

func TestResourceExhaustionShapes(t *testing.T) {
	const feedHead = `<?xml version="1.0"?><rss version="2.0"><channel><title>t</title><link>https://feed.example.com/</link><description>d</description>`

	t.Run("deeply nested XML elements", func(t *testing.T) {
		n := 20000
		doc := feedHead + `<item><title>x</title>` + strings.Repeat(
			"<a>",
			n,
		) + strings.Repeat(
			"</a>",
			n,
		) + `</item></channel></rss>`
		_, err := Decode[*rss.RSS]("", bytes.NewReader([]byte(doc)))
		if err != nil {
			t.Fatalf("decode error: %v", err)
		}
	})

	t.Run("multi-megabyte title", func(t *testing.T) {
		doc := feedHead + `<item><title>` + strings.Repeat(
			"A",
			2_000_000,
		) + `</title><link>https://feed.example.com/1</link></item></channel></rss>`
		feed, err := Decode[*rss.RSS]("", bytes.NewReader([]byte(doc)))
		if err != nil {
			t.Fatalf("decode error: %v", err)
		}

		if feed == nil {
			return
		}

		if len(feed.GetTitle()) > 20_000 {
			t.Errorf("rendered feed is %d bytes; titles are not length-capped", len(feed.GetTitle()))
		}

		_, err = Encode(feed)
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("many items", func(t *testing.T) {
		var sb strings.Builder
		sb.WriteString(feedHead)
		for i := range 20000 {
			fmt.Fprintf(&sb, "<item><title>i%d</title><link>https://feed.example.com/%d</link></item>", i, i)
		}
		sb.WriteString("</channel></rss>")
		start := time.Now()
		feed, err := Decode[*rss.RSS]("", bytes.NewReader([]byte(sb.String())))
		if err != nil {
			t.Fatalf("decode error: %v", err)
		}
		if feed != nil {
			t.Logf(
				"parsed %d items in %v; consider capping items per feed in your ingest code",
				len(feed.Channel.Items),
				time.Since(start),
			)
		}
	})
}
