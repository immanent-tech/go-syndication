// Copyright 2026 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: 	MIT

package extensions_test

import (
	"encoding/xml"
	"strings"
	"testing"

	"github.com/immanent-tech/go-syndication/extensions"
)

func TestNamespaceRewriter_Token(t *testing.T) {
	tests := []struct {
		name    string // description of this test case
		want    xml.Token
		dec     extensions.NamespaceRewriter
		wantErr bool
	}{
		{
			name: "slash namespace",
			dec: extensions.NamespaceRewriter{
				Dec: xml.NewDecoder(
					strings.NewReader(
						`<comments xmlns="http://purl.org/rss/1.0/modules/slash/">4</comments>`,
					),
				),
			},
			want: "slashComments",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := tt.dec.Token()
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("Token() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("Token() succeeded unexpectedly")
			}
			if gotValue, ok := got.(xml.StartElement); !ok {
				t.Fatal("Token() not decoded")
			} else {
				if gotValue.Name.Local != tt.want {
					t.Errorf("Token() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}
