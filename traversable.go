package assetfs

import (
	"github.com/moisespsena-go/assetfs/assetfsapi"
)

type Traversable struct {
	FS           Interface
	WalkFunc     assetfsapi.WalkFunc
	WalkInfoFunc assetfsapi.WalkInfoFunc
	ReadDirFunc  func(dir string, cb assetfsapi.CbWalkInfoFunc, skipDir bool) error
	GlobFunc     assetfsapi.GlobFunc
	GlobInfoFunc assetfsapi.GlobInfoFunc
}

func (t *Traversable) Walk(dir string, cb assetfsapi.CbWalkFunc, mode ...assetfsapi.WalkMode) error {
	m := assetfsapi.WalkAll
	if len(mode) > 0 {
		m = mode[0]
	}
	return t.WalkFunc(dir, cb, m)
}

func (t *Traversable) WalkInfo(dir string, cb assetfsapi.CbWalkInfoFunc, mode ...assetfsapi.WalkMode) error {
	m := assetfsapi.WalkAll
	if len(mode) > 0 {
		m = mode[0]
	}
	return t.WalkInfoFunc(dir, cb, m)
}

func (t *Traversable) ReadDir(dir string, cb assetfsapi.CbWalkInfoFunc, skipDir bool) (err error) {
	return t.ReadDirFunc(dir, cb, skipDir)
}

func (f *Traversable) Glob(pattern assetfsapi.GlobPattern, cb func(pth string, isDir bool) error) error {
	if pth := f.FS.GetPath(); pth != "" {
		l := len(pth)
		oldFormatter := pattern.GetPathFormatter()
		pattern = pattern.PathFormatter(func(pth *string) {
			*pth = (*pth)[l+1:]
			oldFormatter(pth)
		})
	}
	return f.GlobFunc(pattern, cb)
}

func (f *Traversable) GlobInfo(pattern assetfsapi.GlobPattern, cb func(info assetfsapi.FileInfo) error) error {
	if pth := f.FS.GetPath(); pth != "" {
		l := len(pth)
		oldFormatter := pattern.GetPathFormatter()
		pattern = pattern.PathFormatter(func(pth *string) {
			*pth = (*pth)[l+1:]
			oldFormatter(pth)
		})
	}
	return f.GlobInfoFunc(pattern, cb)
}

func (f *Traversable) NewGlob(pattern assetfsapi.GlobPattern) assetfsapi.Glob {
	return NewGlob(f.FS, pattern)
}

func (f *Traversable) NewGlobString(pattern string) assetfsapi.Glob {
	return NewGlob(f.FS, NewGlobPattern(pattern))
}
