// Copyright 2026 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: 	AGPL-3.0-or-later

//go:build mage
// +build mage

package main

import (
	"log/slog"
	"path/filepath"
	"slices"

	"github.com/goforj/godump"
)

var tests = map[string][]string{
	"atom": {"./feedvalidator/testcases/atom/must/*"},
}

func GenerateTestCases() error {
	for group, dirs := range tests {
		slog.Info("Generating tests for group.", slog.String("group", group))
		for dir := range slices.Values(dirs) {
			slog.Info("Generating tests from directory.", slog.String("directory", dir))
			files, err := filepath.Glob(dir)
			if err != nil {
				slog.Warn("Could not glob files.", slog.String("directory", dir), slog.Any("error", err))
			}
			godump.Dump(files)
		}
	}

	return nil
}
