// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: 	MIT

package feeds

import (
	"bufio"
	"bytes"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

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
func parseSource[T any](source T) types.SourceType {
	switch any(source).(type) {
	case *atom.Feed:
		return types.SourceTypeAtom
	case *rss.RSS:
		return types.SourceTypeRSS
	default:
		return ""
	}
}

// DetectSourceType determines the feed source by extracting key signatures from the data. It can detect supported feed
// formats as well as HTML.
func DetectSourceType(r io.Reader) (types.SourceType, error) {
	data := bufio.NewReader(r)

	// Peek enough bytes for content sniffing without consuming the reader.
	peek, err := data.Peek(512)
	if err != nil {
		return types.SourceTypeUnknown, fmt.Errorf("peek at source file: %w", err)
	}

	if looksLikeHTML(peek) {
		return types.SourceTypeHTML, nil
	}

	// Fall back to XML-based root element detection for feeds (and XHTML).
	return detectFeedSourceType(data)
}

func looksLikeHTML(peek []byte) bool {
	// http.DetectContentType implements the WHATWG sniffing algorithm and
	// recognizes common HTML signatures (DOCTYPE, <html>, <head>, <script>, etc.)
	if ct := http.DetectContentType(peek); strings.HasPrefix(ct, "text/html") {
		return true
	}

	// Belt-and-suspenders manual check, in case leading whitespace/BOM/comments
	// push the signature past DetectContentType's window or it's ambiguous.
	trimmed := bytes.TrimSpace(peek)
	lower := bytes.ToLower(trimmed)
	return bytes.HasPrefix(lower, []byte("<!doctype html")) ||
		bytes.HasPrefix(lower, []byte("<html"))
}

func detectFeedSourceType(r io.Reader) (types.SourceType, error) {
	decoder := xml.NewDecoder(r)
	decoder.Strict = false // be lenient with malformed feeds in the wild

	for {
		tok, err := decoder.Token()
		if errors.Is(err, io.EOF) {
			return types.SourceTypeUnknown, fmt.Errorf("%w: no root element found", ErrParseBytes)
		}
		if err != nil {
			return types.SourceTypeUnknown, fmt.Errorf("decode source: %w", err)
		}

		if startElement, ok := tok.(xml.StartElement); ok {
			switch {
			case startElement.Name.Local == "rss":
				return types.SourceTypeRSS, nil
			case startElement.Name.Local == "feed" && startElement.Name.Space == "http://www.w3.org/2005/Atom":
				return types.SourceTypeAtom, nil
			case startElement.Name.Local == "feed": // some feeds omit/misdeclare namespace
				return types.SourceTypeAtom, nil
			case startElement.Name.Local == "RDF":
				return types.SourceTypeRDF, nil
			default:
				return types.SourceTypeUnknown, fmt.Errorf("unrecognized root element: %s", startElement.Name.Local)
			}
		}
	}
}
