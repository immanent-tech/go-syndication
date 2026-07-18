// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: 	MIT

package rss

import (
	"encoding/json"
	"encoding/xml"
	"fmt"

	"github.com/immanent-tech/go-syndication/sanitization"
)

func (c *SYUpdatePeriod) UnmarshalText(data []byte) error {
	c.Value = string(sanitization.SanitizeBytes(data))
	return nil
}

func (c *SYUpdatePeriod) UnmarshalJSON(data []byte) error {
	var chardata struct {
		CharData []byte `json:"CharData"`
	}

	if err := json.Unmarshal(data, &chardata); err != nil {
		return fmt.Errorf("cannot unmarshal chardata: %w", err)
	}

	c.Value = string(sanitization.SanitizeBytes(chardata.CharData))

	return nil
}

func NewContentEncoded(value string, cdata bool) ContentEncoded {
	return ContentEncoded{
		Value: value,
		CDATA: cdata,
	}
}

func (c ContentEncoded) String() string {
	return c.Value
}

// MarshalXML implements xml.Marshaler.
func (c ContentEncoded) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	// Force the literal element name "content:encoded". Go's xml package
	// doesn't manage namespace prefixes well on marshal, so the common
	// workaround is to declare xmlns:content on the root element (see
	// RSSRoot below) and just use the literal prefixed name here.
	start.Name = xml.Name{Local: "content:encoded"}

	if c.CDATA {
		return e.EncodeElement(struct {
			Value string `xml:",cdata"`
		}{c.Value}, start)
	}
	return e.EncodeElement(struct {
		Value string `xml:",chardata"`
	}{c.Value}, start)
}

// UnmarshalXML implements xml.Unmarshaler.
//
// Note: Go's decoder does not distinguish a CDATA section from ordinary
// character data at the token level -- both come back as CharData and get
// concatenated into a plain ",chardata" field. That means this single
// implementation correctly reads content:encoded whether the source feed
// used CDATA-escaping or entity-encoding, per the spec's "entity-encoded or
// CDATA-escaped" wording. We can't reliably recover which form was
// originally used, so CDATA is left at its zero value (false) after
// decoding; set it yourself before re-marshaling if it matters.
func (c *ContentEncoded) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var v struct {
		Value string `xml:",chardata"`
	}
	if err := d.DecodeElement(&v, &start); err != nil {
		return err
	}
	c.Value = v.Value
	return nil
}

func (c ContentEncoded) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.Value)
}

func (c *ContentEncoded) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	c.Value = s
	return nil
}
