// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: 	MIT

package types

import (
	"slices"
	"time"
)

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
