package assetfs

import (
	"os"
	"fmt"
	"errors"
	"strings"
	"net/http"
	"path/filepath"
	"github.com/moisespsena/go-path-helpers"
	rapi "github.com/moisespsena/go-assetfs/repository/api"
	"github.com/moisespsena/orderedmap"
	"io/ioutil"
	"sort"
	"io"
	"github.com/moisespsena/go-assetfs/api"
	"github.com/moisespsena/go-assetfs/repository"
)

type assetFileSystemNameSpaces struct {
	*orderedmap.OrderedMap
}

func (a *assetFileSystemNameSpaces) Get(key string) (*AssetFileSystem, bool) {
	if v, ok := a.OrderedMap.Get(key); ok {
		return v.(*AssetFileSystem), true
	}
	return nil, false
}

// AssetFileSystem AssetFS based on FileSystem
type AssetFileSystem struct {
	api.AssetGetterInterface
	api.TraversableInterface
	parent     api.Interface
	paths      []string
	path       string
	nameSpaces map[string]*AssetFileSystem
	nameSpace  string
	callbacks  []api.PathRegisterCallback
	handler    http.Handler
	plugins    []api.Plugin
}

func NewAssetFileSystem() *AssetFileSystem {
	fs := &AssetFileSystem{}
	fs.init()
	return fs
}

func (fs *AssetFileSystem) init() {
	fs.AssetGetterInterface = &AssetGetter{
		fs: fs,
		AssetFunc: func(name string) (data []byte, err error) {
			var asset api.AssetInterface
			asset, err = filesystemAsset(fs, name)
			if err != nil {
				return
			}
			return asset.GetData(), nil
		},
		AssetInfoFunc: func(path string) (api.FileInfo, error) {
			return filesystemAssetInfo(fs, path)
		},
	}
	fs.TraversableInterface = &Traversable{
		fs: fs,
		WalkFunc: func(dir string, cb api.CbWalkFunc, mode api.WalkMode) error {
			return filesystemWalk(fs, dir, func(info api.FileInfo) error {
				return cb(info.Path(), info.IsDir())
			}, mode)
		},
		WalkInfoFunc: func(dir string, cb api.CbWalkInfoFunc, mode api.WalkMode) error {
			return filesystemWalk(fs, dir, cb, mode)
		},
		GlobFunc: func(pattern api.GlobPattern, cb func(pth string, isDir bool) error) (err error) {
			return filesystemGlob(fs, pattern, cb)
		},
		GlobInfoFunc: func(pattern api.GlobPattern, cb func(info api.FileInfo) error) (err error) {
			return filesystemGlobInfo(fs, pattern, cb)
		},
	}
}

func (fs *AssetFileSystem) OnPathRegister(cb ...api.PathRegisterCallback) {
	fs.callbacks = append(fs.callbacks, cb...)
}

func (fs *AssetFileSystem) GetPath() string {
	return fs.path
}

// RegisterPath register view paths
func (fs *AssetFileSystem) RegisterPath(pth string) error {
	_, err := fs.registerPath(pth, false)
	return err
}

// RegisterPath register view paths
func (fs *AssetFileSystem) RegisterPathFS(pth string) (api.Interface, error) {
	return fs.registerPath(pth, false)
}

// PrependPath prepend path to view paths
func (fs *AssetFileSystem) PrependPath(pth string) error {
	_, err := fs.registerPath(pth, true)
	return err
}

// PrependPath prepend path to view paths
func (fs *AssetFileSystem) PrependPathFS(pth string) (api.Interface, error) {
	return fs.registerPath(pth, true)
}

// RegisterPath register view paths
func (fs *AssetFileSystem) registerPath(pth string, prepend bool) (api.Interface, error) {
	pth = filepath.Clean(pth)
	var pfs api.Interface
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

			pfs = fs.newPathNameSpace(pth)
			pfs.(*AssetFileSystem).path = fs.path

			for _, plugin := range fs.plugins {
				plugin.PathRegisterCallback(pfs)
			}

			for _, cb := range fs.callbacks {
				cb(pfs)
			}
		}

		return pfs, nil
	}
	return nil, errors.New("not found")
}

// Compile compile assetfs
func (fs *AssetFileSystem) Compile() error {
	return nil
}

// NameSpace return namespaced filesystem
func (fs *AssetFileSystem) NameSpace(nameSpace string) api.NameSpacedInterface {
	var (
		ns *AssetFileSystem
		ok bool
	)
	for _, name := range strings.Split(strings.Trim(nameSpace, "/"), "/") {
		if fs.nameSpaces == nil {
			fs.nameSpaces = make(map[string]*AssetFileSystem)
		}
		if ns, ok = fs.nameSpaces[name]; !ok {
			path := name
			if fs.path != "" {
				path = filepath.Join(fs.path, path)
			}
			ns = &AssetFileSystem{path: path, parent: fs, nameSpace: name, plugins: fs.plugins}
			ns.init()
			fs.nameSpaces[name] = ns
		}
		fs = ns
	}
	return fs
}

func (fs *AssetFileSystem) newPathNameSpace(pth string) api.NameSpacedInterface {
	ns := &AssetFileSystem{nameSpace: pth, path: pth, paths:[]string{pth}}
	ns.init()
	return ns
}

func (fs *AssetFileSystem) GetNameSpace() string {
	return fs.nameSpace
}

func (fs *AssetFileSystem) GetParent() api.Interface {
	return fs.parent
}

func (fs *AssetFileSystem) eachPath(reverse bool, cb func(pth string) error) (err error) {
	if reverse {
		for _, pth := range fs.paths {
			err = cb(pth)
			if err != nil {
				return err
			}
		}
	} else {
		for i := len(fs.paths); i > 0; i-- {
			err = cb(fs.paths[i-1])
			if err != nil {
				return err
			}
		}
	}
	return
}

func (fs *AssetFileSystem) ReadDir(dir string, cb api.CbWalkInfoFunc, skipDir bool) (err error) {
	return fs.readDir(dir, cb, false, skipDir)
}

func (fs *AssetFileSystem) readDir(dir string, cb api.CbWalkInfoFunc, parentLookup bool, skipDir bool) (err error) {
	if dir == "" {
		dir = "."
	}

	if !skipDir && fs.nameSpaces != nil {
		if dir == "." {
			for nsName, ns := range fs.nameSpaces {
				err = cb(&NameSpaceFileInfo{FSFileInfoBase{fs, ns.path}, nsName, ns})
				if err != nil {
					return err
				}
			}
		} else {
			parts := strings.SplitN(dir, string(os.PathSeparator), 2)
			if ns, ok := fs.nameSpaces[parts[0]]; ok {
				err = ns.readDir(parts[1], cb, false, skipDir)
				if err != nil {
					return err
				}
			}
		}
	}

	dolsdir := func(root string) (err error) {
		var inf api.FileInfo
		var realPath, pth string
		ites, err := ioutil.ReadDir(root)
		if err != nil {
			return err
		}
		for _, info := range ites {
			realPath = filepath.Join(root, info.Name())
			pth = filepath.Join(fs.path, info.Name())
			if info.IsDir() {
				if skipDir {
					return nil
				}

				inf = &RealDirFileInfo{&RealFileInfo{FSFileInfoBase{fs, pth}, info, realPath}}
			} else {
				inf = &RealFileInfo{FSFileInfoBase{fs, pth}, info, realPath}
			}
			err = cb(inf)
			if err != nil {
				return err
			}
		}
		return nil
	}
	if dir == "." {
		for _, pth := range fs.paths {
			err = dolsdir(pth)
			if err != nil {
				return err
			}
		}
	} else {
		for _, pth := range fs.paths {
			pth2 := filepath.Join(pth, dir)
			if path_helpers.IsExistingDir(pth2) {
				err = dolsdir(pth2)
				if err != nil {
					return err
				}
			}
		}
	}

	if err == nil && fs.parent != nil && parentLookup {
		if dir == "" {
			dir = fs.nameSpace
		} else {
			dir += fmt.Sprint(os.PathSeparator, fs.path)
		}
		return fs.parent.(*AssetFileSystem).readDir(dir, cb, parentLookup, skipDir)
	}
	return
}

func (fs *AssetFileSystem) PathsFrom(pth string, cb func(pth string) error) (err error) {
	err = fs.pathsFrom(pth, cb)
	if err == io.EOF {
		return nil
	}
	return
}

func (fs *AssetFileSystem) pathsFrom(dir string, cb func(pth string) error) (err error) {
	if dir == "" {
		dir = "."
	}

	if dir == "." {
		parent := fs
		prefix := ""
		for parent != nil {
			for _, pth := range parent.paths {
				pth = filepath.Join(pth, prefix)
				_, err := os.Stat(pth)
				if err != nil {
					if os.IsNotExist(err) {
						continue
					}
					return err
				}
				err = cb(pth)
				if err != nil {
					return err
				}
			}
			if parent.parent == nil {
				return
			}
			prefix = filepath.Join(parent.nameSpace, prefix)
			parent = parent.parent.(*AssetFileSystem)
		}
		return
	}

	if fs.nameSpaces != nil {
		parts := strings.SplitN(dir, string(os.PathSeparator), 2)
		if ns, ok := fs.nameSpaces[parts[0]]; ok {
			return ns.pathsFrom(parts[1], cb)
		}
	}

	return fs.newPathNameSpace(dir).(*AssetFileSystem).pathsFrom(".", cb)
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
			p = append(p, ns.GetPaths(true)...)
		}
	}
	return
}

func (fs *AssetFileSystem) NewRepository(pkg string) rapi.Interface {
	repo := repository.NewRepository(pkg)
	repo.Dumper(func(cb func(pth string, stat os.FileInfo, reader io.Reader) error) error {
		return fs.Dump(func(info api.FileInfo) error {
			if info.IsDir() {
				return cb(info.Path(), info, nil)
			}
			reader, err := info.Reader()
			if err != nil {
				return fmt.Errorf("Get Reader of %q [%q] fail: %v", info.Path(), info.RealPath(), err)
			}
			defer reader.Close()
			return cb(info.Path(), info, reader)
		})
	})
	return repo
}

func (fs *AssetFileSystem) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if fs.handler == nil {
		fs.handler = HTTPStaticHandler(fs)
	}
	fs.handler.ServeHTTP(w, r)
}

func (fs *AssetFileSystem) RegisterPlugin(plugins ...api.Plugin) {
	for _, p := range plugins {
		p.Init(fs)
	}
	for _, pth := range fs.paths {
		pthFS := fs.newPathNameSpace(pth)
		for _, p := range plugins {
			p.PathRegisterCallback(pthFS)
		}
	}
	if fs.path != "" {
		for _, p := range plugins {
			p.PathRegisterCallback(fs)
		}
	}
	fs.plugins = append(fs.plugins, plugins...)
}
func (fs *AssetFileSystem) DumpFiles(cb func(info api.FileInfo) error) error {
	return fs.dump(true, cb)
}

func (fs *AssetFileSystem) Dump(cb func(info api.FileInfo) error) error {
	return fs.dump(false, cb)
}

func (fs *AssetFileSystem) dump(onlyFiles bool, cb func(info api.FileInfo) error) error {
	m := map[string]api.FileInfo{}
	err := fs.WalkInfo(".", func(info api.FileInfo) error {
		if info.Path() == "." {
			return nil
		}
		if onlyFiles && info.IsDir() {
			return nil
		}
		m[info.Path()] = info
		return nil
	}, api.WalkAll)
	if err != nil {
		return err
	}
	names := make([]string, len(m))
	i := 0

	for name, _ := range m {
		names[i] = name
		i++
	}

	sort.Strings(names)

	for _, name := range names {
		err = cb(m[name])
		if err != nil {
			return err
		}
	}

	return nil
}
