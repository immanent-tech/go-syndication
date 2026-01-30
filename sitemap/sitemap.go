// Copyright 2026 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: 	AGPL-3.0-or-later

package sitemap

import "encoding/xml"

func NewURLSet(urls ...URL) *URLSet {
	return &URLSet{
		XMLName: xml.Name{
			Space: "http://www.sitemaps.org/schemas/sitemap/0.9",
			Local: "urlset",
		},
		URLs: urls,
	}
}
