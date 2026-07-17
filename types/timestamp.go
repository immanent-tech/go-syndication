// Copyright 2026 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: 	AGPL-3.0-or-later

package types

import "time"

func NewTimestamp(ts time.Time) Timestamp {
	return Timestamp{
		Value: ts,
	}
}

// MarshalText implements the encoding.TextMarshaler interface. Serializes Timestamp to a plain byte slice. Uses RFC822
// for widest compatibility.
func (t Timestamp) MarshalText() ([]byte, error) {
	return []byte(t.Value.Format("02 Jan 2006 15:04 -0700")), nil
}

// UnmarshalText will unmarshal/parse a Timestamp from the given string.
func (t *Timestamp) UnmarshalText(data []byte) error {
	t.Value = tryFormats(string(data))
	return nil
}
