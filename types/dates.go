// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: 	MIT

package types

import (
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"
)

// ErrInvalidDateTimeFormat indicates that the value of the datetime is not one of the defined DateTimeFormats. In most
// cases, this indicates the feed not using a valid datetime format according to its specification.
var ErrInvalidDateTimeFormat = errors.New("invalid datetime format")

// DateTimeFormats are the valid datetime formats across different feed specifications. A DateTime object will try to
// parse a given value as one of these formats.
var DateTimeFormats = []string{
	time.RFC1123Z,
	time.RFC1123,
	"Mon, 2 Jan 2006 15:04:05 -0700", // RFC1123Z without leading zero on day
	"Mon, 2 Jan 2006 15:04:05 MST",   // RFC1123 without leading zero on day
	"Mon, 02 Jan 2006 15:04 -0700",   // RFC1123Z without seconds
	"Mon, 02 Jan 2006 15:04 MST",     // RFC1123 without seconds
	"Mon, 2 Jan 2006 15:04 -0700",    // RFC1123Z without leading zero on day or seconds
	"Mon, 2 Jan 2006 15:04 MST",      // RFC1123 without leading zero on day or seconds
	time.RFC3339,
	"2006-01-02T15:04:05.00Z07:00", // DC Date with fractional seconds.
	time.DateTime,
	time.DateOnly,                       // DC Date only.
	time.DateOnly + "T" + "15:04-07:00", // DC no seconds.
	"2006-01",                           // DC year and month only.
	"2006",                              // DC year only.
	"Jan 2, 2006",                       // Invalid but rarely used...
	"2006 Jan 2 15:04:05 MST",           // Invalid but rarely used...
	"Mon, 2 Jan 2006",                   // Invalid but used...
}

// UnixEpoch is the time.Time value of Unix epoch.
var UnixEpoch = time.Unix(0, 0)

// DateTime is a datetime value for a feed (or item) object, such as its published/updated date.
type DateTime struct {
	time.Time
}

// Valid will determine whether the value of the DateTime is valid. That is, is not the zero value or equal to the Unix
// Epoch.
func (d *DateTime) Valid() (bool, error) {
	switch {
	case d.IsZero():
		return false, fmt.Errorf("%w: is zero time value", ErrInvalidDateTimeFormat)
	case d.Equal(UnixEpoch):
		return false, fmt.Errorf("%w: is unix epoch", ErrInvalidDateTimeFormat)
	default:
		return true, nil
	}
}

// MarshalJSON handles marshaling a DateTime to JSON.
func (d *DateTime) MarshalJSON() ([]byte, error) {
	date, err := json.Marshal(d.Format(DateTimeFormats[0]))
	if err != nil {
		return nil, errors.Join(ErrInvalidDateTimeFormat, err)
	}
	return date, nil
}

// UnmarshalJSON handles unmarshaling a DateTime from JSON.
func (d *DateTime) UnmarshalJSON(data []byte) error {
	var dateStr string
	if err := json.Unmarshal(data, &dateStr); err != nil {
		return errors.Join(ErrInvalidDateTimeFormat, err)
	}
	parsed, err := tryFormats(dateStr)
	if err != nil {
		return err
	}
	d.Time = parsed
	return nil
}

// String returns a string representation of the DateTime.
func (d *DateTime) String() string {
	return d.Format(DateTimeFormats[0])
}

// UnmarshalText will unmarshal/parse a DateTime from the given string.
func (d *DateTime) UnmarshalText(data []byte) error {
	parsed, err := tryFormats(string(data))
	if err != nil {
		return err
	}
	d.Time = parsed
	return nil
}

func tryFormats(data string) (time.Time, error) {
	var parsed time.Time
	for format := range slices.Values(DateTimeFormats) {
		data = strings.TrimSpace(data)
		value, err := time.Parse(format, data)
		if err != nil {
			continue
		}
		parsed = value
	}
	return parsed, nil
}

func GetMedianInterval(data []time.Duration) time.Duration {
	dataCopy := make([]time.Duration, len(data))
	copy(dataCopy, data)

	slices.Sort(dataCopy)

	var median time.Duration
	if l := len(dataCopy); l == 0 {
		return 0
	} else if l%2 == 0 {
		median = (dataCopy[l/2-1] + dataCopy[l/2]) / 2
	} else {
		median = dataCopy[l/2]
	}

	return median
}
