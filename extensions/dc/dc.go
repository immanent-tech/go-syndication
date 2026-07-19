// Copyright 2026 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: 	AGPL-3.0-or-later

package dc

import (
	"encoding/xml"
	"fmt"
	"regexp"
	"time"
)

// w3cdtfLayouts maps each precision to its Go time layout, in the order
// the spec defines them.
var w3cdtfLayouts = []struct {
	pattern *regexp.Regexp
	layout  string
	prec    W3CDTFPrecision
}{
	{
		regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(\.\d+)?(Z|[+-]\d{2}:\d{2})$`),
		"2006-01-02T15:04:05.999999999Z07:00",
		PrecisionSecond,
	},
	{
		regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}(Z|[+-]\d{2}:\d{2})$`),
		"2006-01-02T15:04Z07:00",
		PrecisionMinute,
	},
	{regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`), "2006-01-02", PrecisionDay},
	{regexp.MustCompile(`^\d{4}-\d{2}$`), "2006-01", PrecisionMonth},
	{regexp.MustCompile(`^\d{4}$`), "2006", PrecisionYear},
}

// MarshalXML implements xml.Marshaler, formatting at the stored precision.
//
// Note: PrecisionSecond uses a trimmed fractional layout (".999999999"), so a DCDate with non-zero sub-second precision
// will include a fraction automatically, and one with none will render as plain "...T15:04:05Z07:00" -- both are legal
// W3CDTF, this just means the distinct "with a decimal fraction" form isn't separately tracked as its own precision
// level; it falls out of the actual time.Time value instead.
func (d DCDate) MarshalXML(enc *xml.Encoder, start xml.StartElement) error {
	if d.Value.IsZero() {
		return fmt.Errorf("dcdate: zero time.Time value for <%s>", start.Name.Local)
	}
	var layout string
	switch d.Precision {
	case PrecisionYear:
		layout = "2006"
	case PrecisionMonth:
		layout = "2006-01"
	case PrecisionDay:
		layout = "2006-01-02"
	case PrecisionMinute:
		layout = "2006-01-02T15:04Z07:00"
	default: // PrecisionSecond
		layout = "2006-01-02T15:04:05.999999999Z07:00"
	}
	if err := enc.EncodeToken(start); err != nil {
		return fmt.Errorf("marshal dcdate: %w", err)
	}
	if err := enc.EncodeToken(xml.CharData(d.Value.Format(layout))); err != nil {
		return fmt.Errorf("marshal dcdate: %w", err)
	}
	if err := enc.EncodeToken(start.End()); err != nil {
		return fmt.Errorf("marshal dcdate: %w", err)
	}

	return nil
}

// UnmarshalXML implements xml.Unmarshaler, detecting which of the five legal W3CDTF forms was used and parsing (and
// remembering the precision of) it accordingly.
func (d *DCDate) UnmarshalXML(dec *xml.Decoder, start xml.StartElement) error {
	var value struct {
		Value string `xml:",chardata"`
	}
	if err := dec.DecodeElement(&value, &start); err != nil {
		return fmt.Errorf("unmarshal dcdate: %w", err)
	}
	for _, candidate := range w3cdtfLayouts {
		if !candidate.pattern.MatchString(value.Value) {
			continue
		}
		t, err := time.Parse(candidate.layout, value.Value)
		if err != nil {
			return fmt.Errorf("<%s>: invalid W3CDTF value %q: %w", start.Name.Local, value.Value, err)
		}
		d.Value = t
		d.Precision = candidate.prec
		return nil
	}
	return fmt.Errorf("<%s>: %q does not match any legal W3CDTF form", start.Name.Local, value.Value)
}
