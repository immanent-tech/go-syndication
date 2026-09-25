// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: 	MIT

package types

import (
	"errors"
	"fmt"
	"slices"
	"strings"
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

// NamedZoneOffsetsRFC822 maps RFC 822 zone abbreviations to their UTC offset in seconds. Go's time.Parse does NOT reliably
// resolve these itself. An unrecognized "MST"-style abbreviation is silently assigned a zero offset by the standard
// library, which would misparse e.g. "EST" as UTC. We look these up ourselves instead of trusting the stdlib's
// zone-name parsing.
var NamedZoneOffsetsRFC822 = map[string]int{
	"UT": 0, "GMT": 0, "Z": 0,
	"EST": -5 * 3600, "EDT": -4 * 3600,
	"CST": -6 * 3600, "CDT": -5 * 3600,
	"MST": -7 * 3600, "MDT": -6 * 3600,
	"PST": -8 * 3600, "PDT": -7 * 3600,
}

// DateOnlyLayoutsRFC822 are candidate layouts for parsing timestamps. Note that not all of these do conform to the spec, they
// are gathered from real-world usage in feeds. Where a timezone is included, it is ended in a literal "-0700"
// placeholder for a *numeric* offset. We normalize any named zone abbreviation in the input to a numeric offset before
// trying these, so a single set of layouts covers both cases.
var DateOnlyLayoutsRFC822 = []string{
	"Mon, 02 Jan 2006 15:04:05 -0700",
	"Mon, 02 Jan 06 15:04:05 -0700",
	"Mon, 02 Jan 2006 15:04:05",
	"Mon, 02 Jan 2006",
	"Mon, 2 Jan 2006",
	"Mon, 2 Jan 2006 15:04:05 -0700",
	"Mon, 2 Jan 06 15:04:05 -0700",
	"02 Jan 2006 15:04:05 -0700",
	"02 Jan 06 15:04:05 -0700",
	"Mon, 02 Jan 2006 15:04 -0700",
	"Mon, 02 Jan 06 15:04 -0700",
	"Mon, 2 Jan 2006 15:04 -0700",
	"Mon, 2 Jan 06 15:04 -0700",
	"02 Jan 2006 15:04 -0700",
	"02 Jan 06 15:04 -0700",
	"2006 Jan 02 15:04:05 -0700",
	"2006 Jan 02 15:04:05 MST",
	"Jan 02, 2006",
	"2006-01-02T15:04:05+00:00",
	"2006-01-02T15:04:05+0000",
	"2006-01-02 15:04 MST",
	time.RFC3339,
	time.DateOnly,
	time.DateTime,
}

// ParseRFC822 parses an RSS date-time value leniently: it accepts both numeric zone offsets (+0100, -0600) and the
// named zone abbreviations registered in namedZoneOffsets, with or without a weekday, with a 2- or 4-digit year, and
// with or without seconds.
func ParseRFC822(ts string) (time.Time, error) {
	ts = strings.TrimSpace(ts)

	fields := strings.Fields(ts)
	if len(fields) == 0 {
		return time.Time{}, errors.New("rss date-time: empty value")
	}
	lastIdx := len(fields) - 1
	zone := fields[lastIdx]

	// If the trailing token is a known named zone, rewrite it as a
	// numeric offset so a single family of layouts handles everything.
	if off, ok := NamedZoneOffsetsRFC822[strings.ToUpper(zone)]; ok {
		sign := "+"
		if off < 0 {
			sign = "-"
			off = -off
		}
		fields[lastIdx] = fmt.Sprintf("%s%02d%02d", sign, off/3600, (off%3600)/60)
		ts = strings.Join(fields, " ")
	}
	// Otherwise, if it's already a numeric offset (+0100 / -0600) or a literal "Z", leave it as-is; the loop below will
	// try it against each layout and Go's -0700 verb correctly parses "+HHMM"/"-HHMM".
	var lastErr error
	for layout := range slices.Values(DateOnlyLayoutsRFC822) {
		if t, err := time.Parse(layout, ts); err == nil {
			return t, nil
		} else {
			lastErr = err
		}
	}
	return time.Time{}, fmt.Errorf("rss date-time: could not parse %q: %w", ts, lastErr)
}
