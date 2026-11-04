// Copyright 2026 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: 	AGPL-3.0-or-later

package basicgeo_test

import (
	"encoding/xml"
	"strings"
	"testing"

	feeds "github.com/immanent-tech/go-syndication"
	"github.com/immanent-tech/go-syndication/extensions/basicgeo"
	"github.com/immanent-tech/go-syndication/rss"
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
	XMLName  xml.Name       `xml:"http://www.w3.org/1999/02/22-rdf-syntax-ns# RDF"`
	XMLNSRDF string         `xml:"xmlns:rdf,attr"`
	XMLNSGeo string         `xml:"xmlns:geo,attr"`
	Point    basicgeo.Point `xml:"http://www.w3.org/2003/01/geo/wgs84_pos# Point"`
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
			assert.InEpsilon(t, 12.552, doc.Point.Lon, 0.001)
		})
	}
}

func TestBasicGeo(t *testing.T) {
	exampleFeed := rss.RSS{
		Version: "2.0",
		Channel: rss.Channel{
			Title: "Geotagged items",
			Link:  "http://example.org",
			Items: []rss.Item{
				{
					Title: "A photo taken in Copenhagen",
					Link:  "http://example.org/photo/1",
					Lat:   new(55.701),
					Lon:   new(12.552),
				},
			},
		},
	}
	exampleDoc := `
 <?xml version="1.0"?>
 <?xml-stylesheet href="/eqcenter/catalogs/rssxsl.php?feed=eqs7day-M5.xml" type="text/xsl"
                  media="screen"?>
 <rss version="2.0"
      xmlns:geo="http://www.w3.org/2003/01/geo/wgs84_pos#"
      xmlns:dc="http://purl.org/dc/elements/1.1/">
  <channel>
     <title>USGS M5+ Earthquakes</title>
     <description>Real-time, worldwide earthquake list for the past 7 days</description>
     <link>https://earthquake.usgs.gov/eqcenter/</link>
     <dc:publisher>U.S. Geological Survey</dc:publisher>
     <pubDate>Thu, 27 Dec 2007 23:56:15 PST</pubDate>
     <item>
       <pubDate>Fri, 28 Dec 2007 05:24:17 GMT</pubDate>
       <title>M 5.3, northern Sumatra, Indonesia</title>
       <description>December 28, 2007 05:24:17 GMT</description>
       <link>https://earthquake.usgs.gov/eqcenter/recenteqsww/Quakes/us2007llai.php</link>
       <geo:lat>5.5319</geo:lat>
       <geo:long>95.8972</geo:long>
     </item>
   </channel>
 </rss>
	`

	t.Run("example feed", func(t *testing.T) {
		_, err := feeds.Encode(exampleFeed)
		require.NoError(t, err, "no errors in example feed")
	})

	t.Run("example doc", func(t *testing.T) {
		doc, err := feeds.Decode[*rss.RSS]("", strings.NewReader(exampleDoc))
		require := require.New(t)
		require.NoError(err, "no errors in example feed")
		require.InEpsilon(5.5319, *doc.Channel.Items[0].Lat, 0.0001, "lat matches")
		require.InEpsilon(95.8972, *doc.Channel.Items[0].Lon, 0.0001, "long matches")
	})
}
