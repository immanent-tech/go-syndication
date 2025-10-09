// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: 	AGPL-3.0-or-later

package feeds

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"mime"
	"net/url"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/go-shiori/go-readability"
	"golang.org/x/net/html"
	htmlatom "golang.org/x/net/html/atom"

	"github.com/immanent-tech/go-syndication/atom"
	"github.com/immanent-tech/go-syndication/jsonfeed"
	"github.com/immanent-tech/go-syndication/rss"
	"github.com/immanent-tech/go-syndication/types"
)

var (
	// ErrParseFeed indicates an error parsing the feed content.
	ErrParseFeed  = errors.New("unable to parse")
	ErrParseImage = errors.New("unable to parse an image")
	// ErrUnmarshal indicates an error unmarshaling the feed from its native format.
	ErrUnmarshal = errors.New("unable to unmarshal")
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
		return nil, fmt.Errorf("%w: %w", ErrParseFeed, err)
	}
	source, ok := any(original).(types.FeedSource)
	if !ok {
		return nil, fmt.Errorf("%w: data is not a valid feed type %T", ErrParseFeed, original)
	}
	feed = &Feed{
		FeedSource: source,
	}
	feed.SourceType = parseSource(original)
	err = feed.Validate()
	if err != nil {
		return nil, fmt.Errorf("%w: feed is not valid: %w", ErrParseFeed, err)
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
				result := parseFeedURL(ctx, client, url)
				if result.Err != nil {
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
	deadline, ok := ctx.Deadline()
	if !ok {
		timeout = 30 * time.Second
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
	resp, err := client.R().SetContext(ctx).Get(url)
	if err != nil {
		return FeedResult{URL: url, Err: fmt.Errorf("%w: could not access feed URL: %w", ErrParseFeed, err)}
	}
	// Retrieve the content header so we know what format we are dealing with.
	content := resp.Header().Get("Content-Type")
	if content == "" {
		return FeedResult{URL: url, Err: fmt.Errorf("%w: unable to determine feed type", ErrParseFeed)}
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
		// Likely a feed but mimetype is ambiguous.
		// Try RSS first...
		feed, err = NewFeedFromBytes[*rss.RSS](resp.Body())
		if err != nil {
			slog.DebugContext(ctx, "Failed to parse indeterminate feed as RSS.",
				slog.Any("error", err),
				slog.String("url", url),
			)
		}
		// Try Atom if that failed...
		if err != nil {
			feed, err = NewFeedFromBytes[*atom.Feed](resp.Body())
		}
		if err != nil {
			slog.DebugContext(ctx, "Failed to parse indeterminate feed as Atom.",
				slog.Any("error", err),
				slog.String("url", url),
			)
		}
	case isMimeType(content, types.MimeTypesJSONFeed):
		// JSONFeed
		feed, err = NewFeedFromBytes[*jsonfeed.Feed](resp.Body())
	case isMimeType(content, types.MimeTypesHTML):
		// URL points to a HTML page, not a feed source.
		// Try to find a feed link on the page and then parse that URL.
		url, err := discoverFeedURL(url, resp.Body())
		if err == nil && url != "" {
			return parseFeedURL(ctx, client, url)
		}
		fallthrough
	default:
		// Cannot determine or unsupported content.
		err = fmt.Errorf("%w: %s", ErrUnsupportedFormat, content)
	}

	// (╯°益°)╯彡┻━┻
	if err != nil {
		return FeedResult{URL: url, Err: err}
	}

	// If the source URL is not set, set it.
	if feed.GetSourceURL() == "" || feed.GetSourceURL() != url {
		feed.AddLink(url)
	}

	return FeedResult{URL: url, Feed: feed}
}

// discoverFeedImage attempts to find a suitable image to use for a feed.
func discoverFeedImage(sourceURL string, timeout time.Duration) (*types.ImageInfo, error) {
	page, err := readability.FromURL(sourceURL, timeout)
	if err != nil {
		return nil, fmt.Errorf("failed to parse %s, %w", sourceURL, err)
	}
	switch {
	case page.Image != "":
		return &types.ImageInfo{URL: page.Image}, nil
	case page.Favicon != "":
		return &types.ImageInfo{URL: page.Favicon}, nil
	default:
		return nil, ErrParseImage
	}
}

// discoverFeedURL attempts to find a feed URL within a HTML page.
func discoverFeedURL(path string, content []byte) (string, error) {
	pageURL, err := url.Parse(path)
	if err != nil {
		return "", fmt.Errorf("unable to parse url: %w", err)
	}

	page := html.NewTokenizer(bytes.NewReader(content))
	for {
		tt := page.Next()
		switch tt {
		case html.ErrorToken:
			return "", fmt.Errorf("unable to determine feed url: %w", page.Err())
		case html.SelfClosingTagToken:
			tkn := page.Token()
			if tkn.DataAtom != htmlatom.Link {
				continue
			}
			// "rel" attribute must have a value of "alternate".
			if !slices.ContainsFunc(tkn.Attr, func(a html.Attribute) bool { return a.Key == "rel" && a.Val == "alternate" }) {
				continue
			}
			// "type" attribute must contain valid feed MIME type.
			idx := slices.IndexFunc(tkn.Attr, func(a html.Attribute) bool {
				return a.Key == "type" && slices.Contains(types.MimeTypesFeed, a.Val)
			})
			if idx == 0 {
				continue
			}
			// is a feed url, extract the url.
			idx = slices.IndexFunc(tkn.Attr, func(a html.Attribute) bool {
				return a.Key == "href"
			})
			if idx == 0 {
				continue
			}
			feedURL, err := url.Parse(tkn.Attr[idx].Val)
			if err != nil {
				return tkn.Attr[idx].Val, fmt.Errorf("found URL but unable to parse: %w", err)
			}
			if !feedURL.IsAbs() {
				// Try to create an absolute URL for the feed.
				fullPath, err := url.JoinPath("/", feedURL.Path)
				if err != nil {
					return "", fmt.Errorf("failed to generate feed URL: %w", err)
				}
				pageURL.Path = fullPath
				return pageURL.String(), nil
			}
			return feedURL.String(), nil
		}
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
	// Set the mimetypes we accept. Who knows if this helps but at least we are honest to the server with what mimetypes
	// we want.
	mimeTypes := types.MimeTypesFeed
	mimeTypes = append(mimeTypes, ";q=0.2,*/*", ";q=0.1")
	// Return a client.
	return resty.New().
		SetHeader("Accept", strings.Join(mimeTypes, ","))
}
