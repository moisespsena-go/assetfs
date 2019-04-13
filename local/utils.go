package local

import (
	"path"
	"path/filepath"
)

func FilePath(pth ...string) string {
	return filepath.FromSlash(path.Join(pth...))
}
