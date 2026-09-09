// Copyright 2026 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: 	AGPL-3.0-or-later

package linter

import (
	"slices"
	"strings"
	"time"

	"github.com/immanent-tech/go-syndication/atom"
	"github.com/immanent-tech/go-syndication/rss"
)

// RSSRuleSets contains all of the rulesets for linting RSS feeds.
var RSSRuleSets RuleSet[rss.Channel] = map[string][]Rule[rss.Channel]{
	// Valid ruleset checks the feed is valid as per the RSS specification.
	"RSS-Spec-Validation": {
		{
			Check: func(source rss.Channel) Result {
				metadata := metadata{ID: "valid-feed", Description: "RSS Feed passes validation"}
				if err := source.Validate(); err != nil {
					return fail(metadata, "feed is invalid: %s", err.Error())
				}
				return pass(metadata)
			},
		},
	},
	// RSS-Best-Practices ruleset checks the feed meets the recommendations in the "RSS Best Practices Profile".
	//
	// https://www.rssboard.org/rss-profile
	"RSS-Best-Practices": {
		{
			Check: func(c rss.Channel) Result {
				metadata := metadata{
					ID:          "channel-should-have-atom-link-self",
					Description: "A feed should contain an atom:link with rel=\"self\" identifying its own URL (5.1.1).",
				}
				if c.AtomLink == nil {
					return fail(metadata, "channel has no atom:link rel=\"self\"")
				}
				if c.AtomLink.Rel == nil {
					return fail(metadata, "channel has no atom:link rel=\"self\"")
				}
				if *c.AtomLink.Rel != atom.LinkRelSelf {
					return fail(metadata, "channel has no atom:link rel=\"self\"")
				}
				if !validURL(c.AtomLink.String()) {
					return fail(
						metadata,
						"channel atom:link rel=\"self\" %q is not an absolute URL",
						c.AtomLink.String(),
					)
				}
				return pass(metadata)
			},
		},
		{
			Check: func(c rss.Channel) Result {
				metadata := metadata{
					ID:          "channel-image-title-matches-channel-title",
					Description: "The image's title should match the channel's title (4.1.1.9.2).",
				}
				if c.Image == nil || c.Image.Title == "" {
					return pass(metadata)
				}
				if c.Image.Title != c.Title {
					return fail(
						metadata,
						"channel image title %q does not match channel title %q",
						c.Image.Title,
						c.Title,
					)
				}
				return pass(metadata)
			},
		},
		{
			Check: func(c rss.Channel) Result {
				metadata := metadata{
					ID:          "channel-image-link-matches-channel-link",
					Description: "The image's link should match the channel's link (4.1.1.9.1).",
				}
				if c.Image == nil || c.Image.Link == "" {
					return pass(metadata)
				}
				if c.Image.Link != c.Link {
					return fail(metadata, "channel image link %q does not match channel link %q", c.Image.Link, c.Link)
				}
				return pass(metadata)
			},
		},
		{
			Check: func(c rss.Channel) Result {
				metadata := metadata{
					ID:          "channel-textinput-discouraged",
					Description: "The profile discourages textInput; few aggregators support it (4.1.1.17).",
				}
				if c.TextInput != nil {
					return fail(
						metadata,
						"channel uses textInput, which is discouraged and supported by very few aggregators",
					)
				}
				return pass(metadata)
			},
		},
		{
			Check: func(c rss.Channel) Result {
				metadata := metadata{
					ID:          "channel-managingeditor-email-format",
					Description: "managingEditor should use the form username@hostname.tld (Real Name) (3.3).",
				}
				if c.ManagingEditor != nil {
					return checkRSSRecommendedEmail(metadata, *c.ManagingEditor, "channel managingEditor")
				}
				return ignore(metadata, "no managing editor field")
			},
		},
		{
			Check: func(c rss.Channel) Result {
				metadata := metadata{
					ID:          "channel-webmaster-email-format",
					Description: "webMaster should use the form username@hostname.tld (Real Name) (3.3).",
				}
				if c.WebMaster != nil {
					return checkRSSRecommendedEmail(metadata, *c.WebMaster, "channel webMaster")
				}
				return ignore(metadata, "no webmaster field")
			},
		},
		{
			Check: func(c rss.Channel) Result {
				metadata := metadata{
					ID:          "channel-category-hierarchical",
					Description: "A category's value should be a slash-delimited string identifying a hierarchical position in the taxonomy (4.1.1.4).",
				}
				for _, cat := range c.Categories {
					if strings.TrimSpace(cat.Value) == "" {
						return fail(metadata, "channel has an empty category value")
					}
				}
				return pass(metadata)
			},
		},
		{
			Check: func(c rss.Channel) Result {
				metadata := metadata{
					ID:          "item-guid-should-be-present",
					Description: "A publisher should provide a guid with each item (4.1.1.20.6).",
				}
				for i, item := range c.Items {
					if item.GUID == nil || strings.TrimSpace(item.GUID.Value) == "" {
						return fail(metadata, "item %d has no guid", i)
					}
				}
				return pass(metadata)
			},
		},
		{
			Check: func(c rss.Channel) Result {
				metadata := metadata{
					ID:          "item-pubdate-not-in-future",
					Description: "Items should not be published in the feed until they are ready; a future pubDate is not reliably honored by aggregators (4.1.1.20.8).",
				}
				now := time.Now()
				for i, item := range c.Items {
					if item.PubDate.Value.After(now) {
						return fail(metadata, "item %d has a pubDate (%s) in the future", i, item.PubDate)
					}
				}
				return pass(metadata)
			},
		},
		{
			Check: func(c rss.Channel) Result {
				metadata := metadata{
					ID:          "item-enclosure-zero-length-when-unknown",
					Description: "When an enclosure's size can't be determined (e.g. streaming media), length should be 0 rather than omitted/negative (4.1.1.20.5).",
				}
				for i, item := range c.Items {
					if item.Enclosure != nil {
						if item.Enclosure.Length < 0 {
							return fail(
								metadata,
								"item %d enclosure has a negative length; use 0 when the size is unknown",
								i,
							)
						}
					}
				}
				return pass(metadata)
			},
		},
		{
			Check: func(c rss.Channel) Result {
				metadata := metadata{
					ID:          "item-author-vs-dc-creator",
					Description: "A feed published by an individual should omit item author and rely on channel managingEditor/webMaster instead (4.1.1.20.1).",
				}
				// This is a soft, informational recommendation: we can only flag the case the profile explicitly calls
				// out — author present with no channel-level editorial contact at all, which suggests a single-author
				// feed misusing item author.
				if c.ManagingEditor != nil || c.WebMaster != nil {
					return pass(metadata)
				}
				for i, item := range c.Items {
					if item.Author != nil && strings.TrimSpace(*item.Author) != "" {
						return fail(
							metadata,
							"item %d uses author, but channel has no managingEditor/webMaster; consider dc:creator for individually-published feeds",
							i,
						)
					}
				}
				return pass(metadata)
			},
		},
		{
			Check: func(c rss.Channel) Result {
				metadata := metadata{
					ID:          "channel-category-values-non-empty",
					Description: "Categories should not be empty strings (4.1.1.20.2).",
				}
				for i, item := range c.Items {
					for _, cat := range item.Categories {
						if strings.TrimSpace(cat.Value) == "" {
							return fail(metadata, "item %d has an empty category value", i)
						}
					}
				}
				return pass(metadata)
			},
		},
	},
	// Go-Syndication-Recommended are recommedation for feeds from the developers of go-syndication. These are above and
	// beyond the specification or RSS advisory board recommendations and reflect modern expectations and sensibilities
	// for feeds.
	"Go-Syndication-RSS-Recommended": {
		{
			Check: func(c rss.Channel) Result {
				metadata := metadata{
					ID:          "channel-should-have-image",
					Description: "Channel should supply an image that consumers can use to represent the feed",
				}
				if c.GetImage() == nil {
					return fail(metadata, "channel has no image")
				}
				return pass(metadata)
			},
		},
		{
			Check: func(c rss.Channel) Result {
				metadata := metadata{
					ID:          "items-should-have-images",
					Description: "Items should supply an image consumers can use to represent the item",
				}
				for i, item := range c.Items {
					if item.GetImage() == nil {
						return fail(metadata, "item %d has no image", i)
					}
				}
				return pass(metadata)
			},
		},
		{
			Check: func(c rss.Channel) Result {
				metadata := metadata{
					ID:          "items-should-have-titles",
					Description: "Items should have titles that summarize what the item is about",
				}
				for i, item := range c.Items {
					if item.GetTitle() == "" {
						return fail(metadata, "item %d has no title", i)
					}
				}
				return pass(metadata)
			},
		},
		{
			Check: func(c rss.Channel) Result {
				metadata := metadata{
					ID:          "channel-should-provide-copyright",
					Description: "Channel should supply copyright information so consumers understand sharing and distribution rights of the published content",
				}
				if c.Copyright == nil {
					return fail(metadata, "channel has no copyright")
				}
				return pass(metadata)
			},
		},
		{
			Check: func(c rss.Channel) Result {
				metadata := metadata{
					ID:          "update-period",
					Description: "Feed should supply an update period for consumers to poll for updates. Without one, consumers may excessively poll the feed for updates.",
				}
				if c.SYUdatePeriod == nil {
					return fail(metadata, "channel has no <sy:updatePeriod> set")
				}
				return pass(metadata)
			},
		},
	},
}

func LintAllRSS(feed *rss.RSS) map[string][]Result {
	results := make(map[string][]Result)
	for ruleset, rules := range RSSRuleSets {
		ruleResults := make([]Result, 0, len(rules))
		for rule := range slices.Values(rules) {
			ruleResults = append(ruleResults, rule.Check(feed.Channel))
		}
		results[ruleset] = ruleResults
	}
	return results
}
