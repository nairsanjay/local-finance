package localfinance

import (
	"embed"
	"io/fs"
)

//go:embed all:frontend/dist
var EmbeddedFrontend embed.FS

func GetStaticFS() fs.FS {
	sub, err := fs.Sub(EmbeddedFrontend, "frontend/dist")
	if err != nil {
		return nil
	}
	return sub
}
