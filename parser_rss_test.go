// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: 	AGPL-3.0-or-later

package feeds

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/joshuar/go-syndication/atom"
	"github.com/joshuar/go-syndication/mrss"
	"github.com/joshuar/go-syndication/rss"
	"github.com/stretchr/testify/assert"
)

type testSuite struct {
	wantErr bool
	tests   func(t *testing.T, feed *Feed)
}

var rssMustPass = map[string]testSuite{
	// "admin_errorReportsTo.xml": {wantErr: false},
	// "admin_generatorAgent.xml": false,
	"atom_link2.xml": {
		wantErr: false,
		tests: func(t *testing.T, feed *Feed) {
			t.Helper()
			r := toRSS(t, feed)
			assert.Equal(t, atom.LinkRelSelf, *r.Channel.AtomLink.Rel)
			assert.Equal(t, "http://www.rss-world.info/", r.Channel.AtomLink.Value)
			assert.Equal(t, "http://feeds.feedburner.com/rssworld/news", r.Channel.AtomLink.Href)
		},
	},
	"atom_link.xml": {
		wantErr: false,
		tests: func(t *testing.T, feed *Feed) {
			t.Helper()
			r := toRSS(t, feed)
			assert.Equal(t, atom.LinkRelSelf, *r.Channel.AtomLink.Rel)
			assert.Equal(t, "http://www.rss-world.info/", r.Channel.AtomLink.Value)
			assert.Equal(t, "http://feeds.feedburner.com/rssworld/news", r.Channel.AtomLink.Href)
		},
	},
	// TODO: implement blogChannel
	// "blogChannel_blink.xml":             false,
	// "blogChannel_blogRoll.xml":          false,
	// "blogChannel_changes.xml":           false,
	// "blogChannel_mySubscriptions.xml":   false,
	// "cp_server.xml":                     false,
	"dcdate_complete_date.xml": {
		wantErr: false,
		tests: func(t *testing.T, feed *Feed) {
			t.Helper()
			r := toRSS(t, feed)
			assert.Equal(t, "2002-12-31", r.Channel.DCDate.Value.Format(time.DateOnly))
		},
	},
	"dcdate_fractional_second.xml": {
		wantErr: false,
		tests: func(t *testing.T, feed *Feed) {
			t.Helper()
			r := toRSS(t, feed)
			assert.Equal(t, "2002-12-31T19:20:30.45+01:00", r.Channel.DCDate.Value.Format("2006-01-02T15:04:05.00Z07:00"))
		},
	},
	"dcdate_hours_minutes.xml": {
		wantErr: false,
		tests: func(t *testing.T, feed *Feed) {
			t.Helper()
			r := toRSS(t, feed)
			assert.Equal(t, "2002-12-31T19:20+01:00", r.Channel.DCDate.Value.Format(time.DateOnly+"T"+"15:04-07:00"))
		},
	},
	"dc_date_must_include_timezone.xml": {
		wantErr: true,
	},
	"dcdate_seconds.xml": {
		wantErr: false,
		tests: func(t *testing.T, feed *Feed) {
			t.Helper()
			r := toRSS(t, feed)
			assert.Equal(t, "2002-12-31T19:20:30+01:00", r.Channel.DCDate.Value.Format(time.RFC3339))
		},
	},
	"dc_date_with_just_day.xml": {
		wantErr: false,
		tests: func(t *testing.T, feed *Feed) {
			t.Helper()
			r := toRSS(t, feed)
			assert.Equal(t, "2003-09-24", r.Channel.DCDate.Value.Format(time.DateOnly))
		},
	},
	"dcdate.xml": {
		wantErr: false,
		tests: func(t *testing.T, feed *Feed) {
			t.Helper()
			r := toRSS(t, feed)
			assert.Equal(t, "2002-12-31T01:15:07-05:00", r.Channel.DCDate.Value.Format(time.RFC3339))
		},
	},
	"dcdate_year_and_month.xml": {
		wantErr: false,
		tests: func(t *testing.T, feed *Feed) {
			t.Helper()
			r := toRSS(t, feed)
			assert.Equal(t, "2002-12", r.Channel.DCDate.Value.Format("2006-01"))
		},
	},
	"dcdate_year_only.xml": {
		wantErr: false,
		tests: func(t *testing.T, feed *Feed) {
			t.Helper()
			r := toRSS(t, feed)
			assert.Equal(t, "2002", r.Channel.DCDate.Value.Format("2006"))
		},
	},
	"dclanguage_country_code.xml": {
		wantErr: false,
		tests: func(t *testing.T, feed *Feed) {
			t.Helper()
			r := toRSS(t, feed)
			assert.Equal(t, "en-us", r.Channel.DCLanguage.String())
		},
	},
	"dclanguage.xml": {
		wantErr: false,
		tests: func(t *testing.T, feed *Feed) {
			t.Helper()
			r := toRSS(t, feed)
			assert.Equal(t, "en", r.Channel.DCLanguage.String())
		},
	},
	// "doctype_not_entity.xml": {
	// 	wantErr: false,
	// },
	// "doctype_wrong_version.xml": {
	// 	wantErr: true,
	// 	// TODO: doctype parsing...
	// },
	// doctype.xml
	// ev_enddate.xml
	// ev_startdate.xml
	// foaf_name.xml
	// foaf_person.xml
	"ignorable_whitespace.xml": {
		wantErr: false,
		tests: func(t *testing.T, feed *Feed) {
			t.Helper()
			r := toRSS(t, feed)
			assert.Equal(t, "http://example.com/mt/mt-comments.cgi?entryid=1", r.Channel.Items[0].GetComments())
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
	"invalid_rss_version.xml": {wantErr: true},
	// TODO: implement slash hit parade.
	// "invalid_slash_hit_parade.xml":            true,
	"invalid_sy_updateBase_blank.xml":         {wantErr: true},
	"invalid_sy_updateBase.xml":               {wantErr: true},
	"invalid_sy_updateFrequency_blank.xml":    {wantErr: true},
	"invalid_sy_updateFrequency_decimal.xml":  {wantErr: true},
	"invalid_sy_updateFrequency_negative.xml": {wantErr: true},
	"invalid_sy_updateFrequency_zero.xml":     {wantErr: true},
	"invalid_sy_updatePeriod_blank.xml":       {wantErr: true},
	"invalid_sy_updatePeriod.xml":             {wantErr: true},
	"invalid_xml.xml":                         {wantErr: true},
	// "l_permalink.xml": {
	// 	wantErr: false,
	// 	tests: func(t *testing.T, feed *Feed) {
	// 		t.Helper()
	// 		r := toRSS(t, feed)
	// 		assert.Equal(t, "http://www.example.com/", r.Channel.Items[0].PermaLink.Resource)
	// 	},
	// },
	// missing_namespace2.xml
	// missing_namespace_attr_only.xml
	// missing_namespace.xml
	"missing_rss2.xml": {wantErr: true},
	"missing_rss.xml":  {wantErr: true},
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
		wantErr: false,
		tests: func(t *testing.T, feed *Feed) {
			t.Helper()
			r := toRSS(t, feed)
			assert.Equal(t, "Scripting News", r.Channel.GetTitle())
			assert.Equal(t, "http://www.scripting.com/", r.Channel.GetLink())
			assert.Equal(t, "A weblog about scripting and stuff like that.", r.Channel.GetDescription())
			assert.Equal(t, "en-us", r.Channel.GetLanguage())
			assert.Equal(t, "Copyright 1997-2002 Dave Winer", *r.Channel.Copyright)
			assert.Equal(t, "Mon, 30 Sep 2002 11:00:00 GMT", r.Channel.LastBuildDate.Format(time.RFC1123))
			assert.Equal(t, "http://backend.userland.com/rss", *r.Channel.Docs)
			assert.Equal(t, "Radio UserLand v8.0.5", *r.Channel.Generator)
			assert.True(t, slices.Contains(r.Channel.GetCategories(), "1765"))
			assert.Equal(t, "dave@userland.com", *r.Channel.ManagingEditor)
			assert.Equal(t, "dave@userland.com", *r.Channel.WebMaster)
			assert.Equal(t, 40, *r.Channel.TTL)
			assert.Len(t, r.Channel.GetItems(), 9)
			// Check item contents.
			item := r.Channel.Items[8]
			// assert.Equal(t,
			// 	sanitization.SanitizeString("&quot;rssflowersalignright&quot;With any luck we should have one or two more days of namespaces stuff here on Scripting News. It feels like it's winding down. Later in the week I'm going to a &lt;a href=&quot;http://harvardbusinessonline.hbsp.harvard.edu/b02/en/conferences/conf_detail.jhtml?id=s775stg&amp;pid=144XCF&quot;&gt;conference&lt;/a&gt; put on by the Harvard Business School. So that should change the topic a bit. The following week I'm off to Colorado for the &lt;a href=&quot;http://www.digitalidworld.com/conference/2002/index.php&quot;&gt;Digital ID World&lt;/a&gt; conference. We had to go through namespaces, and it turns out that weblogs are a great way to work around mail lists that are clogged with &lt;a href=&quot;http://www.userland.com/whatIsStopEnergy&quot;&gt;stop energy&lt;/a&gt;. I think we solved the problem, have reached a consensus, and will be ready to move forward shortly."),
			// 	r.Channel.Items[0].GetDescription(),
			// )
			assert.Equal(t, "Sun, 29 Sep 2002 11:13:10 GMT", item.GetPublishedDate().Format(time.RFC1123))
			assert.Equal(t, "http://scriptingnews.userland.com/backissues/2002/09/29#reallyEarlyMorningNocoffeeNotes", item.GUID.Value)
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
	// sy_updateBase.xml
	// sy_updateFrequency.xml
	// sy_updatePeriod_daily.xml
	// sy_updatePeriod_hourly.xml
	// sy_updatePeriod_monthly.xml
	// sy_updatePeriod_weekly.xml
	// sy_updatePeriod_yearly.xml
	// thr_children.xml
	// ulcc_channel_url.xml
	// ulcc_item_url.xml
	"unexpected_text.xml": {wantErr: true},
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

var rss20 = map[string]testSuite{
	"element-channel-image-description/image_no_description.xml": {
		wantErr: false,
		tests: func(t *testing.T, feed *Feed) {
			t.Helper()
			r := toRSS(t, feed)
			assert.NotNil(t, r.Channel.Image)
			assert.Equal(t, "Valid image", r.Channel.Image.Title)
			assert.Equal(t, "http://purl.org/rss/2.0/", r.Channel.Image.Link)
			assert.Equal(t, "http://example.com/image.jpg", r.Channel.Image.URL)
		},
	},
	"element-channel/multiple_category.xml": {
		wantErr: false,
		tests: func(t *testing.T, feed *Feed) {
			t.Helper()
			r := toRSS(t, feed)
			assert.Len(t, r.Channel.Categories, 2)
			assert.Equal(t, "Weblogging", r.Channel.Categories[0].String())
			assert.Equal(t, "Navel-gazing", r.Channel.Categories[1].String())
		},
	},
	"element-channel/missing_channel_description.xml":               {wantErr: true},
	"element-channel/missing_channel_link.xml":                      {wantErr: true},
	"element-channel/missing_channel_title.xml":                     {wantErr: true},
	"element-channel-item/invalid_item_no_title_or_description.xml": {wantErr: true},
	"element-channel-link/invalid_link.xml":                         {wantErr: true},
	"element-channel-link/invalid_link2.xml":                        {wantErr: true},
	"element-channel-link/link.xml": {
		wantErr: false,
		tests: func(t *testing.T, feed *Feed) {
			t.Helper()
			r := toRSS(t, feed)
			assert.Equal(t, "http://purl.org/rss/2.0/", r.Channel.GetLink())
		},
	},
	"element-channel-link/link_contains_comma.xml": {
		wantErr: false,
		tests: func(t *testing.T, feed *Feed) {
			t.Helper()
			r := toRSS(t, feed)
			assert.Equal(t, "http://www.wired.com/news/school/0,1383,54916,00.html", r.Channel.GetLink())
		},
	},
	"element-channel-link/link_ftp.xml": {
		wantErr: false,
		tests: func(t *testing.T, feed *Feed) {
			t.Helper()
			r := toRSS(t, feed)
			assert.Equal(t, "ftp://purl.org/rss/2.0/", r.Channel.GetLink())
		},
	},
	"element-rss/missing_channel.xml":           {wantErr: true},
	"element-rss/missing_version_attribute.xml": {wantErr: true},
}

var rssMedia = map[string]testSuite{
	"example1.xml": {
		wantErr: false,
		tests: func(t *testing.T, feed *Feed) {
			t.Helper()
			r := toRSS(t, feed)
			item := r.Channel.Items[0]
			assert.Equal(t, "Story about something", item.GetTitle())
			assert.Equal(t, "http://www.foo.com/item1.htm", item.GetLink())
			assert.Equal(t, "http://www.foo.com/file.mov", item.Enclosure.URL)
			assert.Equal(t, "video/quicktime", item.Enclosure.Type)
			assert.Equal(t, 320000, item.Enclosure.Length)
		},
	},
	"example2.xml": {
		wantErr: false,
		tests: func(t *testing.T, feed *Feed) {
			t.Helper()
			r := toRSS(t, feed)
			item := r.Channel.Items[0]
			assert.Equal(t, "Movie Title: Is this a good movie?", item.GetTitle())
			assert.Equal(t, "http://www.foo.com/trailer.mov", item.MediaContent.Url)
			assert.Equal(t, 12216320, *item.MediaContent.FileSize)
			assert.Equal(t, "video/quicktime", *item.MediaContent.Type)
			assert.Equal(t, mrss.Sample, *item.MediaContent.Expression)
			assert.Equal(t, "nonadult", item.MediaRating.Value)
		},
	},
	"example3.xml": {
		wantErr: false,
		tests: func(t *testing.T, feed *Feed) {
			t.Helper()
			r := toRSS(t, feed)
			item := r.Channel.Items[0]
			assert.Equal(t, "The latest video from an artist", item.GetTitle())
			if assert.NotNil(t, item.MediaContent) {
				assert.Equal(t, "http://www.foo.com/movie.mov", item.MediaContent.Url)
				assert.Equal(t, 12216320, *item.MediaContent.FileSize)
				assert.Equal(t, "video/quicktime", *item.MediaContent.Type)
				assert.Equal(t, mrss.Full, *item.MediaContent.Expression)
				if assert.NotNil(t, item.MediaContent.MediaPlayer) {
					assert.Equal(t, "http://www.foo.com/player?id=1111", item.MediaContent.MediaPlayer.Url)
					assert.Equal(t, 200, *item.MediaContent.MediaPlayer.Height)
					assert.Equal(t, 400, *item.MediaContent.MediaPlayer.Width)
				} else {
					t.Fail()
				}
				assert.Len(t, item.MediaContent.MediaHashes, 1)
				assert.Equal(t, mrss.Md5, *item.MediaContent.MediaHashes[0].Algo)
				assert.Equal(t, "dfdec888b72151965a34b4b59031290a", item.MediaContent.MediaHashes[0].Value)
				assert.Len(t, item.MediaContent.MediaCredits, 2)
				assert.Equal(t, "producer", *item.MediaContent.MediaCredits[0].Role)
				assert.Equal(t, "producer's name", item.MediaContent.MediaCredits[0].Value)
				assert.Equal(t, "music/artist \n                name/album/song", item.MediaContent.GetCategory())
				assert.Equal(t, "http://blah.com/scheme", *item.MediaContent.MediaCategory.Scheme)
				assert.Len(t, item.MediaContent.MediaTexts, 1)
				assert.Equal(t, "Oh, say, can you see, by the dawn's early light", item.MediaContent.MediaTexts[0].GetText())
				assert.Equal(t, "nonadult", item.MediaContent.MediaRating.Value)
			} else {
				t.Fail()
			}
		},
	},
	"valid_thumbnail_time_hms.xml": {
		wantErr: false,
		tests: func(t *testing.T, feed *Feed) {
			t.Helper()
			r := toRSS(t, feed)
			item := r.Channel.Items[0]
			assert.Equal(t, "Movie Title: Is this a good movie?", item.GetTitle())
			assert.Equal(t, "http://www.foo.com/item1.htm", item.GetLink())
			assert.Equal(t, "http://www.foo.com/trailer.mov", item.MediaContent.Url)
			assert.Equal(t, 12216320, *item.MediaContent.FileSize)
			assert.Equal(t, "video/quicktime", *item.MediaContent.Type)
			assert.Equal(t, mrss.Sample, *item.MediaContent.Expression)
			assert.Len(t, item.MediaThumbnails, 1)
			assert.Equal(t, "http://example.com/thumbnail", item.GetImage().URL())
			assert.Equal(t, "12:34:56", *item.MediaThumbnails[0].Time)
		},
	},
}

var rssTests = map[string]map[string]testSuite{
	"test/assets/rss/must":  rssMustPass,
	"test/assets/ext/media": rssMedia,
	"test/assets/rss20":     rss20,
}

func toRSS(t *testing.T, source *Feed) *rss.RSS {
	t.Helper()
	r, ok := source.FeedSource.(*rss.RSS)
	if !ok {
		t.Fatal("Unable to convert to RSS")
	}
	return r
}

func TestNewFeedFromBytesRSS(t *testing.T) {
	type args struct {
		data []byte
	}
	tests := []struct {
		name  string
		args  args
		want  *Feed
		suite testSuite
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
					suite testSuite
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
			feed, err := NewFeedFromBytes[*rss.RSS](tt.args.data)
			// Check test suite error condition.
			if (err != nil) != tt.suite.wantErr {
				spew.Dump(feed)
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
