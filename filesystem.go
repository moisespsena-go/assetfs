package assetfs

import (
	"os"
	"fmt"
	"errors"
	"net/http"
	"path/filepath"
	"github.com/moisespsena/go-path-helpers"
	"github.com/moisespsena/go-assetfs/repository"
)

// AssetFileSystem AssetFS based on FileSystem
type AssetFileSystem struct {
	AssetGetterInterface
	parent     Interface
	paths      []string
	path       string
	nameSpaces map[string]NameSpacedInterface
	nameSpace  string
	callbacks  []PathRegisterCallback
	handler    http.Handler
}

func NewAssetFileSystem() *AssetFileSystem {
	fs := &AssetFileSystem{}
	fs.init()
	return fs
}

func (fs *AssetFileSystem) init() {
	fs.AssetGetterInterface = &AssetGetter{
		AssetFunc: func(name string) (data []byte, err error) {
			var asset AssetInterface
			asset, err = filesystemAsset(fs, name)
			if err != nil {
				return
			}
			return asset.GetData(), nil
		},
		AssetInfoFunc: func(path string) (os.FileInfo, error) {
			return filesystemAssetInfo(fs, path)
		},
		GlobFunc: func(path string, recursive ...bool) (matches []string, err error) {
			return filesystemGlob(fs, path, recursive...)
		},
	}
}

func (fs *AssetFileSystem) OnPathRegister(cb ...PathRegisterCallback) {
	fs.callbacks = append(fs.callbacks, cb...)
}

func (fs *AssetFileSystem) GetPath() string {
	return fs.path
}

// RegisterPath register view paths
func (fs *AssetFileSystem) RegisterPath(pth string) error {
	return fs.registerPath(pth, false)
}

// PrependPath prepend path to view paths
func (fs *AssetFileSystem) PrependPath(pth string) error {
	return fs.registerPath(pth, true)
}

// RegisterPath register view paths
func (fs *AssetFileSystem) registerPath(pth string, prepend bool) error {
	pth = filepath.Clean(pth)
	if _, err := os.Stat(pth); !os.IsNotExist(err) {
		var existing bool
		for _, p := range fs.paths {
			if p == pth {
				existing = true
				break
			}
		}
		if !existing {
			if prepend {
				fs.paths = append([]string{pth}, fs.paths...)
			} else {
				fs.paths = append(fs.paths, pth)
			}
			pthFS := fs.newPathNameSpace(pth)

			for _, cb := range fs.callbacks {
				cb(pthFS)
			}
		}

		return nil
	}
	return errors.New("not found")
}

// Compile compile assetfs
func (fs *AssetFileSystem) Compile() error {
	return nil
}

// NameSpace return namespaced filesystem
func (fs *AssetFileSystem) NameSpace(nameSpace string) NameSpacedInterface {
	if fs.nameSpaces == nil {
		fs.nameSpaces = make(map[string]NameSpacedInterface)
	}
	if ns, ok := fs.nameSpaces[nameSpace]; ok {
		return ns
	}
	path := nameSpace
	if fs.path != "" {
		path = filepath.Join(fs.path, path)
	}
	ns := &AssetFileSystem{path: path, parent: fs, nameSpace: nameSpace,
		AssetGetterInterface: fs.AssetGetterInterface}
	ns.init()
	fs.nameSpaces[nameSpace] = ns
	return ns
}

// NameSpace return namespaced filesystem
func (fs *AssetFileSystem) newPathNameSpace(path string) NameSpacedInterface {
	return &AssetFileSystem{path: path, parent: fs, nameSpace: "",
		AssetGetterInterface: fs.AssetGetterInterface}
}

func (fs *AssetFileSystem) GetNameSpace() string {
	return fs.nameSpace
}

func (fs *AssetFileSystem) GetParent() Interface {
	return fs.parent
}

func (fs *AssetFileSystem) WalkFiles(dir string, cb WalkFunc, onError ...OnErrorFunc) (err error) {
	return fs.WalkFilesInfo(dir, func(prefix, path string, info os.FileInfo, err error) error {
		return cb(prefix, path, err)
	}, onError...)
}

func (fs *AssetFileSystem) WalkFilesInfo(dir string, cb WalkInfoFunc, onError ...OnErrorFunc) (err error) {
	return fs.walkFilesInfo(dir, cb, onError...)
}

func (fs *AssetFileSystem) walkFilesInfo(dir string, cb WalkInfoFunc, onError ...OnErrorFunc) (err error) {
	if len(onError) == 0 {
		onError = append(onError, nil)
	}

	var walkFileInfo WalkInfoFunc

	if onError[0] == nil {
		walkFileInfo = cb
	} else {
		walkFileInfo = func(prefix, path string, info os.FileInfo, err error) error {
			if err == nil {
				err = cb(prefix, path, info, err)
			}
			if err != nil {
				if !onError[0](path, err) {
					return err
				}
			}
			return nil
		}
	}

	if dir == "" {
		for _, pth := range fs.paths {
			err = filepath.Walk(pth, func(path string, info os.FileInfo, err error) error {
				return walkFileInfo(pth+string(filepath.Separator), path, info, err)
			})
			if err != nil {
				break
			}
		}
	} else {
		for _, pth := range fs.paths {
			pth2 := filepath.Join(pth, dir)
			if path_helpers.IsExistingDir(pth2) {
				err = filepath.Walk(pth2, func(path string, info os.FileInfo, err error) error {
					return walkFileInfo(pth2+string(filepath.Separator), path, info, err)
				})
				if err != nil {
					break
				}
			}
		}
	}

	if err == nil && fs.parent != nil {
		if dir == "" {
			dir = fs.nameSpace
		} else {
			dir += fmt.Sprint(os.PathSeparator, fs.path)
		}
		return fs.parent.(*AssetFileSystem).walkFilesInfo(dir, walkFileInfo, onError...)
	}
	return
}

func (fs *AssetFileSystem) GetPaths(recursive ...bool) (p []*path_helpers.Path) {
	rec := len(recursive) > 0 && recursive[0]
	fspath := fs.path
	if fspath == "" {
		fspath = "."
	}
	for _, pth := range fs.paths {
		p = append(p, &path_helpers.Path{Real: pth, Alias: fspath})
	}
	if rec && fs.nameSpaces != nil {
		for _, ns := range fs.nameSpaces {
			p = append(p, ns.(*AssetFileSystem).GetPaths(true)...)
		}
	}
	return
}

func (fs *AssetFileSystem) NewRepository(pkg string) repository.Interface {
	repo := repository.NewRepository(pkg)
	repo.AddSourcePath(fs.GetPaths(true)...)
	return repo
}

func (fs *AssetFileSystem) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if fs.handler == nil {
		fs.handler = HTTPStaticHandler(fs)
	}
	fs.handler.ServeHTTP(w, r)
}
