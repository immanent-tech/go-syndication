// Copyright 2026 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: 	MIT

package basicgeo

import (
	"encoding/xml"
	"fmt"
	"strconv"
	"strings"
)

const geoNS = "http://www.w3.org/2003/01/geo/wgs84_pos#"

func (ll LatLong) MarshalXML(enc *xml.Encoder, start xml.StartElement) error {
	if err := enc.EncodeToken(start); err != nil {
		return fmt.Errorf("marshal geo:lat_long: %w", err)
	}
	value := strconv.FormatFloat(ll.Lat, 'f', -1, 64) + "," + strconv.FormatFloat(ll.Lon, 'f', -1, 64)
	if err := enc.EncodeToken(xml.CharData(value)); err != nil {
		return fmt.Errorf("marshal geo:lat_long: %w", err)
	}
	if err := enc.EncodeToken(start.End()); err != nil {
		return fmt.Errorf("marshal geo:lat_long: %w", err)
	}
	return nil
}

func (ll *LatLong) UnmarshalXML(dec *xml.Decoder, start xml.StartElement) error {
	var v struct {
		Value string `xml:",chardata"`
	}
	if err := dec.DecodeElement(&v, &start); err != nil {
		return fmt.Errorf("unmarshal geo:lat_long: %w", err)
	}
	parts := strings.Split(strings.TrimSpace(v.Value), ",")
	if len(parts) != 2 {
		return fmt.Errorf("geo:lat_long: expected \"lat,long\", got %q", v.Value)
	}
	lat, err := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
	if err != nil {
		return fmt.Errorf("geo:lat_long: invalid latitude %q: %w", parts[0], err)
	}
	long, err := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
	if err != nil {
		return fmt.Errorf("geo:lat_long: invalid longitude %q: %w", parts[1], err)
	}
	ll.Lat, ll.Lon = lat, long
	return nil
}

// ToPosition / FromLatLong convert between the two equivalent encodings, since a document might legitimately use
// either.
func (ll LatLong) ToPosition() Position {
	return Position{Lat: ll.Lat, Lon: ll.Lon}
}

func FromPosition(p Position) (LatLong, bool) {
	return LatLong{Lat: p.Lat, Lon: p.Lon}, true
}
