package local

import (
	"context"
	"os"
	"path/filepath"
)

type contextKey uint8

const (
	CtxLocalNames contextKey = iota
)

func SetLocalNames(ctx context.Context, names ...string) context.Context {
	return context.WithValue(ctx, CtxLocalNames, names)
}

func GetLocalNames(ctx context.Context) (names []string) {
	if ctx == nil {
		return
	}
	if v := ctx.Value(CtxLocalNames); v != nil {
		return v.([]string)
	}
	return
}

type SourceDir struct {
	Dir string
}

func NewSourceDir(dir string) *SourceDir {
	return &SourceDir{Dir: dir}
}

func (assets *SourceDir) Get(name string) (pth string, info os.FileInfo, ok bool) {
	pth = filepath.Join(assets.Dir, FilePath(name))
	if info, err := os.Stat(pth); err == nil {
		return pth, info, true
	}
	return "", nil, false
}

func (assets *SourceDir) GetDir(pth string) (dirPath string, ok bool) {
	dirPath = filepath.Join(assets.Dir, FilePath(pth))
	if _, err := os.Stat(dirPath); err == nil {
		return dirPath, true
	}
	return "", false
}
