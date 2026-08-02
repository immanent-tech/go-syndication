// Copyright 2026 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: 	AGPL-3.0-or-later

package source

import (
	"encoding/xml"
	"fmt"
	"strings"
	"time"
)

func (m Markdown) MarshalXML(enc *xml.Encoder, start xml.StartElement) error {
	if m.AsCDATA {
		if err := enc.EncodeElement(struct {
			Value string `xml:",cdata"`
		}{m.Value}, start); err != nil {
			return fmt.Errorf("marshal markdown: %w", err)
		}
	}
	if err := enc.EncodeElement(struct {
		Value string `xml:",chardata"`
	}{m.Value}, start); err != nil {
		return fmt.Errorf("marshal markdown: %w", err)
	}
	return nil
}

func (m *Markdown) UnmarshalXML(dec *xml.Decoder, start xml.StartElement) error {
	var v struct {
		Value string `xml:",chardata"`
	}
	if err := dec.DecodeElement(&v, &start); err != nil {
		return fmt.Errorf("unmarshal markdown: %w", err)
	}
	m.Value = v.Value
	return nil
}

const archiveDayLayout = "2006-01-02"

func (d ArchiveDay) MarshalXML(enc *xml.Encoder, start xml.StartElement) error {
	if d.Time.IsZero() {
		return fmt.Errorf("source:archive: zero date for <%s>", start.Name.Local)
	}
	if err := enc.EncodeToken(start); err != nil {
		return fmt.Errorf("marshal archiveday: %w", err)
	}
	if err := enc.EncodeToken(xml.CharData(d.Time.Format(archiveDayLayout))); err != nil {
		return fmt.Errorf("marshal archiveday: %w", err)
	}
	if err := enc.EncodeToken(start.End()); err != nil {
		return fmt.Errorf("marshal archiveday: %w", err)
	}
	return nil
}

func (d *ArchiveDay) UnmarshalXML(dec *xml.Decoder, start xml.StartElement) error {
	var v struct {
		Value string `xml:",chardata"`
	}
	if err := dec.DecodeElement(&v, &start); err != nil {
		return fmt.Errorf("unmarshal archiveday: %w", err)
	}
	t, err := time.Parse(archiveDayLayout, strings.TrimSpace(v.Value))
	if err != nil {
		return fmt.Errorf("<%s>: invalid yyyy-mm-dd date %q: %w", start.Name.Local, v.Value, err)
	}
	d.Time = t
	return nil
}

// EffectiveEndDay implements "If <source:endDay> [is absent, it] defaults to the pubDate of the feed, if it's
// specified. If not, it defaults to the current date."
func (a Archive) EffectiveEndDay(channelPubDate *time.Time) time.Time {
	if !a.EndDay.Time.IsZero() {
		return a.EndDay.Time
	}
	if channelPubDate != nil {
		return *channelPubDate
	}
	return time.Now()
}

func (r InReplyTo) EffectiveIsPermaLink() bool {
	return r.IsPermalink == nil || *r.IsPermalink
}
