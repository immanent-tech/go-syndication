// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: 	MIT

package types

import (
	"errors"
	"slices"
	"time"
)

// ErrInvalidDateTimeFormat indicates that the value of the datetime is not one of the defined DateTimeFormats. In most
// cases, this indicates the feed not using a valid datetime format according to its specification.
var ErrInvalidDateTimeFormat = errors.New("invalid datetime format")

// UnixEpoch is the time.Time value of Unix epoch.
var UnixEpoch = time.Unix(0, 0)

// GetMedianInterval calculates the median of the given set of time.Duration values.
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
