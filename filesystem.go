package assetfs

import (
	"context"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/moisespsena-go/assetfs/local"

	"github.com/moisespsena-go/assetfs/assetfsapi"
	"github.com/moisespsena-go/file-utils"
	"github.com/moisespsena-go/path-helpers"
	"github.com/moisespsena/orderedmap"
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
	assetfsapi.AssetGetterInterface
	assetfsapi.TraversableInterface
	parent        assetfsapi.Interface
	paths         []string
	path          string
	nameSpaces    map[string]*AssetFileSystem
	nameSpace     string
	callbacks     []assetfsapi.PathRegisterCallback
	handler       http.Handler
	plugins       []assetfsapi.Plugin
	pathsFromFunc func(ctx context.Context, dir string, cb func(pth string) error) (err error)
	LocalSources  local.Sources
}

type RawFileSystem struct {
	*AssetFileSystem
}

func (r *RawFileSystem) rawPathsFrom(ctx context.Context, pth string, cb func(pth string) error) error {
	pth = filepath.FromSlash(path.Join(r.path, pth))
	_, err := os.Stat(pth)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	return cb(pth)
}

func (r *RawFileSystem) init() {
	r.pathsFromFunc = r.rawPathsFrom
	r.AssetFileSystem.init()
}

func NewAssetFileSystem() *AssetFileSystem {
	fs := &AssetFileSystem{}
	fs.init()
	return fs
}

func (fs *AssetFileSystem) init() {
	fs.AssetGetterInterface = &AssetGetter{
		fs: fs,
		AssetFunc: func(ctx context.Context, name string) (data []byte, err error) {
			var asset assetfsapi.AssetInterface
			asset, err = filesystemAsset(ctx, fs, name)
			if err != nil {
				return
			}
			return asset.Data()
		},
		AssetInfoFunc: func(ctx context.Context, path string) (assetfsapi.FileInfo, error) {
			return filesystemAssetInfo(ctx, fs, path)
		},
	}
	fs.TraversableInterface = &Traversable{
		FS: fs,
		WalkFunc: func(dir string, cb assetfsapi.CbWalkFunc, mode assetfsapi.WalkMode) error {
			return filesystemWalk(fs, dir, func(info assetfsapi.FileInfo) error {
				return cb(info.Path(), info.IsDir())
			}, mode)
		},
		WalkInfoFunc: func(dir string, cb assetfsapi.CbWalkInfoFunc, mode assetfsapi.WalkMode) error {
			return filesystemWalk(fs, dir, cb, mode)
		},
		GlobFunc: func(pattern assetfsapi.GlobPattern, cb func(pth string, isDir bool) error) (err error) {
			return filesystemGlob(fs, pattern, cb)
		},
		GlobInfoFunc: func(pattern assetfsapi.GlobPattern, cb func(info assetfsapi.FileInfo) error) (err error) {
			return filesystemGlobInfo(fs, pattern, cb)
		},
	}
}

func (fs *AssetFileSystem) OnPathRegister(cb ...assetfsapi.PathRegisterCallback) {
	fs.callbacks = append(fs.callbacks, cb...)
}

func (fs *AssetFileSystem) GetPath() string {
	return fs.path
}

// RegisterPath register view paths
func (fs *AssetFileSystem) RegisterPath(pth string, ignoreExists ...bool) error {
	_, err := fs.registerPath(pth, false, ignoreExists...)
	return err
}

// RegisterPath register view paths
func (fs *AssetFileSystem) RegisterPathFS(pth string, ignoreExists ...bool) (assetfsapi.Interface, error) {
	return fs.registerPath(pth, false, ignoreExists...)
}

// PrependPath prepend path to view paths
func (fs *AssetFileSystem) PrependPath(pth string, ignoreExists ...bool) error {
	_, err := fs.registerPath(pth, true, ignoreExists...)
	return err
}

// PrependPath prepend path to view paths
func (fs *AssetFileSystem) PrependPathFS(pth string, ignoreExists ...bool) (assetfsapi.Interface, error) {
	return fs.registerPath(pth, true, ignoreExists...)
}

// RegisterPath register view paths
func (fs *AssetFileSystem) registerPath(pth string, prepend bool, ignoreExists ...bool) (assetfsapi.Interface, error) {
	var onlyExists = true
	for _, ige := range ignoreExists {
		if ige {
			onlyExists = false
			break
		}
	}
	pth = filepath.Clean(pth)
	var pfs assetfsapi.Interface
	if _, err := os.Stat(pth); !onlyExists || !os.IsNotExist(err) {
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

			pfs = fs.newRawFS(pth)

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

func (fs *AssetFileSystem) GetNameSpace(nameSpace string) (assetfsapi.NameSpacedInterface, error) {
	var (
		ns *AssetFileSystem
		ok bool
	)
	for _, name := range strings.Split(strings.Trim(nameSpace, "/"), "/") {
		if fs.nameSpaces == nil {
			return nil, os.ErrNotExist
		} else if ns, ok = fs.nameSpaces[name]; !ok {
			return nil, os.ErrNotExist
		}
		fs = ns
	}
	return ns, nil
}

func (fs *AssetFileSystem) NameSpaces() (items []assetfsapi.NameSpacedInterface) {
	for _, v := range fs.nameSpaces {
		items = append(items, v)
	}
	return
}

// NameSpace return namespaced filesystem
func (fs *AssetFileSystem) NameSpace(nameSpace string) assetfsapi.NameSpacedInterface {
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

func (fs *AssetFileSystem) newPathNameSpace(pth string) assetfsapi.NameSpacedInterface {
	ns := &AssetFileSystem{nameSpace: pth, path: pth}
	ns.parent = fs
	ns.init()
	return ns
}

func (fs *AssetFileSystem) newRawFS(pth string) assetfsapi.Interface {
	ns := &AssetFileSystem{paths: []string{pth}, path: fs.path}
	rfs := &RawFileSystem{ns}
	rfs.init()
	return rfs
}

func (fs *AssetFileSystem) GetName() string {
	return fs.nameSpace
}

func (fs *AssetFileSystem) GetParent() assetfsapi.Interface {
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

func (fs *AssetFileSystem) ReadDir(dir string, cb assetfsapi.CbWalkInfoFunc, skipDir bool) (err error) {
	return fs.readDir(dir, cb, false, skipDir)
}

func (fs *AssetFileSystem) readDir(dir string, cb assetfsapi.CbWalkInfoFunc, parentLookup bool, skipDir bool) (err error) {
	if dir == "" {
		dir = "."
	}

	if !skipDir && fs.nameSpaces != nil {
		if dir == "." {
			for nsName, ns := range fs.nameSpaces {
				err = cb(&NameSpaceFileInfo{assetfsapi.NewCleanedBasicFileInfo(ns.path, nsName), ns})
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
		var inf assetfsapi.FileInfo
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

				inf = &RealDirFileInfo{&RealFileInfo{basicFileInfo(pth, info), realPath}}
			} else {
				inf = &RealFileInfo{basicFileInfo(pth, info), realPath}
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
			pth = filepath.Join(pth, dir)
			if path_helpers.IsExistingDir(pth) {
				err = dolsdir(pth)
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
			dir = fs.nameSpace + string(os.PathSeparator) + dir
		}
		return fs.parent.(*AssetFileSystem).readDir(dir, cb, parentLookup, skipDir)
	}
	return
}

func (fs *AssetFileSystem) PathsFrom(ctx context.Context, pth string, cb func(pth string) error) (err error) {
	var parent = fs
	for parent != nil {
		if parent.parent == nil {
			break
		}
		parent = parent.parent.(*AssetFileSystem)
	}

	for _, name := range local.GetLocalNames(ctx) {
		if src := fs.LocalSources.Get(name); src != nil {
			if dirPth, ok := src.GetDir(pth); ok {
				if err = cb(dirPth); err != nil {
					return
				}
			}
		}
	}

	err = fs.pathsFrom(ctx, pth, cb)
	if err == io.EOF {
		return nil
	}
	return
}

func (fs *AssetFileSystem) pathsFrom(ctx context.Context, dir string, cb func(pth string) error) (err error) {
	if fs.pathsFromFunc != nil {
		return fs.pathsFromFunc(ctx, dir, cb)
	}

	if dir == "" {
		dir = "."
	}

	if dir == "." {
		parent := fs
		prefix := ""
		for parent != nil {
			for _, pth := range parent.paths {
				pth = path.Join(pth, prefix)

				if _, err = os.Stat(pth); err != nil {
					if os.IsNotExist(err) {
						err = nil
						continue
					}
					return err
				}

				if err = cb(pth); err != nil {
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
			return ns.pathsFrom(ctx, parts[1], cb)
		}
	}

	x := fs.newPathNameSpace(dir).(*AssetFileSystem).pathsFrom(ctx, ".", cb)
	return x
}

func (fs *AssetFileSystem) GetPaths(recursive ...bool) (p []*fileutils.Dir) {
	rec := len(recursive) > 0 && recursive[0]
	fspath := fs.path
	if fspath == "" {
		fspath = "."
	}
	for _, pth := range fs.paths {
		p = append(p, &fileutils.Dir{Src: pth, Destation: fileutils.Destation{fspath}})
	}
	if rec && fs.nameSpaces != nil {
		for _, ns := range fs.nameSpaces {
			p = append(p, ns.GetPaths(true)...)
		}
	}
	return
}

func (fs *AssetFileSystem) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if fs.handler == nil {
		fs.handler = HTTPStaticHandler(fs)
	}
	fs.handler.ServeHTTP(w, r)
}

func (fs *AssetFileSystem) RegisterPlugin(plugins ...assetfsapi.Plugin) {
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

func (fs *AssetFileSystem) DumpFiles(cb func(info assetfsapi.FileInfo) error) error {
	return fs.dump(nil, true, cb)
}

func (fs *AssetFileSystem) DumpFilesC(ctx context.Context, cb func(info assetfsapi.FileInfo) error) error {
	return fs.dump(ctx, true, cb)
}

func (fs *AssetFileSystem) Dump(cb func(info assetfsapi.FileInfo) error, ignore ...func(pth string) bool) error {
	return fs.dump(nil, false, cb, ignore...)
}

func (fs *AssetFileSystem) DumpC(ctx context.Context, cb func(info assetfsapi.FileInfo) error, ignore ...func(pth string) bool) error {
	return fs.dump(ctx, false, cb, ignore...)
}

func (fs *AssetFileSystem) TreeNames(ctx context.Context, onlyFiles bool, ignore ...func(pth string) bool) (result []assetfsapi.FileInfo, err error) {
	m := map[string]assetfsapi.FileInfo{}
	err = fs.WalkInfo(".", func(info assetfsapi.FileInfo) error {
		pth := info.Path()
		if pth == "." {
			return nil
		}
		if onlyFiles && info.IsDir() {
			return nil
		}
		for _, ignore := range ignore {
			if ignore(pth) {
				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
		}

		if !info.IsDir() {
			for _, localName := range local.GetLocalNames(ctx) {
				if src := fs.LocalSources.Get(localName); localAssets != nil {
					pth := info.Path()
					if rpth, stat, ok := src.Get(pth); ok {
						m[info.Path()] = &RealFileInfo{basicFileInfo(pth, stat), rpth}
						return nil
					}
				}
			}
		}

		m[info.Path()] = info

		return nil
	}, assetfsapi.WalkAll)
	if err != nil {
		return nil, err
	}
	names := make([]string, len(m))
	result = make([]assetfsapi.FileInfo, len(m))
	i := 0

	for name, _ := range m {
		names[i] = name
		i++
	}

	sort.Strings(names)

	for i, name := range names {
		result[i] = m[name]
		delete(m, name)
	}

	return result, nil
}

func (fs *AssetFileSystem) dump(ctx context.Context, onlyFiles bool, cb func(info assetfsapi.FileInfo) error, ignore ...func(pth string) bool) error {
	files, err := fs.TreeNames(ctx, onlyFiles, ignore...)
	if err != nil {
		return err
	}
	for _, f := range files {
		err = cb(f)
		if err != nil {
			return err
		}
	}

	return nil
}
