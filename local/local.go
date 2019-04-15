package local

import (
	"context"
	"os"
	"path/filepath"

	"github.com/moisespsena-go/assetfs/assetfsapi"
)

type contextKey uint8

const (
	CtxNames contextKey = iota
	CtxSources
)

func SetNames(ctx context.Context, names ...string) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, CtxNames, names)
}

func UnshiftNames(ctx context.Context, names ...string) context.Context {
	return SetNames(ctx, append(names, GetNames(ctx)...)...)
}

func GetNames(ctx context.Context) (names []string) {
	if ctx != nil {
		if v := ctx.Value(CtxNames); v != nil {
			return v.([]string)
		}
	}
	return
}

func SetSources(ctx context.Context, sources ...assetfsapi.LocalSource) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, CtxSources, sources)
}

func UnshiftSources(ctx context.Context, sources ...assetfsapi.LocalSource) context.Context {
	return SetSources(ctx, append(sources, GetSources(ctx)...)...)
}

func GetSources(ctx context.Context) (sources []assetfsapi.LocalSource) {
	if ctx != nil {
		if v := ctx.Value(CtxSources); v != nil {
			return v.([]assetfsapi.LocalSource)
		}
	}
	return
}

func AllSources(register assetfsapi.LocalSourceRegister, ctx context.Context) (sources []assetfsapi.LocalSource) {
	if ctx != nil {
		sources = GetSources(ctx)
		if register != nil {
			for _, name := range GetNames(ctx) {
				if src := register.Get(name); src != nil {
					sources = append(sources, src)
				}
			}
		}
	}
	return
}

type SourceDirInfo struct {
	os.FileInfo
	LocalPath string
}

func (info SourceDirInfo) Path() string {
	return info.LocalPath
}

type SourceDir struct {
	Path string
}

func (d *SourceDir) Dir() string {
	return d.Path
}

func NewSourceDir(dir string) *SourceDir {
	return &SourceDir{Path: dir}
}

func (d *SourceDir) Get(name string) (info assetfsapi.LocalSourceInfo, err error) {
	pth := filepath.Join(d.Path, FilePath(name))
	if info, err := os.Stat(pth); err == nil {
		return &SourceDirInfo{info, pth}, nil
	}
	return nil, os.ErrNotExist
}
