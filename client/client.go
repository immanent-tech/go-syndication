// Copyright 2026 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: 	AGPL-3.0-or-later

package client

import (
	"io"
	"strings"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"
)

var (
	// DefaultRequestTimeout is the maximum time allowed for a HTTP request issued by the library to execute.
	DefaultRequestTimeout = 30 * time.Second
	// UserAgent is the user agent that is sent when making http requests. Change this as needed.
	UserAgent = "go-syndication (+https://github.com/immanent-tech/go-syndication)"
)

var client *resty.Client

// LoadHTTPClient creates a new http client for the package. It only does initialisation once, the client is then reused
// with each subsquent call.
var LoadHTTPClient = sync.OnceValue(func() *resty.Client {
	client = resty.New().
		SetHeader("User-Agent", UserAgent).
		SetHeader("Accept", "*/*").
		SetHeader("Accept-Encoding", "gzip, deflate")
	return client
})

// HeadReader wraps an io.Reader and is used specifically for reading the <head> element in HTML.
type HeadReader struct {
	r       io.Reader
	buf     []byte
	done    bool
	total   int
	maxRead int
}

// NewHeadReader creates a new HeadReader from the given io.Reader.
func NewHeadReader(r io.Reader, maxBytes int) *HeadReader {
	return &HeadReader{r: r, maxRead: maxBytes}
}

// Read reads the <head> tag from the reader (if any) from the given byte array.
func (h *HeadReader) Read(src []byte) (int, error) {
	if h.done {
		return 0, io.EOF
	}
	if h.total >= h.maxRead {
		return 0, io.EOF
	}
	n, err := h.r.Read(src)
	h.total += n
	// Look for </head> in what we just read to stop early
	chunk := strings.ToLower(string(src[:n]))
	if idx := strings.Index(chunk, "</head>"); idx != -1 {
		h.done = true
		return idx + len("</head>"), io.EOF
	}
	return n, err
}
