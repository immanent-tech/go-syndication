// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: 	MIT

package sanitization

import (
	"bytes"
	"html"
	"slices"
	"strings"

	"github.com/microcosm-cc/bluemonday"
)

// Option is a functional option applied to a sanitisation method.
type Option func(*config)

// WithPolicy will set a custom policy for a sanitisation method.
func WithPolicy(policy *bluemonday.Policy) Option {
	return func(s *config) {
		s.policy = policy
	}
}

// config holds configuration for sanitisation methods.
type config struct {
	policy *bluemonday.Policy
}

// SanitizeString attempts to "sanitize" a string value from a Feed/Item object. It will strip any leading/trailing
// whitespace and then run the string through bluemonday to remove dangerous components. This should retain HTML5
// content.
func SanitizeString(str string, options ...Option) string {
	cfg := &config{
		policy: bluemonday.UGCPolicy(),
	}
	for option := range slices.Values(options) {
		option(cfg)
	}
	return strings.TrimSpace(html.UnescapeString(cfg.policy.Sanitize(str)))
}

// SanitizeBytes attempts to "sanitize" a []byte value from a Feed/Item object. It will strip any leading/trailing
// whitespace and then run the string through bluemonday to remove dangerous components.
func SanitizeBytes(data []byte, options ...Option) []byte {
	cfg := &config{
		policy: bluemonday.UGCPolicy(),
	}
	for option := range slices.Values(options) {
		option(cfg)
	}
	return cfg.policy.SanitizeBytes(bytes.TrimSpace(data))
}
