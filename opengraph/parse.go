// Copyright 2026 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: 	AGPL-3.0-or-later

package opengraph

import (
	"bytes"
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/go-resty/resty/v2"
	"github.com/immanent-tech/go-syndication/client"
	"golang.org/x/net/html"
)

type parseOptions struct {
	httpClient *resty.Client
}

// ParseOption is a functional option to apply to an Open Graph parsing method.
type ParseOption func(*parseOptions)

// WithClient option specifies an existing http client to use for parsing remote content.
func WithClient(c *resty.Client) ParseOption {
	return func(po *parseOptions) {
		po.httpClient = c
	}
}

// ParseURL will parse the given URL and return any Open Graph metadata found on the page.
func ParseURL(ctx context.Context, pageURL string, options ...ParseOption) (*OpenGraph, error) {
	opts := &parseOptions{}
	for option := range slices.Values(options) {
		option(opts)
	}

	if opts.httpClient == nil {
		opts.httpClient = client.LoadHTTPClient()
	}

	resp, err := opts.httpClient.R().SetContext(ctx).Get(pageURL)
	if err != nil {
		return nil, fmt.Errorf("get url: %w", err)
	}
	if resp.IsError() {
		return nil, fmt.Errorf("%s: %s", resp.Status(), resp.Error())
	}

	return parse(resp.Body())
}

// ParseBytes will parse the given byte array and return any Open Graph metadata found within. Use with existing HTML
// page data.
func ParseBytes(data []byte, options ...ParseOption) (*OpenGraph, error) {
	opts := &parseOptions{}
	for option := range slices.Values(options) {
		option(opts)
	}

	return parse(data)
}

func parse(data []byte) (*OpenGraph, error) {
	htmlData, err := html.Parse(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("parse data: %w", err)
	}

	og := &OpenGraph{
		AdditionalProperties: make(map[string]string),
	}

	visitNode(htmlData, og)

	return og, nil
}

// visitNode recursively walks the node tree, extracting og: meta tags.
// Returns true to signal the caller to stop descending (entered <body>).
func visitNode(n *html.Node, og *OpenGraph) (done bool) {
	if n.Type == html.ElementNode {
		switch n.Data {
		case "body":
			return true
		case "meta":
			extractMeta(n, og)
			return false
		}
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if visitNode(c, og) {
			return true
		}
	}
	return false
}

// extractMeta reads property/name and content attributes from a <meta> node.
func extractMeta(n *html.Node, og *OpenGraph) {
	var property, content string
	for _, a := range n.Attr {
		switch strings.ToLower(a.Key) {
		case "property", "name":
			property = strings.ToLower(a.Val)
		case "content":
			content = a.Val
		}
	}
	if strings.HasPrefix(property, "og:") {
		og.set(property, content)
	}
}
