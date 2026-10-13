// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: 	MIT

package syndication

import "bytes"

func (p SYUpdatePeriod) MarshalText() ([]byte, error) {
	return []byte(p), nil
}

func (p *SYUpdatePeriod) UnmarshalText(data []byte) error {
	cleaned := bytes.TrimSpace(data)
	value := SYUpdatePeriod(string(cleaned))
	*p = value
	return nil
}
