// Copyright 2026 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: 	AGPL-3.0-or-later

package linter

import (
	"slices"

	"github.com/immanent-tech/go-syndication/atom"
)

// AtomRuleSets contains all the rulesets for linting RSS feeds.
var AtomRuleSets RuleSet[atom.Feed] = map[string][]Rule[atom.Feed]{
	// Valid ruleset checks the feed is valid as per the Atom specification.
	"Atom-Spec-Validation": {
		{
			Check: func(f atom.Feed) Result {
				metadata := metadata{ID: "valid-feed", Description: "RSS Feed passes validation"}
				if err := f.Validate(); err != nil {
					return fail(metadata, "feed is invalid: %s", err.Error())
				}
				return pass(metadata)
			},
		},
	},
	// Go-Syndication-Recommended are recommendations for feeds from the developers of go-syndication. These are above and
	// beyond the specification or RSS advisory board recommendations and reflect modern expectations and sensibilities
	// for feeds.
	"Go-Syndication-Atom-Recommended": {
		{
			Check: func(f atom.Feed) Result {
				metadata := metadata{
					ID:          "feed-should-have-image",
					Description: "Feed should supply an <atom:logo> or <atom:image> that consumers can use to represent it",
				}
				if f.GetImage() == nil {
					return fail(metadata, "feed has no <atom:logo> or <atom:icon>")
				}
				return pass(metadata)
			},
		},
		{
			Check: func(f atom.Feed) Result {
				metadata := metadata{
					ID:          "items-should-have-images",
					Description: "Items should supply an <media:thumbnail> consumers can use to represent the item",
				}
				for i, item := range f.Entries {
					if item.GetImage() == nil {
						return fail(metadata, "item %d has no <media:thumbnail>", i)
					}
				}
				return pass(metadata)
			},
		},
		{
			Check: func(f atom.Feed) Result {
				metadata := metadata{
					ID:          "feed-should-provide-copyright",
					Description: "Feed should have an <atom:rights> element so consumers understand sharing and distribution rights of the published content",
				}
				if f.Rights == nil {
					return fail(metadata, "feed has no <atom:rights>")
				}
				return pass(metadata)
			},
		},
	},
}

func LintAllAtom(feed *atom.Feed) map[string][]Result {
	results := make(map[string][]Result)
	for ruleset, rules := range AtomRuleSets {
		ruleResults := make([]Result, 0, len(rules))
		for rule := range slices.Values(rules) {
			ruleResults = append(ruleResults, rule.Check(*feed))
		}
		results[ruleset] = ruleResults
	}
	return results
}
