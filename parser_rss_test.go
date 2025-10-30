// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: 	AGPL-3.0-or-later

package feeds

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
	"time"

	"github.com/immanent-tech/go-syndication/atom"
	"github.com/immanent-tech/go-syndication/extensions/media"
	"github.com/immanent-tech/go-syndication/rss"
	"github.com/stretchr/testify/assert"
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
			t.Helper()
			assert.Equal(t, atom.LinkRelSelf, feed.Channel.AtomLink.Rel)
			assert.Equal(t, "http://www.rss-world.info/", feed.Channel.AtomLink.Value)
			assert.Equal(t, "http://feeds.feedburner.com/rssworld/news", feed.Channel.AtomLink.Href)
		},
	},
	"atom_link.xml": {
		wantInvalid: false,
		tests: func(t *testing.T, feed *rss.RSS) {
			t.Helper()
			assert.Equal(t, atom.LinkRelSelf, feed.Channel.AtomLink.Rel)
			assert.Equal(t, "http://www.rss-world.info/", feed.Channel.AtomLink.Value)
			assert.Equal(t, "http://feeds.feedburner.com/rssworld/news", feed.Channel.AtomLink.Href)
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
			assert.Equal(t, "2002-12-31", feed.Channel.DCDate.Format(time.DateOnly))
		},
	},
	"dcdate_fractional_second.xml": {
		wantInvalid: false,
		tests: func(t *testing.T, feed *rss.RSS) {
			t.Helper()
			assert.Equal(t, "2002-12-31T19:20:30.45+01:00", feed.Channel.DCDate.Format("2006-01-02T15:04:05.00Z07:00"))
		},
	},
	"dcdate_hours_minutes.xml": {
		wantInvalid: false,
		tests: func(t *testing.T, feed *rss.RSS) {
			t.Helper()
			assert.Equal(t, "2002-12-31T19:20+01:00", feed.Channel.DCDate.Format(time.DateOnly+"T"+"15:04-07:00"))
		},
	},
	"dc_date_must_include_timezone.xml": {
		wantInvalid: true,
	},
	"dcdate_seconds.xml": {
		wantInvalid: false,
		tests: func(t *testing.T, feed *rss.RSS) {
			t.Helper()
			assert.Equal(t, "2002-12-31T19:20:30+01:00", feed.Channel.DCDate.Format(time.RFC3339))
		},
	},
	"dc_date_with_just_day.xml": {
		wantInvalid: false,
		tests: func(t *testing.T, feed *rss.RSS) {
			t.Helper()
			assert.Equal(t, "2003-09-24", feed.Channel.DCDate.Format(time.DateOnly))
		},
	},
	"dcdate.xml": {
		wantInvalid: false,
		tests: func(t *testing.T, feed *rss.RSS) {
			t.Helper()

			assert.Equal(t, "2002-12-31T01:15:07-05:00", feed.Channel.DCDate.Format(time.RFC3339))
		},
	},
	"dcdate_year_and_month.xml": {
		wantInvalid: false,
		tests: func(t *testing.T, feed *rss.RSS) {
			t.Helper()

			assert.Equal(t, "2002-12", feed.Channel.DCDate.Format("2006-01"))
		},
	},
	"dcdate_year_only.xml": {
		wantInvalid: false,
		tests: func(t *testing.T, feed *rss.RSS) {
			t.Helper()

			assert.Equal(t, "2002", feed.Channel.DCDate.Format("2006"))
		},
	},
	"dclanguage_country_code.xml": {
		wantInvalid: false,
		tests: func(t *testing.T, feed *rss.RSS) {
			t.Helper()

			assert.Equal(t, "en-us", feed.Channel.DCLanguage)
		},
	},
	"dclanguage.xml": {
		wantInvalid: false,
		tests: func(t *testing.T, feed *rss.RSS) {
			t.Helper()

			assert.Equal(t, "en", feed.Channel.DCLanguage)
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

			assert.Equal(t, "http://example.com/mt/mt-comments.cgi?entryid=1", feed.Channel.Items[0].Comments.String())
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
	"invalid_sy_updateBase_blank.xml":         {wantInvalid: true},
	"invalid_sy_updateBase.xml":               {wantInvalid: true},
	"invalid_sy_updateFrequency_blank.xml":    {wantInvalid: true},
	"invalid_sy_updateFrequency_decimal.xml":  {wantInvalid: true},
	"invalid_sy_updateFrequency_negative.xml": {wantInvalid: true},
	"invalid_sy_updateFrequency_zero.xml":     {wantInvalid: true},
	// "invalid_sy_updatePeriod_blank.xml":       {wantInvalid: true},
	// "invalid_sy_updatePeriod.xml": {wantInvalid: true},
	"invalid_xml.xml": {wantInvalid: true},
	// "l_permalink.xml":
	"missing_namespace2.xml":          {wantInvalid: true},
	"missing_namespace_attr_only.xml": {wantInvalid: true},
	"missing_namespace.xml":           {wantInvalid: true},
	"missing_rss2.xml":                {wantInvalid: true},
	"missing_rss.xml":                 {wantInvalid: true},
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
			assert.Equal(t, "en-us", feed.Channel.GetLanguage())
			assert.Equal(t, "Copyright 1997-2002 Dave Winer", feed.Channel.Copyright.String())
			assert.Equal(t, "Mon, 30 Sep 2002 11:00:00 GMT", feed.Channel.LastBuildDate.Format(time.RFC1123))
			assert.Equal(t, "http://backend.userland.com/rss", feed.Channel.Docs)
			assert.Equal(t, "Radio UserLand v8.0.5", feed.Channel.Generator.String())
			assert.True(t, slices.Contains(feed.Channel.GetCategories(), "1765"))
			assert.Equal(t, "dave@userland.com", feed.Channel.ManagingEditor.String())
			assert.Equal(t, "dave@userland.com", feed.Channel.WebMaster.String())
			assert.Equal(t, 40, feed.Channel.TTL)
			assert.Len(t, feed.Channel.GetItems(), 9)
			// Check item contents.
			item := feed.Channel.Items[8]
			// assert.Equal(t,
			// 	sanitization.SanitizeString("&quot;rssflowersalignright&quot;With any luck we should have one or two more days of namespaces stuff here on Scripting News. It feels like it's winding down. Later in the week I'm going to a &lt;a href=&quot;http://harvardbusinessonline.hbsp.harvard.edu/b02/en/conferences/conf_detail.jhtml?id=s775stg&amp;pid=144XCF&quot;&gt;conference&lt;/a&gt; put on by the Harvard Business School. So that should change the topic a bit. The following week I'm off to Colorado for the &lt;a href=&quot;http://www.digitalidworld.com/conference/2002/index.php&quot;&gt;Digital ID World&lt;/a&gt; conference. We had to go through namespaces, and it turns out that weblogs are a great way to work around mail lists that are clogged with &lt;a href=&quot;http://www.userland.com/whatIsStopEnergy&quot;&gt;stop energy&lt;/a&gt;. I think we solved the problem, have reached a consensus, and will be ready to move forward shortly."),
			// 	feed.Channel.Items[0].GetDescription(),
			// )
			assert.Equal(t, "Sun, 29 Sep 2002 11:13:10 GMT", item.GetPublishedDate().Format(time.RFC1123))
			assert.Equal(t, "http://scriptingnews.userland.com/backissues/2002/09/29#reallyEarlyMorningNocoffeeNotes", item.GUID.Value.String())
			assert.Equal(t, "http://scriptingnews.userland.com/backissues/2002/09/29#reallyEarlyMorningNocoffeeNotes", item.GetLink())
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
	"element-channel-link/invalid_link2.xml":                        {wantInvalid: true},
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
			assert.Equal(t, "http://www.foo.com/trailer.mov", item.MediaContent.Url)
			assert.Equal(t, 12216320, item.MediaContent.FileSize)
			assert.Equal(t, "video/quicktime", item.MediaContent.Type)
			assert.Equal(t, media.Sample, item.MediaContent.Expression)
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
				assert.Equal(t, "http://www.foo.com/movie.mov", item.MediaContent.Url)
				assert.Equal(t, 12216320, item.MediaContent.FileSize)
				assert.Equal(t, "video/quicktime", item.MediaContent.Type)
				assert.Equal(t, media.Full, item.MediaContent.Expression)
				if assert.NotNil(t, item.MediaContent.MediaPlayer) {
					assert.Equal(t, "http://www.foo.com/player?id=1111", item.MediaContent.MediaPlayer.Url)
					assert.Equal(t, 200, item.MediaContent.MediaPlayer.Height)
					assert.Equal(t, 400, item.MediaContent.MediaPlayer.Width)
				} else {
					t.Fail()
				}
				assert.Len(t, item.MediaContent.MediaHashes, 1)
				assert.Equal(t, media.Md5, item.MediaContent.MediaHashes[0].Algo)
				assert.Equal(t, "dfdec888b72151965a34b4b59031290a", item.MediaContent.MediaHashes[0].Value)
				assert.Len(t, item.MediaContent.MediaCredits, 2)
				assert.Equal(t, "producer", item.MediaContent.MediaCredits[0].Role)
				assert.Equal(t, "producer's name", item.MediaContent.MediaCredits[0].Value)
				assert.Equal(t, "music/artist \n                name/album/song", item.MediaContent.GetCategory())
				assert.Equal(t, "http://blah.com/scheme", item.MediaContent.MediaCategory.Scheme)
				assert.Len(t, item.MediaContent.MediaTexts, 1)
				assert.Equal(t, "Oh, say, can you see, by the dawn's early light", item.MediaContent.MediaTexts[0].GetText())
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

			assert.Len(t, feed.Channel.Items[0], 1)
			item := feed.Channel.Items[0]
			assert.Equal(t, "Movie Title: Is this a good movie?", item.GetTitle())
			assert.Equal(t, "http://www.foo.com/item1.htm", item.GetLink())
			assert.Equal(t, "http://www.foo.com/trailer.mov", item.MediaContent.Url)
			assert.Equal(t, 12216320, item.MediaContent.FileSize)
			assert.Equal(t, "video/quicktime", item.MediaContent.Type)
			assert.Equal(t, media.Sample, item.MediaContent.Expression)
			assert.Len(t, item.MediaThumbnails, 1)
			assert.Equal(t, "http://example.com/thumbnail", item.GetImage().GetURL())
			assert.Equal(t, "12:34:56", item.MediaThumbnails[0].Time)
		},
	},
}

var rssTests = map[string]map[string]rssTestSuite{
	"test/assets/rss/must":  rssMustPass,
	"test/assets/ext/media": rssMedia,
	"test/assets/rss20":     rss20,
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
			data, err := os.ReadFile(testFile) // #nosec G304
			if err != nil {
				t.Error("could not read file: " + name)
			} else {
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			feed, err := Decode[*rss.RSS]("", tt.args.data)
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
