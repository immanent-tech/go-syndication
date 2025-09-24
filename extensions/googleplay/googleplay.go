// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: 	AGPL-3.0-or-later

package googleplay

import "github.com/immanent-tech/go-syndication/sanitization"

func (c *Category) String() string {
	return sanitization.SanitizeString(c.Text)
}
