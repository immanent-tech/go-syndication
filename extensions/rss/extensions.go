// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: 	MIT

package rss

import (
	"encoding/json"
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
