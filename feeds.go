// Copyright 2026 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: 	MIT

package feeds

func (a Link) String() string {
	return a.Value
}

func (p Person) String() string {
	return p.Name
}
