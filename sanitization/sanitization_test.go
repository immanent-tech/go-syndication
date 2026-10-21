// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: 	MIT

package sanitization_test

import (
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/immanent-tech/go-syndication/sanitization"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

func TestSanitizeURLString(t *testing.T) {
	cases := []struct {
		in   string
		want bool
	}{
		{"https://example.com/a", true},
		{"http://example.com", true},
		{"HTTPS://EXAMPLE.COM/", true},
		{"javascript:alert(1)", false},
		{"JaVaScRiPt:alert(1)", false},
		{" javascript:alert(1)", false},
		{"java\tscript:alert(1)", false},
		{"java\nscript:alert(1)", false},
		{"\x01javascript:alert(1)", false},
		{"data:text/html,<script>alert(1)</script>", false},
		{"vbscript:msgbox(1)", false},
		{"file:///etc/passwd", false},
		{"gopher://127.0.0.1:6379/_INFO", false},
		{"ftp://example.com/x", true},
		{"mailto:a@example.com", true},
		{"//evil.example/x", false},
		{"/relative", false},
		{"example.com/page", false},
		{"https:///nohost", false},
		{"https://exa mple.com/", false},
		{"https://user:pass@example.com/", false},
		{"https://example.com@evil.example/", false},
		{"", false},
	}
	for _, c := range cases {
		if got, err := sanitization.SanitizeURLString(c.in); (err != nil) == c.want {
			t.Errorf("SanitizeURLString(%q) = %v, want %v", c.in, got, c.want)
		}
	}
}

func TestSanitizeHTMLKeepsGoodContent(t *testing.T) {
	in := `<p>Hello <strong>world</strong> <a href="https://example.com/ok">link</a> <img src="https://example.com/i.png" alt="pic"></p><blockquote>quote</blockquote>`
	out := sanitization.SanitizeHTML(in)
	for _, want := range []string{"<strong>world</strong>", `href="https://example.com/ok"`, `src="https://example.com/i.png"`, `alt="pic"`, "<blockquote>quote</blockquote>"} {
		if !strings.Contains(out, want) {
			t.Errorf("legitimate markup lost: want %q in %q", want, out)
		}
	}
}

func TestSanitizeText(t *testing.T) {
	cases := []struct{ name, in, want string }{
		{"bidi override", "invoice\u202Efdp.exe", "invoicefdp.exe"},
		{"zero width", "pay\u200Bpal", "paypal"},
		{"BOM", "\uFEFFtitle", "title"},
		{"control chars", "a\x00b\x1b[31mc", "ab[31mc"},
		{"newlines", "line1\nline2\r\nline3", "line1 line2 line3"},
		{"coalesces runs", "a   b\t\t c\n\n\nd", "a b c d"},
		{"trims ends", "  \n lead and trail \t ", "lead and trail"},
		{"space around stripped char", "a \u200B b", "a b"},
		{"NBSP and ideographic space", "a\u00a0\u00a0b\u3000\u3000c", "a b c"},
		{"only whitespace", " \n\t ", ""},
		{"keeps emoji ZWJ", "\U0001F468\u200D\U0001F469\u200D\U0001F467", "\U0001F468\u200D\U0001F469\u200D\U0001F467"},
		{"invalid utf8", "ok\xff\xfebad", "ok\uFFFDbad"},
	}
	for _, c := range cases {
		if got := sanitization.SanitizeText(c.in, 300); got != c.want {
			t.Errorf("%s: SanitizeText(%q) = %q, want %q", c.name, c.in, got, c.want)
		}
	}
	long := sanitization.SanitizeText(strings.Repeat("\u00e9", 10000), 300)
	if utf8.RuneCountInString(long) != 300 || !utf8.ValidString(long) {
		t.Errorf("length cap failed: %d runes", utf8.RuneCountInString(long))
	}
}

func TestResourceExhaustion(t *testing.T) {
	n := 5000
	body := strings.Repeat("<blockquote>", n) + "x" + strings.Repeat("</blockquote>", n)
	out := sanitization.SanitizeHTML(body)
	ctx := &html.Node{Type: html.ElementNode, Data: "div", DataAtom: atom.Div}
	_, err := html.ParseFragment(strings.NewReader(out), ctx)
	assert.Error(t, err, "sanitized output does not parse: %v", err)
}
