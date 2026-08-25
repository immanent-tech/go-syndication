// Copyright 2026 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: 	AGPL-3.0-or-later

package georss

import (
	"encoding/xml"
	"fmt"
	"strconv"
	"strings"

	"github.com/immanent-tech/go-syndication/validation"
)

const (
	georssNS = "http://www.georss.org/georss"
	gmlNS    = "http://www.opengis.net/gml"
)

// Shared coordinate parsing. The spec explicitly permits commas as an alternative separator ("parsers should just treat
// commas as whitespace"), so tokenizing splits on either.

func splitCoordTokens(s string) []string {
	return strings.FieldsFunc(s, func(r rune) bool {
		return r == ',' || r == ' ' || r == '\t' || r == '\n' || r == '\r'
	})
}

func parseFloats(s string) ([]float64, error) {
	tokens := splitCoordTokens(s)
	out := make([]float64, len(tokens))
	for i, t := range tokens {
		v, err := strconv.ParseFloat(t, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid coordinate value %q: %w", t, err)
		}
		out[i] = v
	}
	return out, nil
}

func formatFloats(vs []float64) string {
	parts := make([]string, len(vs))
	for i, v := range vs {
		parts[i] = strconv.FormatFloat(v, 'f', -1, 64)
	}
	return strings.Join(parts, " ")
}

func coordsFromFloats(vs []float64) []Coord {
	out := make([]Coord, len(vs)/2)
	for i := range out {
		out[i] = Coord{Lat: vs[2*i], Lon: vs[2*i+1]}
	}
	return out
}

func floatsFromCoords(cs []Coord) []float64 {
	out := make([]float64, 0, len(cs)*2)
	for _, c := range cs {
		out = append(out, c.Lat, c.Lon)
	}
	return out
}

// checkAntimeridianRule enforces the spec's rule limiting how far apart consecutive points in a line/polygon may be in
// longitude (informally called the "179th meridian rule"), which exists to keep segments from being ambiguous about
// which way around the globe they go.
func checkAntimeridianRule(cs []Coord) error {
	for i := 1; i < len(cs); i++ {
		if d := cs[i].Lon - cs[i-1].Lon; d > 179 || d < -179 {
			return fmt.Errorf(
				"segment %d->%d spans more than 179 degrees of longitude (%v); this is disallowed to avoid ambiguous wrap-around",
				i-1,
				i,
				d,
			)
		}
	}
	return nil
}

func (p Point) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	return e.EncodeElement(formatFloats([]float64{p.Lat, p.Lon}), start)
}

func (p *Point) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var v struct {
		Value string `xml:",chardata"`
	}
	if err := d.DecodeElement(&v, &start); err != nil {
		return err
	}
	vs, err := parseFloats(v.Value)
	if err != nil {
		return fmt.Errorf("georss:point: %w", err)
	}
	if len(vs) != 2 {
		return fmt.Errorf("georss:point: expected exactly 2 values (lat lon), got %d", len(vs))
	}
	p.Lat, p.Lon = vs[0], vs[1]
	return nil
}

func (l Line) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	return e.EncodeElement(formatFloats(floatsFromCoords(l)), start)
}

func (l *Line) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var v struct {
		Value string `xml:",chardata"`
	}
	if err := d.DecodeElement(&v, &start); err != nil {
		return err
	}
	vs, err := parseFloats(v.Value)
	if err != nil {
		return fmt.Errorf("georss:line: %w", err)
	}
	if len(vs)%2 != 0 {
		return fmt.Errorf("georss:line: odd number of coordinate values (%d); must be lat/lon pairs", len(vs))
	}
	*l = coordsFromFloats(vs)
	return nil
}

func (l Line) Validate() error {
	if err := validation.ValidateField(l, "gte=2,unique"); err != nil {
		return fmt.Errorf("georss:line: %w", err)
	}
	if err := checkAntimeridianRule(l); err != nil {
		return fmt.Errorf("georss:line: %w", err)
	}
	return nil
}

func (p Polygon) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	return e.EncodeElement(formatFloats(floatsFromCoords(p)), start)
}

func (p *Polygon) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var v struct {
		Value string `xml:",chardata"`
	}
	if err := d.DecodeElement(&v, &start); err != nil {
		return err
	}
	vs, err := parseFloats(v.Value)
	if err != nil {
		return fmt.Errorf("georss:polygon: %w", err)
	}
	if len(vs)%2 != 0 {
		return fmt.Errorf("georss:polygon: odd number of coordinate values (%d); must be lat/lon pairs", len(vs))
	}
	*p = coordsFromFloats(vs)
	return nil
}

func (p Polygon) Validate() error {
	if err := validation.ValidateField(p, "gte=4"); err != nil {
		return fmt.Errorf("georss:polygon: %w", err)
	}
	if p[0] != p[len(p)-1] {
		return fmt.Errorf(
			"georss:polygon: first and last points must be identical to close the ring (got %v and %v)",
			p[0],
			p[len(p)-1],
		)
	}
	if err := checkAntimeridianRule(p); err != nil {
		return fmt.Errorf("georss:polygon: %w", err)
	}
	return nil
}

func (b Box) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	return e.EncodeElement(formatFloats([]float64{b.Lower.Lat, b.Lower.Lon, b.Upper.Lat, b.Upper.Lon}), start)
}

func (b *Box) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var v struct {
		Value string `xml:",chardata"`
	}
	if err := d.DecodeElement(&v, &start); err != nil {
		return err
	}
	vs, err := parseFloats(v.Value)
	if err != nil {
		return fmt.Errorf("georss:box: %w", err)
	}
	if len(vs) != 4 {
		return fmt.Errorf(
			"georss:box: expected exactly 4 values (lowerLat lowerLon upperLat upperLon), got %d",
			len(vs),
		)
	}
	b.Lower = Coord{Lat: vs[0], Lon: vs[1]}
	b.Upper = Coord{Lat: vs[2], Lon: vs[3]}
	return nil
}

func (c Circle) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	return e.EncodeElement(formatFloats([]float64{c.Center.Lat, c.Center.Lon, c.Radius}), start)
}

func (c *Circle) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var v struct {
		Value string `xml:",chardata"`
	}
	if err := d.DecodeElement(&v, &start); err != nil {
		return err
	}
	vs, err := parseFloats(v.Value)
	if err != nil {
		return fmt.Errorf("georss:circle: %w", err)
	}
	if len(vs) != 3 {
		return fmt.Errorf("georss:circle: expected exactly 3 values (lat lon radius), got %d", len(vs))
	}
	c.Center = Coord{Lat: vs[0], Lon: vs[1]}
	c.Radius = vs[2]
	return nil
}

func (w Where) Validate() error {
	set := 0
	for _, present := range []bool{w.Point != nil, w.LineString != nil, w.Polygon != nil, w.Envelope != nil} {
		if present {
			set++
		}
	}
	if set != 1 {
		return fmt.Errorf("georss:where: must contain exactly one geometry, found %d", set)
	}
	return nil
}

// Validate checks each populated geometry's own rules, and flags (as a
// soft, common-sense rule the spec doesn't state outright) more than one
// geometry being set at once -- almost certainly a mistake, since an
// item/entry conceptually has one location.
func (g GeoRSSSimple) Validate() error {
	geometries := 0
	if g.Point != nil {
		geometries++
	}
	if g.Line != nil {
		if err := g.Line.Validate(); err != nil {
			return err
		}
		geometries++
	}
	if g.Polygon != nil {
		if err := g.Polygon.Validate(); err != nil {
			return err
		}
		geometries++
	}
	if g.Box != nil {
		geometries++
	}
	if g.Circle != nil {
		geometries++
	}
	if g.Where != nil {
		if err := g.Where.Validate(); err != nil {
			return err
		}
		geometries++
	}
	if geometries > 1 {
		return fmt.Errorf("georss: %d geometry elements set simultaneously; expected at most one", geometries)
	}
	return nil
}
