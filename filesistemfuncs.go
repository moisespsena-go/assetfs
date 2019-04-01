package assetfs

import (
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/moisespsena-go/os-common"

	"github.com/moisespsena/go-assetfs/assetfsapi"
	"github.com/moisespsena-go/path-helpers"
)

type FileInfoAsset struct {
	info assetfsapi.FileInfo
	name string
	data []byte
}

func (a *FileInfoAsset) GetName() string {
	return a.name
}

func (a *FileInfoAsset) GetData() []byte {
	return a.data
}

func (a *FileInfoAsset) GetString() string {
	return string(a.data)
}

func (a *FileInfoAsset) GetPath() string {
	return a.info.RealPath()
}

// Names list matched files from assetfs
func filesystemGlob(fs *AssetFileSystem, pattern assetfsapi.GlobPattern, cb func(pth string, isDir bool) error) error {
	return filesystemGlobInfo(fs, pattern, func(info assetfsapi.FileInfo) error {
		return cb(info.Path(), info.IsDir())
	})
}

// Names list matched files from assetfs
func filesystemGlobInfo(fs *AssetFileSystem, pattern assetfsapi.GlobPattern, cb func(info assetfsapi.FileInfo) error) error {
	set := make(map[string]bool)
	cb2 := func(info assetfsapi.FileInfo) error {
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
		return fs.WalkInfo(pattern.Dir(), cb2, assetfsapi.WalkAll^assetfsapi.WalkDirs)
	}
	return fs.readDir(pattern.Dir(), cb2, true, true)
}

// Asset get content with name from assetfs
func filesystemAsset(fs *AssetFileSystem, name string) (assetfsapi.AssetInterface, error) {
	info, err := filesystemAssetInfo(fs, name)
	if err != nil {
		return nil, err
	}
	data, err := info.Data()
	if err != nil {
		return nil, err
	}
	return &FileInfoAsset{info, name, data}, nil
}

func filesystemAssetInfo(fs *AssetFileSystem, pth string) (info assetfsapi.FileInfo, err error) {
	var r string
	dir, base := filepath.Split(pth)
	err = fs.PathsFrom(dir, func(pth string) error {
		pth = filepath.Join(pth, base)
		if _, err := os.Stat(pth); err == nil {
			r = pth
			return io.EOF
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	if r == "" {
		return nil, oscommon.ErrNotFound(pth)
	}
	stat, err := os.Stat(r)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, oscommon.ErrNotFound(pth)
		}
		return nil, err
	}
	return &RealFileInfo{FSFileInfoBase{fs, pth}, stat, r}, nil
}

func filesystemWalk(fs *AssetFileSystem, dir string, cb assetfsapi.CbWalkInfoFunc, mode assetfsapi.WalkMode) (err error) {
	if dir == "" {
		dir = "."
	}

	if dir == "." {
		if fs.nameSpaces != nil {
			for _, ns := range fs.nameSpaces {
				err = filesystemWalk(ns, ".", func(info assetfsapi.FileInfo) error {
					npth := strings.TrimPrefix(ns.path, fs.path)
					if npth[0] == '/' {
						npth = npth[1:]
					}
					switch t := info.(type) {
					case *RealDirFileInfo:
						t.path = filepath.Join(npth, t.path)
					case *RealFileInfo:
						t.path = filepath.Join(npth, t.path)
					}
					return cb(info)
				}, mode|assetfsapi.WalkNameSpacesLookUp^assetfsapi.WalkParentLookUp)
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
				var inf assetfsapi.FileInfo
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
				err = filesystemWalk(ns, parts[1], cb, mode|assetfsapi.WalkNameSpacesLookUp^assetfsapi.WalkParentLookUp)
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

					var inf assetfsapi.FileInfo

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
		if mode.IsNameSpacesLookUp() {
			mode ^= assetfsapi.WalkNameSpacesLookUp
		}
		return filesystemWalk(fs.parent.(*AssetFileSystem), dir, cb, mode)
	}
	return
}
