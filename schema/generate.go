// Copyright 2024 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: 	AGPL-3.0-or-later

// Package schema contains the OpenAPI schema definitions for go-syndication.
package schema

//go:generate go tool oapi-codegen -config types-cfg.yaml types.yaml
//go:generate go tool oapi-codegen -config types-attributes-cfg.yaml types-attributes.yaml
//go:generate go tool oapi-codegen -config types-elements-cfg.yaml types-elements.yaml
//go:generate go tool oapi-codegen -config atom-cfg.yaml atom.yaml
//go:generate go tool oapi-codegen -config dc-cfg.yaml dc.yaml
//go:generate go tool oapi-codegen -config media-rss-cfg.yaml media-rss.yaml
//go:generate go tool oapi-codegen -config itunes-cfg.yaml itunes.yaml
//go:generate go tool oapi-codegen -config googleplay-cfg.yaml googleplay.yaml
//go:generate go tool oapi-codegen -config rss-ext-cfg.yaml rss-ext.yaml
//go:generate go tool oapi-codegen -config rss.cfg.yaml rss.yaml
//go:generate go tool oapi-codegen -config jsonfeed-cfg.yaml jsonfeed.yaml
//go:generate go tool oapi-codegen -config opml-cfg.yaml opml.yaml
