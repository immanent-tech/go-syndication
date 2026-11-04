// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: 	MIT

package syndication

import (
	"bytes"
	"time"
)

func (p SYUpdatePeriod) MarshalText() ([]byte, error) {
	return []byte(p), nil
}

func (p *SYUpdatePeriod) UnmarshalText(data []byte) error {
	cleaned := bytes.TrimSpace(data)
	value := SYUpdatePeriod(string(cleaned))
	*p = value
	return nil
}

func (p SYUpdatePeriod) Interval(freq *SYUpdateFrequency) time.Duration {
	var baseInterval time.Duration
	switch p {
	case "hourly":
		baseInterval = time.Hour
	case "daily":
		baseInterval = 24 * time.Hour
	case "weekly":
		baseInterval = 7 * 24 * time.Hour
	case "yearly":
		baseInterval = 365 * 24 * time.Hour
	default:
		baseInterval = 5 * time.Hour
	}
	if freq != nil && *freq > 1 {
		return time.Duration(int64(float64(baseInterval) / float64(*freq)))
	}
	return baseInterval
}
