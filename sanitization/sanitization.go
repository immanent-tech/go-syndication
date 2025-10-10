// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: 	AGPL-3.0-or-later

package sanitization

import (
	"bytes"
	"html"
	"strings"

	"github.com/microcosm-cc/bluemonday"
)

var safePrinter = bluemonday.UGCPolicy()

// SanitizeString attempts to "sanitize" a string value from a Feed/Item object. It will strip any leading/trailing
// whitespace and then run the string through bluemonday to remove dangerous components. This should retain HTML5
// content.
func SanitizeString(str string) string {
	return strings.TrimSpace(html.UnescapeString(safePrinter.Sanitize(str)))
}

// SanitizeBytes attempts to "sanitize" a []byte value from a Feed/Item object. It will strip any leading/trailing
// whitespace and then run the string through bluemonday to remove dangerous components.
func SanitizeBytes(b []byte) []byte {
	return safePrinter.SanitizeBytes(bytes.TrimSpace(b))
}
