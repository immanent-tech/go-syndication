// Copyright 2026 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: 	AGPL-3.0-or-later

package linter

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/url"
	"regexp"

	feeds "github.com/immanent-tech/go-syndication"
	"github.com/immanent-tech/go-syndication/atom"
	"github.com/immanent-tech/go-syndication/jsonfeed"
	"github.com/immanent-tech/go-syndication/rdf"
	"github.com/immanent-tech/go-syndication/rss"
	"github.com/immanent-tech/go-syndication/types"
)

const (
	Passed  ResultStatus = "passed"
	Failed  ResultStatus = "failed"
	Ignored ResultStatus = "ignored"
)

type ResultStatus string

// Lintable represents a syndication source that can be linted.
type Lintable interface {
	rss.Channel | atom.Feed | jsonfeed.Feed | rdf.Channel
}

// Check is a function that evaluates whether a given linter rule is valid. The function takes a source, performs
// whatever validation/checks are required for the rule, and returns a Result indicating whether the rule passed.
type Check[T Lintable] func(source T) Result

// Result contains the result of a rule. It consists of a status indicating whether the rule passed, failed or was
// ignored and an optional message that explains the result. For Status "failed" or "ignored", Message is required. For
// Status "passed", Message is optional.
type Result struct {
	metadata

	Status  ResultStatus `validate:"required,oneof=passed failed ignored"                 json:"status"`
	Message *string      `validate:"required_if=Status failed|required_if=Status ignored" json:"message,omitempty"`
}

type metadata struct {
	ID          string `validate:"required"`
	Description string `validate:"required"`
}

// Rule is a linter rule. It is represented by Check function that will return a Result indicating whether the rule's
// conditions are valid. The Result will also include and ID and description to identify the Rule.
type Rule[T Lintable] struct {
	Check Check[T] `validate:"required"`
}

func pass(m metadata) Result {
	return Result{
		ID:          m.ID,
		Description: m.Description,
		Status:      Passed,
	}
}

func fail(m metadata, format string, args ...any) Result {
	return Result{
		ID:          m.ID,
		Description: m.Description,
		Status:      Failed,
		Message:     msg(format, args...),
	}
}

func ignore(m metadata, format string, args ...any) Result {
	return Result{
		ID:          m.ID,
		Description: m.Description,
		Status:      Ignored,
		Message:     msg(format, args...),
	}
}

func msg(format string, args ...any) *string {
	s := fmt.Sprintf(format, args...)
	return &s
}

// RuleSet is a group of rules, demarcated by some ID, that are similar in purpose. It allows defining a group of rules
// to be applied together to a source.
type RuleSet[T Lintable] map[string][]Rule[T]

func Lint(r io.Reader) (map[string][]Result, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("read data: %w", err)
	}

	// Parse the response as a feed type.
	switch feedType, err := feeds.DetectSourceType(bytes.NewReader(data)); {
	case err != nil:
		return nil, fmt.Errorf("detect feed type: %w", err)
	case feedType == types.SourceUnknown:
		return nil, errors.New("cannot determine feed type")
	case feedType == types.SourceAtom:
		// Atom feed.
		feed, err := feeds.Decode[*atom.Feed]("", bytes.NewReader(data))
		if err != nil {
			return nil, fmt.Errorf("parse atom: %w", err)
		}
		return LintAllAtom(feed), nil
	case feedType == types.SourceRSS:
		// RSS 2.0 feed.
		feed, err := feeds.Decode[*rss.RSS]("", bytes.NewReader(data))
		if err != nil {
			return nil, fmt.Errorf("parse rss: %w", err)
		}
		return LintAllRSS(feed), nil
	case feedType == types.SourceRDF:
		// RDF/RSS 1.0 feed.
		// feedData, err = feeds.NewDecoder[*rdf.RDF](bytes.NewReader(data))
		// if err != nil {
		// 	return nil, fmt.Errorf("parse rdf: %w", err)
		// }
	case feedType == types.SourceJSONFeed:
		// JSONFeed.
		// feedData, err = feeds.NewDecoder[*jsonfeed.Feed](bytes.NewReader(data))
		// if err != nil {
		// 	return nil, fmt.Errorf("parse jsonfeed: %w", err)
		// }
	case feedType == types.SourceHTML:
		return nil, errors.New("go html, not feed format")
	}
	return nil, errors.New("unsupported media type")
}

// validURL reports whether s begins with a URI scheme (as required by section 3.4) and is not a relative reference.
func validURL(s string) bool {
	if s == "" {
		return false
	}
	u, err := url.Parse(s)
	if err != nil {
		return false
	}
	return u.IsAbs()
}

// rssEmailRe loosely matches the recommended "username@hostname.tld (Real Name)" form from section 3.3. It's intentionally
// permissive since the profile doesn't mandate this exact format, only recommends it.
var rssEmailRe = regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+( \(.+\))?$`)

func checkRSSRecommendedEmail(m metadata, value, label string) Result {
	if value == "" {
		return pass(m)
	}
	if !rssEmailRe.MatchString(value) {
		return fail(m, "%s %q should use the form username@hostname.tld (Real Name)", label, value)
	}
	return pass(m)
}
