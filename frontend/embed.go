package frontend

import (
	"embed"
	"io/fs"
)

//go:embed index.html
var files embed.FS

func FS() fs.FS {
	return files
}
