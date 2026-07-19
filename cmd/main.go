// Copyright 2026 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: 	MIT

package main

import (
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/signal"
	"slices"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/alecthomas/kong"
	"github.com/go-resty/resty/v2"
	feeds "github.com/immanent-tech/go-syndication"
	"github.com/immanent-tech/go-syndication/atom"
	"github.com/immanent-tech/go-syndication/rss"
	"github.com/immanent-tech/go-syndication/types"
)

type Globals struct{}

// CLI contains all options and commands.
type CLI struct {
	Globals

	Fetch FetchCMD `cmd:"" help:"Fetch a feed from a URL"`
	Parse ParseCMD `cmd:"" help:"Parse a feed file"`
}

func init() {
	// Following is copied from https://git.kernel.org/pub/scm/libs/libcap/libcap.git/tree/goapps/web/web.go
	// ensureNotEUID aborts the program if it is running setuid something, or being invoked by root.

	if euid, uid, egid, gid := syscall.Geteuid(), syscall.Getuid(), syscall.Getegid(), syscall.Getgid(); uid != euid ||
		gid != egid ||
		uid == 0 {
		panic(errors.New("foragd should not be run with additional privileges or as root"))
	}
}

func main() {
	cli := CLI{
		Globals: Globals{},
	}

	ctx := kong.Parse(&cli,
		kong.Name("Go Syndication CLI"),
		kong.Description(
			"Go Syndication CLI provides a way to view and generate syndicated formats (e.g., RSS, Atom).",
		),
		kong.UsageOnError(),
	)

	err := ctx.Run(cli.Globals)
	ctx.FatalIfErrorf(err)
}

var client *resty.Client

var LoadHTTPClient = sync.OnceValue(func() *resty.Client {
	client = resty.New().
		SetHeader("User-Agent", "go-syndication").
		SetHeader("Accept", "*/*").
		SetHeader("Accept-Encoding", "gzip, deflate")
	return client
})

// FetchCMD command will fetch a feed at the given URL and display it.
type FetchCMD struct {
	URL string `arg:"" help:"The URL of the feed"`
}

func (c *FetchCMD) Run() error {
	// Set up context.
	ctx, cancelFunc := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancelFunc()

	// Parse the URL to ensure its valid.
	sourceURL, err := url.Parse(c.URL)
	if err != nil {
		return fmt.Errorf("could not parse URL: %w", err)
	}

	// Fetch feed.
	resp, err := LoadHTTPClient().R().
		SetContext(ctx).
		SetDoNotParseResponse(true).
		Get(sourceURL.String())
	switch {
	case err != nil:
		return fmt.Errorf("fetch feed: %w", err)
	case resp.IsError():
		return fmt.Errorf("fetch feed response: %s", resp.Status())
	}
	defer resp.RawBody().Close()

	// Read response data into buffer.
	var feedBuf bytes.Buffer
	if resp.Header().Get("Content-Encoding") == "gzip" {
		// For gzipped response, uncompress first.
		reader, err := gzip.NewReader(resp.RawBody())
		if err != nil {
			return fmt.Errorf("read gzip response: %w", err)
		}
		defer reader.Close()
		const maxBodySize = 10 * 1024 * 1024 // 10 MB limit
		limitReader := io.LimitReader(reader, maxBodySize)
		if _, err := io.Copy(&feedBuf, limitReader); err != nil {
			return fmt.Errorf("read response: %w", err)
		}
	} else {
		// Read response directly.
		if _, err := io.Copy(&feedBuf, resp.RawBody()); err != nil {
			return fmt.Errorf("read response: %w", err)
		}
	}

	feed, err := parseFeedData(&feedBuf)
	if err != nil {
		return fmt.Errorf("parse feed data: %w", err)
	}
	showFeedDetails(feed)

	return nil
}

// ParseCMD command will parse the given file as a feed and display it.
type ParseCMD struct {
	File string `arg:"" help:"The file contaning the feed"`
}

func (c *ParseCMD) Run() error {
	fileBuf, err := os.Open(c.File)
	if err != nil {
		return fmt.Errorf("open file: %w", err)
	}

	feed, err := parseFeedData(fileBuf)
	if err != nil {
		return fmt.Errorf("parse feed data: %w", err)
	}

	showFeedDetails(feed)

	return nil
}

func parseFeedData(r io.Reader) (*feeds.Feed, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("read data: %w", err)
	}

	// Parse the response as a feed type.
	var feedData *feeds.Feed
	switch feedType, err := feeds.DetectSourceType(bytes.NewReader(data)); {
	case err != nil:
		return nil, fmt.Errorf("detect feed type: %w", err)
	case feedType == types.SourceTypeUnknown:
		return nil, errors.New("cannot determine feed type")
	case feedType == types.SourceTypeAtom:
		// Atom feed.
		feedData, err = feeds.NewDecoder[*atom.Feed](bytes.NewReader(data))
		if err != nil {
			return nil, fmt.Errorf("parse atom: %w", err)
		}
	case feedType == types.SourceTypeRSS:
		// RSS feed.
		feedData, err = feeds.NewDecoder[*rss.RSS](bytes.NewReader(data))
		if err != nil {
			return nil, fmt.Errorf("parse rss: %w", err)
		}
	case feedType == types.SourceTypeHTML:
		return nil, errors.New("go html, not feed format")
	default:
		return nil, errors.New("unsupported media type")
	}

	// Handle getting through the switch but still not parsing the content.
	if feedData == nil {
		return nil, errors.New("no feed data returned")
	}

	return feedData, nil
}

func showFeedDetails(feed *feeds.Feed) {
	var str strings.Builder

	str.WriteString("Feed: ")
	str.WriteString(feed.GetTitle())
	str.WriteRune('\n')
	str.WriteString("Link: ")
	str.WriteString(feed.GetLink())
	str.WriteRune('\n')
	str.WriteString("Type: ")
	str.WriteString(string(feed.SourceType))
	str.WriteRune('\n')
	if feed.GetDescription() != "" {
		str.WriteString("Description:")
		str.WriteRune('\n')
		str.WriteString(feed.GetDescription())
		str.WriteRune('\n')
	}
	if feed.GetPublishedDate() != nil {
		str.WriteString("Created: ")
		str.WriteString(feed.GetPublishedDate().Format(time.DateTime))
		str.WriteRune('\n')
	}
	str.WriteString("Updated: ")
	str.WriteString(feed.GetUpdatedDate().Format(time.DateTime))
	str.WriteRune('\n')
	if len(feed.GetCategories()) > 0 {
		str.WriteString("Categories: ")
		str.WriteString(strings.Join(feed.GetCategories(), ","))
		str.WriteRune('\n')
	}
	if feed.GetImage() != nil {
		str.WriteString("Image: ")
		str.WriteString(feed.GetImage().URL)
	}
	str.WriteRune('\n')
	str.WriteRune('\n')

	for item := range slices.Values(feed.GetItems()) {
		str.WriteString("---")
		str.WriteRune('\n')
		if item.GetID() != "" {
			str.WriteString("Item ID: ")
			str.WriteString(item.GetID())
			str.WriteRune('\n')
		}
		str.WriteString("Title: ")
		str.WriteString(item.GetTitle())
		str.WriteRune('\n')
		str.WriteString("Link: ")
		str.WriteString(item.GetLink())
		str.WriteRune('\n')
		if len(item.GetAuthors()) > 0 {
			str.WriteString("Authors: ")
			str.WriteString(strings.Join(item.GetAuthors(), ","))
			str.WriteRune('\n')
		}
		if item.GetDescription() != "" {
			str.WriteString("Description:")
			str.WriteRune('\n')
			str.WriteString(item.GetDescription())
			str.WriteRune('\n')
		}
		if item.GetPublishedDate() != nil {
			str.WriteString("Published: ")
			str.WriteString(item.GetPublishedDate().Format(time.DateTime))
			str.WriteRune('\n')
		}
		if len(item.GetCategories()) > 0 {
			str.WriteString("Categories: ")
			str.WriteString(strings.Join(item.GetCategories(), ","))
			str.WriteRune('\n')
		}
		if item.GetImage() != nil {
			str.WriteString("Image: ")
			str.WriteString(item.GetImage().URL)
		}
		if item.GetContent() != nil {
			str.WriteString("Content:")
			str.WriteRune('\n')
			str.WriteString(*item.GetContent())
		}
		str.WriteRune('\n')
	}

	fmt.Fprintf(os.Stdout, "%s", str.String())
}
