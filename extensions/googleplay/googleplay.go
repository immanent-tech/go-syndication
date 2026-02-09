// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: 	MIT

package googleplay

import "github.com/immanent-tech/go-syndication/sanitization"

func (c *Category) String() string {
	if c != nil {
		return sanitization.SanitizeString(c.Text)
	}
	return ""
}
