// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: 	AGPL-3.0-or-later

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
	"github.com/go-playground/validator/v10"
	"github.com/go-resty/resty/v2"
	"golang.org/x/net/html"
	htmlatom "golang.org/x/net/html/atom"

	"github.com/immanent-tech/go-syndication/atom"
	"github.com/immanent-tech/go-syndication/jsonfeed"
	"github.com/immanent-tech/go-syndication/rss"
	"github.com/immanent-tech/go-syndication/types"
)

// DefaultRequestTimeout is the maximum time allowed for a HTTP request issued by the library to execute.
var DefaultRequestTimeout = 30 * time.Second

var (
	// ErrParseBytes indicates an error occurred trying to parse a byte array as a feed.
	ErrParseBytes = errors.New("unable to parse bytes as feed")
	// ErrParseURL indicates an error occurred trying to parse a URL as a feed.
	ErrParseURL = errors.New("unable to parse URL as feed")
	// ErrUnsupportedFormat indicates that feed format is not known and cannot be parsed.
	ErrUnsupportedFormat = errors.New("unsupported feed format")
)

// FeedResult is returned when calling NewFeedsFromURLs and contains the results for parsing an individual URL. It
// will contain the original URL and either a new Feed or a non-nil error.
type FeedResult struct {
	URL  string
	Feed *Feed
	Err  error
}

// FeedItemsResult is returned when calling NewItemsFromURLs and contains the results for parsing an individual URL. It
// will contain the original URL, any items parsed and a non-nil error if a problem occurred.
type FeedItemsResult struct {
	URL   string
	Items []Item
	Err   error
}

// ItemsResult is returned when calling NewItemsFromURLs and contains the results for parsing an individual Feed URL. It
// will contain the original URL and either a slice of Items or a non-nil error.
type ItemsResult []FeedItemsResult

// NewFeedFromBytes will create a new Feed of the given type from the given byte array.
func NewFeedFromBytes[T any](data []byte) (*Feed, error) {
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
	err = feed.Validate()
	if err != nil {
		return nil, fmt.Errorf("%w: feed is not valid: %w", ErrParseBytes, err)
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

// NewFeedFromURL will attempt to create new Feed object from the given URL.
func NewFeedFromURL(ctx context.Context, url string) (*Feed, error) {
	client := newWebClient()
	result := parseFeedURL(ctx, client, url)
	return result.Feed, result.Err
}

// NewFeedsFromURLs will attempt to create new Feed objects from the given list of URLs. It returns a slice containing:
// the URL, any Feed object that was created, else, an non-nil error explaining the problem creating the Feed.
func NewFeedsFromURLs(ctx context.Context, urls ...string) []FeedResult {
	client := newWebClient()

	results := make([]FeedResult, 0, len(urls))
	workerCh := make(chan FeedResult)
	var wg sync.WaitGroup

	go func() {
		defer close(workerCh)
		for url := range slices.Values(urls) {
			wg.Add(1)
			go func(url string) {
				defer wg.Done()
				workerCh <- parseFeedURL(ctx, client, url)
			}(url)
		}
		wg.Wait()
	}()
	// Gather results.
	for result := range workerCh {
		results = append(results, result)
	}

	return results
}

// NewItemsFromURLs will attempt to create new Item objects from the given list of Feed URLs. It returns a slice
// containing: the Feed URL, a slice of Items for that Feed URL, else, an non-nil error explaining the problem fetching
// Items.
func NewItemsFromURLs(ctx context.Context, urls ...string) ItemsResult {
	client := newWebClient()
	results := make(ItemsResult, 0, len(urls))
	workerCh := make(chan FeedItemsResult)
	var wg sync.WaitGroup

	go func() {
		defer close(workerCh)
		for url := range slices.Values(urls) {
			wg.Add(1)
			go func(url string) {
				defer wg.Done()
				if result := parseFeedURL(ctx, client, url); result.Err != nil {
					workerCh <- FeedItemsResult{
						URL: url,
						Err: result.Err,
					}
				} else {
					workerCh <- FeedItemsResult{
						URL:   url,
						Items: result.Feed.GetItems(),
					}
				}
			}(url)
		}
		wg.Wait()
	}()
	// Gather results.
	for result := range workerCh {
		results = append(results, result)
	}

	return results
}

// FindFeedImage will try to find an image to represent the feed. Useful to call if the feed does not define an image
// itself.
func FindFeedImage(ctx context.Context, feed *Feed) error {
	var timeout time.Duration
	if deadline, ok := ctx.Deadline(); !ok {
		timeout = DefaultRequestTimeout
	} else {
		timeout = time.Until(deadline)
	}
	image, err := discoverFeedImage(feed.GetLink(), timeout)
	if err != nil {
		return fmt.Errorf("unable to find feed image: %w", err)
	}
	if image != nil {
		feed.SetImage(image)
	}
	return nil
}

// parseFeedURL attempts to parse the given URL as a feed source.
func parseFeedURL(ctx context.Context, client *resty.Client, url string) FeedResult {
	// Get the feed data.
	resp, err := client.R().
		SetContext(ctx).
		Get(url)
	if err != nil {
		return FeedResult{URL: url, Err: fmt.Errorf("%w: %w", ErrParseURL, err)}
	}
	if resp.IsError() {
		return FeedResult{Err: fmt.Errorf("%w: %s", ErrParseURL, resp.Status())}
	}
	// Retrieve the content header so we know what format we are dealing with.
	content := resp.Header().Get("Content-Type")
	if content == "" {
		return FeedResult{URL: url, Err: fmt.Errorf("%w: missing Content-Type header", ErrParseURL)}
	}
	// Try to parse the response body as a valid feed type.
	var feed *Feed
	switch {
	case isMimeType(content, types.MimeTypesRSS):
		// RSS.
		feed, err = NewFeedFromBytes[*rss.RSS](resp.Body())
	case isMimeType(content, types.MimeTypesAtom):
		// Atom.
		feed, err = NewFeedFromBytes[*atom.Feed](resp.Body())
	case isMimeType(content, types.MimeTypesIndeterminate):
		// Likely a feed but mimetype is ambiguous. Try to find a relevant starting type for the specific type.
		switch {
		case bytes.Contains(resp.Body(), []byte("<feed")):
			feed, err = NewFeedFromBytes[*atom.Feed](resp.Body())
			if err != nil && errors.Is(err, &validator.InvalidValidationError{}) {
				return FeedResult{Err: fmt.Errorf("could not parse as atom: %w", err)}
			}
		case bytes.Contains(resp.Body(), []byte("<rss")):
			feed, err = NewFeedFromBytes[*rss.RSS](resp.Body())
			if err != nil && errors.Is(err, &validator.InvalidValidationError{}) {
				return FeedResult{Err: fmt.Errorf("could not parse as rss: %w", err)}
			}
		}
	case isMimeType(content, types.MimeTypesJSONFeed):
		// JSONFeed
		feed, err = NewFeedFromBytes[*jsonfeed.Feed](resp.Body())
	case isMimeType(content, types.MimeTypesHTML):
		// URL points to a HTML page, not a feed source.
		// Try to find a feed link on the page and then parse that URL.
		if newURL, err := discoverFeedURL(url, resp.Body()); err == nil && newURL != "" {
			return parseFeedURL(ctx, client, newURL)
		}
		fallthrough
	default:
		// Cannot determine or unsupported content.
		err = fmt.Errorf("%w: unsupported feed media type: %s", ErrParseURL, content)
	}

	// (╯°益°)╯彡┻━┻
	if err != nil {
		return FeedResult{URL: url, Err: err}
	}

	// If the source URL is not set, set it.
	if feed.GetSourceURL() == "" || feed.GetSourceURL() != url {
		feed.SetSourceURL(url)
	}

	return FeedResult{URL: url, Feed: feed}
}

// discoverFeedImage attempts to find a suitable image to use for a feed.
func discoverFeedImage(feed string, timeout time.Duration) (*types.ImageInfo, error) {
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

// discoverFeedURL attempts to find a feed URL within a HTML page.
//
// There are a couple of "canonical" places the feed URL is located. Firstly, as per the RSS spec, look for a link
// element with rel="alternate" and type="application/rss+xml". Secondly, check for a link element with a URL that ends
// with feed, rss or atom, which would indicate a feed URL.
//
//nolint:gocognit,funlen
func discoverFeedURL(path string, content []byte) (string, error) {
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

func newWebClient() *resty.Client {
	return resty.New().SetHeader("User-Agent", "go-syndication")
}
