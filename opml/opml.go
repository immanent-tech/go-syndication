// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: 	AGPL-3.0-or-later

package opml

import (
	"bytes"
	"fmt"
	"slices"
	"time"

	"encoding/xml"

	"github.com/immanent-tech/go-syndication/types"
	"golang.org/x/net/html/charset"
)

// NewOPMLFromBytes generates an OPML object from the given byte array.
func NewOPMLFromBytes(b []byte) (*OPML, error) {
	var root OPML

	reader := bytes.NewReader(b)
	decoder := xml.NewDecoder(reader)
	decoder.CharsetReader = charset.NewReaderLabel
	err := decoder.Decode(&root)
	if err != nil {
		return nil, fmt.Errorf("could not decode byte array to OPML: %w", err)
	}

	return &root, nil
}

// NewOPML creates a new OPML object.
func NewOPML(options ...Option) *OPML {
	opml := &OPML{
		Version: "2.0",
		Head: Head{
			DateCreated: types.DateTime{Time: time.Now()},
		},
	}

	for option := range slices.Values(options) {
		option(opml)
	}

	return opml
}

// Option is a functional option to apply to an OPML object.
type Option func(*OPML)

// WithTitle option sets a title for the OPML object.
func WithTitle(title string) Option {
	return func(o *OPML) {
		o.Head.Title = title
	}
}

// WithOutlines option appends the given outlines to the OPML object.
func WithOutlines(outlines ...Outline) Option {
	return func(o *OPML) {
		o.Body = append(o.Body, outlines...)
	}
}
