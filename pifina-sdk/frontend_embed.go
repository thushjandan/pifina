// Copyright (c) 2023 Thushjandan Ponnudurai
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package pifina

import (
	"embed"
	"io/fs"
)

//go:embed all:frontend/build
var assets embed.FS

func Assets() (fs.FS, error) {
	return fs.Sub(assets, "frontend/build")
}
