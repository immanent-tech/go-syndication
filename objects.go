// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: 	MIT

package feeds

import (
	"encoding/json"
	"errors"
	"fmt"
	"slices"

	"github.com/immanent-tech/go-syndication/atom"
	"github.com/immanent-tech/go-syndication/jsonfeed"
	"github.com/immanent-tech/go-syndication/rss"
	"github.com/immanent-tech/go-syndication/types"
)

// ErrUnmarshal indicates an error occurred trying to unmarshal data into a given feed object.
var ErrUnmarshal = errors.New("unmarshaling object failed")

// Item represents a single item or entry (or article) in a feed.
type Item struct {
	types.ItemSource `json:"source"`

	SourceType types.SourceType `json:"type"`
	FeedTitle  string           `json:"feed_title"`
}

// UnmarshalJSON handles unmarshaling of an Item from JSON.
func (i *Item) UnmarshalJSON(v []byte) error {
	// Unmarshal the FeedSource based on the type field value.
	sourceType, source, err := sourceFromBytes(v)
	if err != nil {
		return err
	}
	switch sourceType {
	case types.SourceTypeAtom:
		i.SourceType = sourceType
		i.ItemSource, err = unmarshalSource[*atom.Entry](source)
		if err != nil {
			return fmt.Errorf("%w: unable to unmarshal into Atom: %w", ErrUnmarshal, err)
		}
		return nil
	case types.SourceTypeRSS:
		i.SourceType = sourceType
		i.ItemSource, err = unmarshalSource[*rss.Item](source)
		if err != nil {
			return fmt.Errorf("%w: unable to unmarshal into RSS: %w", ErrUnmarshal, err)
		}
		return nil
	case types.SourceTypeJSONFeed:
		i.SourceType = sourceType
		i.ItemSource, err = unmarshalSource[*jsonfeed.Item](source)
		if err != nil {
			return fmt.Errorf("%w: unable to unmarshal into JSONFeed: %w", ErrUnmarshal, err)
		}
		return nil
	}
	return fmt.Errorf("%w: unknown data type", ErrUnmarshal)
}

// Feed represents any feed type containing a number of items.
type Feed struct {
	types.FeedSource `json:"source"`

	SourceType types.SourceType `json:"type"`
}

// GetItems retrieves a slice of Item for the Feed.
func (f *Feed) GetItems() []Item {
	items := make([]Item, 0, len(f.FeedSource.GetItems()))
	for item := range slices.Values(f.FeedSource.GetItems()) {
		items = append(items,
			Item{
				ItemSource: item,
				SourceType: f.SourceType,
				FeedTitle:  f.GetTitle(),
			})
	}
	return items
}

// UnmarshalJSON handles unmarshaling of a Feed from JSON.
func (f *Feed) UnmarshalJSON(v []byte) error {
	// Unmarshal the FeedSource based on the type field value.
	sourceType, source, err := sourceFromBytes(v)
	if err != nil {
		return err
	}
	switch sourceType {
	case types.SourceTypeAtom:
		f.SourceType = sourceType
		f.FeedSource, err = unmarshalSource[*atom.Feed](source)
		if err != nil {
			return fmt.Errorf("%w: unable to unmarshal into Atom: %w", ErrUnmarshal, err)
		}
		return nil
	case types.SourceTypeRSS:
		f.SourceType = sourceType
		f.FeedSource, err = unmarshalSource[*rss.RSS](source)
		if err != nil {
			return fmt.Errorf("%w: unable to unmarshal into RSS: %w", ErrUnmarshal, err)
		}
		return nil
	case types.SourceTypeJSONFeed:
		f.SourceType = sourceType
		f.FeedSource, err = unmarshalSource[*jsonfeed.Feed](source)
		if err != nil {
			return fmt.Errorf("%w: unable to unmarshal into JSONFeed: %w", ErrUnmarshal, err)
		}
		return nil
	}
	return fmt.Errorf("%w: unknown data type", ErrUnmarshal)
}

func sourceFromBytes(v []byte) (types.SourceType, json.RawMessage, error) {
	topLevel := make(map[string]json.RawMessage)
	err := json.Unmarshal(v, &topLevel)
	if err != nil {
		return "", nil, fmt.Errorf("%w: %w", ErrUnmarshal, err)
	}
	// Check for a type field and unmarshal its value if found.
	rawType, found := topLevel["type"]
	if !found {
		return "", nil, fmt.Errorf("%w: unknown data type", ErrUnmarshal)
	}
	var sourceType types.SourceType
	err = json.Unmarshal(rawType, &sourceType)
	if err != nil {
		return "", nil, fmt.Errorf("%w: %w", ErrUnmarshal, err)
	}
	return sourceType, topLevel["source"], nil
}

func unmarshalSource[T any](v json.RawMessage) (T, error) {
	var source T
	if err := json.Unmarshal(v, &source); err != nil {
		return source, fmt.Errorf("%w: %w", ErrUnmarshal, err)
	}
	return source, nil
}
