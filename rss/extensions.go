// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: 	AGPL-3.0-or-later

package rss

import "github.com/joshuar/go-syndication/sanitization"

func (c *ContentEncoded) String() string {
	return sanitization.SanitizeString(c.Value)
}
