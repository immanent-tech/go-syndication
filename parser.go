// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: 	MIT

package feeds

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/immanent-tech/go-syndication/atom"
	"github.com/immanent-tech/go-syndication/jsonfeed"
	"github.com/immanent-tech/go-syndication/rss"
	"github.com/immanent-tech/go-syndication/types"
)

var (
	// ErrParseBytes indicates an error occurred trying to parse a byte array as a feed.
	ErrParseBytes = errors.New("unable to parse bytes as feed")
)

// NewDecoder will create a new Feed of the given type from the given io.Reader.
func NewDecoder[T any](data io.Reader) (*Feed, error) {
	var (
		original T
		feed     *Feed
		err      error
	)
	if _, ok := any(original).(*jsonfeed.Feed); ok {
		// If the original is JSONFeed, unmarshal as JSON.
		rd := json.NewDecoder(data)
		err = rd.Decode(&original)
	} else {
		// Otherwise, unmarshal as XML.
		original, err = Decode[T]("", data)
	}
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrParseBytes, err)
	}
	source, ok := any(original).(types.FeedSource)
	if !ok {
		return nil, fmt.Errorf("%w: data is not a valid feed type %T", ErrParseBytes, original)
	}
	feed = &Feed{
		FeedSource: source,
	}
	feed.SourceType = parseSource(original)

	return feed, nil
}

// NewFeedFromSource will create a new Feed from the given source that satisfies the FeedSource interface. This can be
// used to create a Feed from an existing rss.RSS or atom.Feed object.
func NewFeedFromSource[T types.FeedSource](source T) *Feed {
	feed := &Feed{
		FeedSource: source,
	}
	feed.SourceType = parseSource(source)
	return feed
}

// parseSource will attempt to determine the appropriate SourceType value from the given interface object.
func parseSource[T any](source T) SourceType {
	switch any(source).(type) {
	case *atom.Feed:
		return TypeAtom
	case *rss.RSS:
		return TypeRSS
	default:
		return ""
	}
}
