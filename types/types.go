// Copyright 2026 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: 	AGPL-3.0-or-later

package types

import "strings"

func (a Link) String() string {
	return a.Value
}

func (p Person) String() string {
	return p.Name
}

func (i Image) String() string {
	return strings.Join([]string{i.GetURL(), i.GetTitle()}, " ")

}
