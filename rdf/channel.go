// Copyright 2026 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: 	AGPL-3.0-or-later

package rdf

import (
	"strings"
	"time"
)

func (c *Channel) GetAuthors() []string {
	if c.Creator != nil {
		return *c.Creator
	}
	return nil
}

func (c *Channel) GetContributors() []string {
	if c.Contributor != nil {
		return *c.Contributor
	}
	return nil
}

func (c *Channel) GetCategories() []string {
	if c.Subject != nil {
		return *c.Subject
	}
	return nil
}

func (c *Channel) GetDescription() string {
	return c.Description
}

func (c *Channel) GetTitle() string {
	return c.Title
}

func (c *Channel) GetLanguage() *string {
	if c.Language != nil {
		return new(strings.Join(*c.Language, " "))
	}
	return nil
}

func (c *Channel) GetLink() string {
	return c.Link
}

func (c *Channel) GetSourceURL() string {
	return c.About
}

func (c *Channel) SetSourceURL(value string) {
	c.About = value
}

func (c *Channel) GetPublishedDate() *time.Time {
	if c.Date != nil {
		v := (*c.Date)[0].Value
		return &v
	}
	return nil
}

func (c *Channel) GetRights() *string {
	if c.Rights != nil {
		return new(strings.Join(*c.Rights, " "))
	}
	return nil
}

func (c *Channel) GetUpdateInterval() time.Duration {
	if c.SYUdatePeriod != nil {
		var baseInterval time.Duration
		switch c.SYUdatePeriod.Value {
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
		if c.SYUpdateFrequency != nil && c.SYUpdateFrequency.Value > 1 {
			return time.Duration(int64(float64(baseInterval) / float64(c.SYUpdateFrequency.Value)))
		}
		return baseInterval
	}
	return 0
}
