// Copyright 2026 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: 	AGPL-3.0-or-later

package basicgeo

import (
	"encoding/xml"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var standalone = `
<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"
        xmlns:geo="http://www.w3.org/2003/01/geo/wgs84_pos#">
  <geo:Point>
    <geo:lat>55.701</geo:lat>
    <geo:long>12.552</geo:long>
  </geo:Point>
</rdf:RDF>
`

// A minimal standalone RDF document matching the vocabulary's own basic example:
// <rdf:RDF><geo:Point><geo:lat/><geo:long/></geo:Point></rdf:RDF>.
type RDFDoc struct {
	XMLName  xml.Name `xml:"http://www.w3.org/1999/02/22-rdf-syntax-ns# RDF"`
	XMLNSRDF string   `xml:"xmlns:rdf,attr"`
	XMLNSGeo string   `xml:"xmlns:geo,attr"`
	Point    Point    `xml:"http://www.w3.org/2003/01/geo/wgs84_pos# Point"`
}

func TestStandalone(t *testing.T) {
	tests := []struct {
		name string
		data string
	}{
		{
			name: "standalone",
			data: standalone,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var doc RDFDoc
			err := xml.Unmarshal([]byte(tt.data), &doc)
			require.NoError(t, err)
			assert.InEpsilon(t, 55.701, doc.Point.Lat, 0.001)
			assert.InEpsilon(t, 12.552, doc.Point.Long, 0.001)
		})
	}
}
