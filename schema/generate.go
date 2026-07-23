// Copyright 2024 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: 	MIT

// Package schema contains the OpenAPI schema definitions for go-syndication.
package schema

//go:generate go tool oapi-codegen -config types-cfg.yaml types.yaml
//go:generate go tool oapi-codegen -config atom-cfg.yaml atom.yaml
//go:generate go tool oapi-codegen -config dc-cfg.yaml dc.yaml
//go:generate go tool oapi-codegen -config media-rss-cfg.yaml media-rss.yaml
//go:generate go tool oapi-codegen -config itunes-cfg.yaml itunes.yaml
//go:generate go tool oapi-codegen -config googleplay-cfg.yaml googleplay.yaml
//go:generate go tool oapi-codegen -config rss-ext-cfg.yaml rss-ext.yaml
//go:generate go tool oapi-codegen -config rss.cfg.yaml rss.yaml
//go:generate go tool oapi-codegen -config jsonfeed-cfg.yaml jsonfeed.yaml
//go:generate go tool oapi-codegen -config opml-cfg.yaml opml.yaml
//go:generate go tool oapi-codegen -config extensions-cfg.yaml extensions.yaml
//go:generate go tool oapi-codegen -config rdf-cfg.yaml rdf.yaml
