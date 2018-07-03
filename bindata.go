package assetfs

import (
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"github.com/moisespsena/go-assetfs/api"
	"github.com/moisespsena/go-assetfs/repository"
	rapi "github.com/moisespsena/go-assetfs/repository/api"
	"github.com/moisespsena/go-path-helpers"
)

var ERR_BINDATA_FILE = errors.New("Bindata file.")
var ERR_BINDATA_DIR = errors.New("Bindata dir.")

type BindataFileInfoBase struct {
	realPath string
}

func (b *BindataFileInfoBase) Path() string {
	return b.realPath
}

func (b *BindataFileInfoBase) RealPath() string {
	return b.realPath
}
func (b *BindataFileInfoBase) TrimPrefix(prefix string) {
	b.realPath = strings.TrimPrefix(b.realPath, prefix)
	if b.realPath[0] == filepath.Separator {
		b.realPath = b.realPath[1:]
	}
}

type BindataFileInfo struct {
	os.FileInfo
	BindataFileInfoBase
	AssetFunc func() ([]byte, error)
}

func NewBindataFileInfo(info os.FileInfo, pth string, assetFunc func() ([]byte, error)) *BindataFileInfo {
	return &BindataFileInfo{info, BindataFileInfoBase{pth}, assetFunc}
}

func (BindataFileInfo) Type() api.FileType {
	return api.FileTypeBindata | api.FileTypeNormal
}

func (rf *BindataFileInfo) Reader() (io.ReadCloser, error) {
	return os.Open(rf.realPath)
}

func (BindataFileInfo) Writer() (io.WriteCloser, error) {
	return nil, ERR_BINDATA_FILE
}

func (BindataFileInfo) Appender() (io.WriteCloser, error) {
	return nil, ERR_BINDATA_FILE
}

func (b *BindataFileInfo) Data() ([]byte, error) {
	return b.AssetFunc()
}

func (b *BindataFileInfo) PkgPath() string {
	return reflect.TypeOf(b.AssetFunc).PkgPath()
}

func (b *BindataFileInfo) String() string {
	return StringifyFileInfo(b)
}

type BindataDirInfo struct {
	BindataFileInfoBase
	ChildrenFunc func(cb func(info api.FileInfo) error) error
}

func NewBindataDirInfo(pth string, childrenFunc func(cb func(info api.FileInfo) error) error) *BindataDirInfo {
	return &BindataDirInfo{BindataFileInfoBase{pth}, childrenFunc}
}

func (BindataDirInfo) Type() api.FileType {
	return api.FileTypeBindata | api.FileTypeDir
}

func (b BindataDirInfo) Name() string {
	return filepath.Base(b.realPath)
}

func (b BindataDirInfo) Size() int64 {
	return -1
}

func (b BindataDirInfo) Mode() os.FileMode {
	return os.ModeDir
}

func (b BindataDirInfo) ModTime() time.Time {
	return now
}

func (b BindataDirInfo) IsDir() bool {
	return true
}

func (b BindataDirInfo) Sys() interface{} {
	return nil
}

func (b *BindataDirInfo) String() string {
	return StringifyFileInfo(b)
}

func (rf *BindataDirInfo) Reader() (io.ReadCloser, error) {
	return nil, ERR_BINDATA_DIR
}

func (BindataDirInfo) Writer() (io.WriteCloser, error) {
	return nil, ERR_BINDATA_DIR
}

func (BindataDirInfo) Appender() (io.WriteCloser, error) {
	return nil, ERR_BINDATA_DIR
}

func (b *BindataDirInfo) Data() ([]byte, error) {
	return nil, ERR_BINDATA_DIR
}

func (b *BindataDirInfo) PkgPath() string {
	return reflect.TypeOf(b.ChildrenFunc).PkgPath()
}

func (b *BindataDirInfo) ReadDir(cb func(child api.FileInfo) error) (err error) {
	return b.ChildrenFunc(cb)
}

type BindataFileSystem struct {
	api.AssetGetterInterface
	api.TraversableInterface
	getAssetFunc     api.AssetReaderFunc
	getAssetInfoFunc api.GetAssetInfoFunc
	fileWalkFunc     api.WalkFunc
	fileWalkInfoFunc api.WalkInfoFunc
	globFunc         api.GlobFunc
	globInfoFunc     api.GlobInfoFunc
	root             *BindataFileSystem
	path             string
	parent           *BindataFileSystem
	nameSpaces       map[string]*BindataFileSystem
	nameSpace        string
	callbacks        []api.PathRegisterCallback
	HttpHandler      http.Handler
	notExists        bool
}

func NewBindataFileSystem(getAsset func(name string) ([]byte, error), getAssetInfo func(pth string) (api.FileInfo, error), fileWalk api.WalkFunc,
	fileWalkInfo api.WalkInfoFunc, glob api.GlobFunc, globInfo api.GlobInfoFunc) (fs *BindataFileSystem) {
	fs = &BindataFileSystem{
		getAssetFunc:     getAsset,
		getAssetInfoFunc: getAssetInfo,
		fileWalkFunc:     fileWalk,
		fileWalkInfoFunc: fileWalkInfo,
		globFunc:         glob,
		globInfoFunc:     globInfo,
	}
	fs.init()
	return
}

func (fs *BindataFileSystem) init() {
	fs.AssetGetterInterface = &AssetGetter{
		AssetFunc: func(path string) ([]byte, error) {
			return bindataAsset(fs, fs.getAssetFunc, path)
		},
		AssetInfoFunc: func(path string) (api.FileInfo, error) {
			return bindataAssetInfo(fs, fs.getAssetInfoFunc, path)
		},
	}
	fs.TraversableInterface = &Traversable{
		fs,
		fs.fileWalkFunc,
		fs.fileWalkInfoFunc,
		func(dir string, cb api.CbWalkInfoFunc, skipDir bool) error {
			return nil
		},
		func(pattern api.GlobPattern, cb func(pth string, isDir bool) error) (err error) {
			return bindataGlob(fs, fs.globFunc, pattern, cb)
		},
		func(pattern api.GlobPattern, cb func(info api.FileInfo) error) error {
			return bindataGlobInfo(fs, fs.globInfoFunc, pattern, cb)
		},
	}
	if fs.parent != nil {
		asset, err := fs.parent.AssetInfo(fs.nameSpace)
		if err != nil || !asset.IsDir() {
			fs.notExists = true
		}
	}
}

func (fs *BindataFileSystem) OnPathRegister(cb ...api.PathRegisterCallback) {
	fs.callbacks = append(fs.callbacks, cb...)
}

func (fs *BindataFileSystem) GetPath() string {
	return fs.path
}

// RegisterPath register view paths
func (fs *BindataFileSystem) RegisterPath(pth string) error {
	return nil
}

// PrependPath prepend path to view paths
func (fs *BindataFileSystem) PrependPath(pth string) error {
	return nil
}

func (fs *BindataFileSystem) Root() api.Interface {
	return fs.root
}

// Compile compile assetfs
func (fs *BindataFileSystem) Compile() error {
	return nil
}

// NameSpace return namespaced filesystem
func (fs *BindataFileSystem) NameSpace(nameSpace string) api.NameSpacedInterface {
	if nameSpace == "" || nameSpace == "." {
		return nil
	}

	var (
		ns *BindataFileSystem
		ok bool
	)
	for _, name := range strings.Split(strings.Trim(nameSpace, "/"), "/") {
		if fs.nameSpaces == nil {
			fs.nameSpaces = make(map[string]*BindataFileSystem)
		}
		if ns, ok = fs.nameSpaces[name]; !ok {
			path := name
			root := fs.root
			if root == nil {
				root = fs
			} else {
				path = filepath.Join(fs.path, path)
			}
			ns = &BindataFileSystem{path: path, root: root, parent: fs, nameSpace: name,
				getAssetFunc: fs.getAssetFunc, getAssetInfoFunc: fs.getAssetInfoFunc, fileWalkFunc: fs.fileWalkFunc,
				fileWalkInfoFunc: fs.fileWalkInfoFunc, globFunc: fs.globFunc, globInfoFunc: fs.globInfoFunc}
			ns.init()
			fs.nameSpaces[nameSpace] = ns
		}
		fs = ns
	}

	return fs
}

func (fs *BindataFileSystem) GetNameSpace() string {
	return fs.nameSpace
}

func (fs *BindataFileSystem) GetParent() api.Interface {
	return fs.parent
}

func (fs *BindataFileSystem) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if fs.HttpHandler == nil {
		fs.HttpHandler = HTTPStaticHandler(fs)
	}
	fs.HttpHandler.ServeHTTP(w, r)
}

func (fs *BindataFileSystem) NewRepository(pkg string) rapi.Interface {
	repo := repository.NewRepository(pkg)
	repo.AddSourcePath(&path_helpers.Path{})
	return repo
}

func (fs *BindataFileSystem) RegisterPlugin(plugins ...api.Plugin) {
	for _, p := range plugins {
		p.Init(fs)
	}
	for _, p := range plugins {
		p.PathRegisterCallback(fs)
	}
}

func (fs *BindataFileSystem) DumpFiles(cb func(info api.FileInfo) error) error {
	return fs.dump(true, cb)
}

func (fs *BindataFileSystem) Dump(cb func(info api.FileInfo) error) error {
	return fs.dump(false, cb)
}

func (fs *BindataFileSystem) dump(onlyFiles bool, cb func(info api.FileInfo) error) error {
	mode := api.WalkAll
	if onlyFiles {
		mode ^= api.WalkDirs
	}
	return fs.WalkInfo(".", cb, mode)
}
