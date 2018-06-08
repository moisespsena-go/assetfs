package assetfs

import (
	"os"
	"path/filepath"
	"io"
	"github.com/moisespsena/go-assetfs/api"
	"github.com/moisespsena/go-path-helpers"
	"strings"
)

// Names list matched files from assetfs
func filesystemGlob(fs *AssetFileSystem, pattern api.GlobPattern, cb func(pth string, isDir bool) error) error {
	return filesystemGlobInfo(fs, pattern, func(info api.FileInfo) error {
		return cb(info.Path(), info.IsDir())
	})
}
// Names list matched files from assetfs
func filesystemGlobInfo(fs *AssetFileSystem, pattern api.GlobPattern, cb func(info api.FileInfo) error) error {
	set := make(map[string]bool)
	cb2 := func(info api.FileInfo) error {
		if info.IsDir() {
			if !pattern.AllowDirs() {
				return nil
			}
		} else {
			if !pattern.AllowFiles() {
				return nil
			}
		}
		pth := info.Path()
		ok := pattern.Match(filepath.Base(pth))
		if !ok {
			return nil
		}
		if _, ok := set[pth]; !ok {
			if err := cb(info); err != nil {
				return err
			}
			set[pth] = true
		}
		return nil
	}
	if pattern.IsRecursive() {
		return fs.WalkInfo(pattern.Dir(), cb2, api.WalkAll ^ api.WalkDirs)
	}
	return fs.ReadDir(pattern.Dir(), cb2, true)
}

// Asset get content with name from assetfs
func filesystemAsset(fs *AssetFileSystem, name string) (api.AssetInterface, error) {
	info, err := filesystemAssetInfo(fs, name)
	if err != nil {
		return nil, err
	}
	data, err := info.Data()
	if err != nil {
		return nil, err
	}
	return NewAsset(name, data), nil
}

func filesystemAssetInfo(fs *AssetFileSystem, path string) (info api.FileInfo, err error) {
	var r string
	err = fs.PathsFrom(path, func(pth string) error {
		r = pth
		return io.EOF
	})
	if err != nil {
		return nil, err
	}
	if r == "" {
		return nil, api.NotFound(path)
	}
	stat, err := os.Stat(r)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, api.NotFound(path)
		}
		return nil, err
	}
	return &RealFileInfo{FSFileInfoBase{fs, path}, stat, r}, nil
}

func filesystemWalk(fs *AssetFileSystem, dir string, cb api.CbWalkInfoFunc, mode api.WalkMode) (err error) {
	if dir == "" {
		dir = "."
	}

	if dir == "." {
		if fs.nameSpaces != nil {
			for _, ns := range fs.nameSpaces {
				err = filesystemWalk(ns, ".", cb, mode | api.WalkNameSpacesLookUp ^ api.WalkParentLookUp)
				if err != nil {
					return err
				}
			}
		}

		if err != nil {
			return
		}

		err = fs.eachPath(mode.IsReverse(), func(root string) error {
			return filepath.Walk(root, func(realPath string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				if realPath == "." || realPath == root {
					return nil
				}
				pth := strings.TrimPrefix(realPath, root)
				if pth[0] == filepath.Separator {
					pth = pth[1:]
				}
				var inf api.FileInfo
				if info.IsDir() {
					if !mode.IsDirs() {
						return nil
					}

					inf = &RealDirFileInfo{&RealFileInfo{FSFileInfoBase{fs, pth}, info, realPath}}
				} else {
					if !mode.IsFiles() {
						return nil
					}
					inf = &RealFileInfo{FSFileInfoBase{fs, pth}, info, realPath}
				}
				return cb(inf)
			})
		})
		if err != nil {
			return
		}
	} else {
		if mode.IsNameSpacesLookUp() && fs.nameSpaces != nil {
			parts := strings.SplitN(dir, string(os.PathSeparator), 2)
			if ns, ok := fs.nameSpaces[parts[0]]; ok {
				err = filesystemWalk(ns, parts[1], cb, mode | api.WalkNameSpacesLookUp ^ api.WalkParentLookUp)
				if err != nil {
					return err
				}
			}
		}

		err = fs.eachPath(mode.IsReverse(), func(root string) (err error) {
			root = filepath.Join(root, dir)
			if path_helpers.IsExistingDir(root) {
				err = filepath.Walk(root, func(realPath string, info os.FileInfo, err error) error {
					if err != nil {
						return err
					}
					if realPath == "." || realPath == root {
						return nil
					}

					pth := filepath.Join(fs.path, strings.TrimPrefix(realPath, root))
					if pth[0] == filepath.Separator {
						pth = pth[1:]
					}

					var inf api.FileInfo

					if info.IsDir() {
						if !mode.IsDirs() {
							return nil
						}

						inf = &RealDirFileInfo{&RealFileInfo{FSFileInfoBase{fs, pth}, info, realPath}}
					} else {
						if !mode.IsFiles() {
							return nil
						}
						inf = &RealFileInfo{FSFileInfoBase{fs, pth}, info, realPath}
					}
					return cb(inf)
				})
			}
			return
		})
		if err != nil {
			return
		}
	}

	if err == nil && fs.parent != nil && mode.IsParentLookUp() {
		if dir == "." {
			dir = fs.nameSpace
		} else {
			dir = filepath.Join(fs.nameSpace, dir)
		}
		return filesystemWalk(fs.parent.(*AssetFileSystem), dir, cb, mode ^ api.WalkNameSpacesLookUp)
	}
	return
}