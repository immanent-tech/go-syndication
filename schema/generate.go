/*
 * Copyright (c) Joshua Rich <joshua.rich@gmail.com>.
 * SPDX-License-Identifier: MIT
 */

// Package schema contains the OpenAPI schema definitions for go-syndication.
package schema

//go:generate go tool oapi-codegen -config types-cfg.yaml types.yaml
//go:generate go tool oapi-codegen -config atom-cfg.yaml atom.yaml
//go:generate go tool oapi-codegen -config dc-cfg.yaml dc.yaml
//go:generate go tool oapi-codegen -config basicgeo-cfg.yaml basicgeo.yaml
//go:generate go tool oapi-codegen -config georss-cfg.yaml georss.yaml
//go:generate go tool oapi-codegen -config media-rss-cfg.yaml media-rss.yaml
//go:generate go tool oapi-codegen -config itunes-cfg.yaml itunes.yaml
//go:generate go tool oapi-codegen -config googleplay-cfg.yaml googleplay.yaml
//go:generate go tool oapi-codegen -config syndication-cfg.yaml syndication.yaml
//go:generate go tool oapi-codegen -config content-cfg.yaml content.yaml
//go:generate go tool oapi-codegen -config link-cfg.yaml link.yaml
//go:generate go tool oapi-codegen -config wfw-cfg.yaml wfw.yaml
//go:generate go tool oapi-codegen -config rss.cfg.yaml rss.yaml
//go:generate go tool oapi-codegen -config jsonfeed-cfg.yaml jsonfeed.yaml
//go:generate go tool oapi-codegen -config opml-cfg.yaml opml.yaml
//go:generate go tool oapi-codegen -config extensions-cfg.yaml extensions.yaml
//go:generate go tool oapi-codegen -config rdf-cfg.yaml rdf.yaml
//go:generate go tool oapi-codegen -config source-cfg.yaml source.yaml
//go:generate go tool oapi-codegen -config slash-cfg.yaml slash.yaml
