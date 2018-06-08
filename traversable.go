package assetfs

import "github.com/moisespsena/go-assetfs/api"

type Traversable struct {
	fs Interface
	WalkFunc     api.WalkFunc
	WalkInfoFunc api.WalkInfoFunc
	ReadDirFunc  func(dir string, cb api.CbWalkInfoFunc, skipDir bool) error
	GlobFunc      api.GlobFunc
	GlobInfoFunc  api.GlobInfoFunc
}

func (t *Traversable) Walk(dir string, cb api.CbWalkFunc, mode ...api.WalkMode) error {
	m := api.WalkAll
	if len(mode) > 0 {
		m = mode[0]
	}
	return t.WalkFunc(dir, cb, m)
}

func (t *Traversable) WalkInfo(dir string, cb api.CbWalkInfoFunc, mode ...api.WalkMode) error {
	m := api.WalkAll
	if len(mode) > 0 {
		m = mode[0]
	}
	return t.WalkInfoFunc(dir, cb, m)
}

func (t *Traversable) ReadDir(dir string, cb api.CbWalkInfoFunc, skipDir bool) (err error) {
	return t.ReadDirFunc(dir, cb, skipDir)
}

func (f *Traversable) Glob(pattern api.GlobPattern, cb func(pth string, isDir bool) error) error {
	if pth := f.fs.GetPath(); pth != "" {
		l := len(pth)
		oldFormatter := pattern.GetPathFormatter()
		pattern = pattern.PathFormatter(func(pth *string) {
			*pth = (*pth)[l+1:]
			oldFormatter(pth)
		})
	}
	return f.GlobFunc(pattern, cb)
}

func (f *Traversable) GlobInfo(pattern api.GlobPattern, cb func(info api.FileInfo) error) error {
	if pth := f.fs.GetPath(); pth != "" {
		l := len(pth)
		oldFormatter := pattern.GetPathFormatter()
		pattern = pattern.PathFormatter(func(pth *string) {
			*pth = (*pth)[l+1:]
			oldFormatter(pth)
		})
	}
	return f.GlobInfoFunc(pattern, cb)
}

func (f *Traversable) NewGlob(pattern api.GlobPattern) api.Glob {
	return NewGlob(f.fs, pattern)
}

func (f *Traversable) NewGlobString(pattern string) api.Glob {
	return NewGlob(f.fs, NewGlobPattern(pattern))
}