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
