// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: 	MIT

package feeds

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"mime"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"sync"
	"time"

	"codeberg.org/readeck/go-readability/v2"
	"github.com/go-resty/resty/v2"
	"golang.org/x/net/html"
	htmlatom "golang.org/x/net/html/atom"

	"github.com/immanent-tech/go-syndication/atom"
	"github.com/immanent-tech/go-syndication/jsonfeed"
	"github.com/immanent-tech/go-syndication/rss"
	"github.com/immanent-tech/go-syndication/types"
	"github.com/immanent-tech/go-syndication/validation"
)

var (
	// DefaultRequestTimeout is the maximum time allowed for a HTTP request issued by the library to execute.
	DefaultRequestTimeout = 30 * time.Second
	// UserAgent is the user agent that is sent when making http requests. Change this as needed.
	UserAgent = "go-syndication (+https://github.com/immanent-tech/go-syndication)"
)

var (
	// ErrParseBytes indicates an error occurred trying to parse a byte array as a feed.
	ErrParseBytes = errors.New("unable to parse bytes as feed")
	// ErrParseURL indicates an error occurred trying to parse a URL as a feed.
	ErrParseURL = errors.New("unable to parse URL as feed")
	// ErrUnsupportedFormat indicates that feed format is not known and cannot be parsed.
	ErrUnsupportedFormat = errors.New("unsupported feed format")
)

type ParserOptions struct {
	validate bool
	client   *resty.Client
}

type ParseOption func(*ParserOptions)

// PerformValidation option controls whether validation will be performed after the feed data has been parsed.
// This allows a feed that doesn't strictly pass its format specification to be returned.
func PerformValidation(value bool) ParseOption {
	return func(po *ParserOptions) {
		po.validate = value
	}
}

// WithClient option allows using a custom client for any network requests to fetch feed resources.
func WithClient(client *resty.Client) ParseOption {
	return func(po *ParserOptions) {
		po.client = client
	}
}

// NewFeedFromBytes will create a new Feed of the given type from the given byte array.
func NewFeedFromBytes[T any](data []byte, options ...ParseOption) (*Feed, error) {
	// Parse and set options.
	opts := &ParserOptions{}
	for option := range slices.Values(options) {
		option(opts)
	}

	var (
		original T
		feed     *Feed
		err      error
	)
	if _, ok := any(original).(*jsonfeed.Feed); ok {
		// If the original is JSONFeed, unmarshal as JSON.
		err = json.Unmarshal(data, &original)
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
	if opts.validate {
		if err = feed.Validate(); err != nil {
			return nil, fmt.Errorf("%w: feed is not valid: %w", ErrParseBytes, err)
		}
	}

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

// NewFeedFromURL attempts to parse the given URL as a feed source.
func NewFeedFromURL(ctx context.Context, feedURL string, options ...ParseOption) (*Feed, error) {
	// Parse and set options.
	opts := &ParserOptions{}
	for option := range slices.Values(options) {
		option(opts)
	}
	// Use the default client if one is not specified.
	if opts.client == nil {
		opts.client = newWebClient()
	}

	// Get the feed data.
	resp, err := opts.client.R().
		SetContext(ctx).
		Get(feedURL)
	switch {
	case err != nil || resp.IsError():
		return nil, &HTTPError{Code: resp.StatusCode(), Message: resp.Status()}
	case err != nil:
		return nil, &HTTPError{Code: http.StatusInternalServerError, Message: err.Error()}
	}

	// Retrieve the content header so we know what format we are dealing with.
	content := resp.Header().Get("Content-Type")
	if content == "" {
		return nil, fmt.Errorf("%w: missing Content-Type header", ErrParseURL)
	}

	// Try to parse the response body as a valid feed type.
	var feed *Feed
	switch {
	case isMimeType(content, types.MimeTypesRSS):
		// RSS.
		feed, err = NewFeedFromBytes[*rss.RSS](resp.Body(), options...)
		if err != nil {
			return nil, fmt.Errorf("could not parse as rss: %w", err)
		}
	case isMimeType(content, types.MimeTypesAtom):
		// Atom.
		feed, err = NewFeedFromBytes[*atom.Feed](resp.Body(), options...)
		if err != nil {
			return nil, fmt.Errorf("could not parse as atom: %w", err)
		}
	case isMimeType(content, types.MimeTypesIndeterminate):
		// Likely a feed but mimetype is ambiguous. Try to find a relevant starting type for the specific type.
		switch {
		case bytes.Contains(resp.Body(), []byte("<feed")):
			feed, err = NewFeedFromBytes[*atom.Feed](resp.Body(), options...)
			if err != nil && errors.Is(err, &validation.StructError{}) {
				return nil, fmt.Errorf("could not parse as atom: %w", err)
			}
		case bytes.Contains(resp.Body(), []byte("<rss")):
			feed, err = NewFeedFromBytes[*rss.RSS](resp.Body(), options...)
			if err != nil && errors.Is(err, &validation.StructError{}) {
				return nil, fmt.Errorf("could not parse as rss: %w", err)
			}
		default:
			return nil, fmt.Errorf("%w: unsupported feed media type: %s", ErrParseURL, content)
		}
	case isMimeType(content, types.MimeTypesJSONFeed):
		// JSONFeed
		feed, err = NewFeedFromBytes[*jsonfeed.Feed](resp.Body(), options...)
		if err != nil {
			return nil, fmt.Errorf("could not parse as jsonfeed: %w", err)
		}
	case isMimeType(content, types.MimeTypesHTML):
		// URL points to a HTML page, not a feed source.
		// Try to find a feed link on the page and then parse that URL.
		if newURL, err := DiscoverFeedURL(feedURL, resp.Body()); err == nil && newURL != "" {
			return NewFeedFromURL(ctx, newURL)
		}
		return nil, fmt.Errorf("could not find a feed URL on page: %s", feedURL)
	default:
		// Cannot determine or unsupported content.
		return nil, fmt.Errorf("%w: unsupported feed media type: %s", ErrParseURL, content)
	}

	// Handle getting through the switch but still not parsing the content.
	if feed == nil {
		return nil, fmt.Errorf("%w: %w", ErrParseURL, err)
	}

	// If the source URL is not set, set it.
	if feed.GetSourceURL() == "" || feed.GetSourceURL() != feedURL {
		feed.SetSourceURL(feedURL)
	}

	return feed, nil
}

// DiscoverFeedImage attempts to find a suitable image to use for a feed.
func DiscoverFeedImage(feed string, timeout time.Duration) (*types.ImageInfo, error) {
	// Parse feed string as URL.
	sourceURL, err := url.Parse(feed)
	if err != nil {
		return nil, fmt.Errorf("discover feed image: failed to parse url %s: %w", feed, err)
	}
	// Parse URL and extract readability content.
	page, err := readability.FromURL(sourceURL.String(), timeout)
	if err != nil {
		return nil, fmt.Errorf("discover feed image: failed to parse url %s: %w", feed, err)
	}
	// Determine best image from readability content.
	var img string
	switch {
	case page.ImageURL() != "":
		img = page.ImageURL()
	case page.Favicon() != "":
		img = page.Favicon()
	default:
		return nil, fmt.Errorf("discover feed image: %s: %d", feed, http.StatusNotFound)
	}
	// Parse the image string as a URL.
	imgURL, err := url.Parse(img)
	if err != nil {
		return nil, fmt.Errorf("discover feed image: failed to parse image url %s: %w", img, err)
	}
	// If the image URL is not absolute, assume it is based on the feed base URL and generate a new URL as appropriate.
	if !imgURL.IsAbs() {
		sourceURL.Path = imgURL.String()
		return &types.ImageInfo{URL: sourceURL.String()}, nil
	}
	return &types.ImageInfo{URL: imgURL.String()}, nil
}

// DiscoverFeedURL attempts to find a feed URL within a HTML page.
//
// There are a couple of "canonical" places the feed URL is located. Firstly, as per the RSS spec, look for a link
// element with rel="alternate" and type="application/rss+xml". Secondly, check for a link element with a URL that ends
// with feed, rss or atom, which would indicate a feed URL.
//
//nolint:gocognit,funlen
func DiscoverFeedURL(path string, content []byte) (string, error) {
	pageURL, err := url.Parse(path)
	if err != nil {
		return "", fmt.Errorf("discover feed url: %w", err)
	}

	page := html.NewTokenizer(bytes.NewReader(content))
	for {
		tt := page.Next()
		var feedURL *url.URL
		switch tt { //nolint:exhaustive // we don't want to check evey token.
		case html.ErrorToken:
			return "", fmt.Errorf("discover feed url: %w", page.Err())
		case html.SelfClosingTagToken:
			tkn := page.Token()
			if tkn.DataAtom != htmlatom.Link {
				continue
			}
			var found bool
			// Canonical link with appropriate attributes.
			if slices.ContainsFunc(tkn.Attr,
				func(a html.Attribute) bool { return a.Key == "rel" && a.Val == "alternate" }) &&
				slices.ContainsFunc(
					tkn.Attr,
					func(a html.Attribute) bool { return a.Key == "type" && slices.Contains(types.MimeTypesFeed, a.Val) },
				) {
				found = true
			}
			// Link ends in feed type...
			if slices.ContainsFunc(tkn.Attr,
				func(a html.Attribute) bool {
					if a.Key == "href" && slices.Contains([]string{"feed", "rss", "atom"}, a.Val) {
						return true
					}
					return false
				}) {
				found = true
			}
			// Dont' continue if no feed URL found.
			if !found {
				continue
			}
			idx := slices.IndexFunc(tkn.Attr, func(a html.Attribute) bool {
				return a.Key == "href"
			})
			if idx == 0 {
				continue
			}
			feedURL, err = url.Parse(tkn.Attr[idx].Val)
			if err != nil {
				return tkn.Attr[idx].Val, fmt.Errorf("discover feed url: %w", err)
			}
		case html.StartTagToken:
			tkn := page.Token()
			if tkn.Data != "a" {
				continue
			}
			var found bool
			// Link ends in feed...
			if slices.ContainsFunc(tkn.Attr,
				func(a html.Attribute) bool { return a.Key == "href" && strings.HasSuffix(a.Val, "feed") }) {
				found = true
			}
			if !found {
				continue
			}
			// Get the link.
			idx := slices.IndexFunc(
				tkn.Attr,
				func(a html.Attribute) bool { return a.Key == "href" && strings.HasSuffix(a.Val, "feed") },
			)
			if idx == -1 {
				continue
			}
			feedURL, err = url.Parse(tkn.Attr[idx].Val)
			if err != nil {
				return tkn.Attr[idx].Val, fmt.Errorf("discover feed url: %w", err)
			}
		}
		if feedURL == nil {
			continue
		}
		if !feedURL.IsAbs() {
			// Try to create an absolute URL for the feed.
			fullPath, err := url.JoinPath("/", feedURL.Path)
			if err != nil {
				return "", fmt.Errorf("discover feed url: %w", err)
			}
			pageURL.Path = fullPath
			return pageURL.String(), nil
		}
		return feedURL.String(), nil
	}
}

// isMimeType returns a boolean indicating whether the Content Type header string is included in the given mimeType
// slice.
func isMimeType(content string, mimeTypes []string) bool {
	mediatype, _, err := mime.ParseMediaType(content)
	if err != nil {
		return false
	}
	return slices.Contains(mimeTypes, mediatype)
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

var newWebClient = sync.OnceValue(func() *resty.Client {
	return resty.New().SetHeader("User-Agent", UserAgent)
})

type HTTPError struct {
	Code    int
	Message string
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("%d: %s", e.Code, e.Message)
}
