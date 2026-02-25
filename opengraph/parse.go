// Copyright 2026 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: 	AGPL-3.0-or-later

package opengraph

import (
	"context"
	"encoding/xml"
	"fmt"
	"slices"
	"strings"

	"github.com/go-resty/resty/v2"
	"github.com/immanent-tech/go-syndication/client"
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

	// Set up a reader to just read the head element.
	headReader := client.NewHeadReader(strings.NewReader(string(resp.Body())), 256*1024)

	// Set up the xml decoder.
	dec := xml.NewDecoder(headReader)
	dec.Strict = false
	dec.AutoClose = xml.HTMLAutoClose
	dec.Entity = xml.HTMLEntity

	// Advance to the <head> element.
	for {
		tok, err := dec.Token()
		if err != nil {
			return nil, fmt.Errorf("decode xml: %w", err)
		}
		se, ok := tok.(xml.StartElement)
		if !ok {
			continue
		}
		var og OpenGraph
		if err := og.UnmarshalXML(dec, se); err != nil {
			return nil, fmt.Errorf("unmarshal xml: %w", err)
		}
		return &og, nil
	}
}
