package assetfs

import (
	"path/filepath"
	"os"
	"net/http"
)

type BindataFileSystem struct {
	AssetGetterInterface
	TraversableInterface
	getAssetFunc     AssetReaderFunc
	getAssetInfoFunc GetAssetInfoFunc
	fileWalkFunc     FileWalkFunc
	fileWalkInfoFunc FileWalkInfoFunc
	globFunc         GlobFunc
	root             Interface
	path             string
	parent           Interface
	nameSpaces       map[string]NameSpacedInterface
	nameSpace        string
	callbacks        []PathRegisterCallback
	HttpHandler      http.Handler
}

func NewBindataFileSystem(getAsset AssetReaderFunc, getAssetInfo GetAssetInfoFunc, fileWalk FileWalkFunc,
	fileWalkInfo FileWalkInfoFunc, glob GlobFunc) (fs *BindataFileSystem) {
	fs = &BindataFileSystem{getAssetFunc: getAsset, getAssetInfoFunc: getAssetInfo, fileWalkFunc: fileWalk,
		fileWalkInfoFunc: fileWalkInfo, globFunc: glob}
	fs.init()
	return
}

func (fs *BindataFileSystem) init() {
	fs.AssetGetterInterface = &AssetGetter{
		AssetFunc: func(path string) ([]byte, error) {
			return bindataAsset(fs, fs.getAssetFunc, path)
		},
		AssetInfoFunc: func(path string) (os.FileInfo, error) {
			return bindataAssetInfo(fs, fs.getAssetInfoFunc, path)
		},
		GlobFunc: func(pattern string, recursive ...bool) (matches []string, err error) {
			return bindataGlob(fs, fs.globFunc, pattern, recursive...)
		}}
	fs.TraversableInterface = &Traversable{
		fs.AssetGetterInterface,
		func(dir string, cb WalkFunc, onError ...OnErrorFunc) error {
			return fs.fileWalkFunc(dir, cb)
		},
		func(dir string, cb WalkInfoFunc, onError ...OnErrorFunc) error {
			return fs.fileWalkInfoFunc(dir, cb)
		}}
}

func (fs *BindataFileSystem) OnPathRegister(cb ...PathRegisterCallback) {
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

func (fs *BindataFileSystem) Root() Interface {
	return fs.root
}

// Compile compile assetfs
func (fs *BindataFileSystem) Compile() error {
	return nil
}

// NameSpace return namespaced filesystem
func (fs *BindataFileSystem) NameSpace(nameSpace string) NameSpacedInterface {
	if fs.nameSpaces == nil {
		fs.nameSpaces = make(map[string]NameSpacedInterface)
	}
	if ns, ok := fs.nameSpaces[nameSpace]; ok {
		return ns
	}
	path := nameSpace
	root := fs.root
	if root == nil {
		root = fs
	} else {
		path = filepath.Join(fs.path, path)
	}
	ns := &BindataFileSystem{path: path, root: root, parent: fs, nameSpace: nameSpace,
		getAssetFunc: fs.getAssetFunc, getAssetInfoFunc: fs.getAssetInfoFunc, fileWalkFunc: fs.fileWalkFunc,
		fileWalkInfoFunc: fs.fileWalkInfoFunc, globFunc: fs.globFunc}
	ns.init()
	fs.nameSpaces[nameSpace] = ns
	return ns
}

func (fs *BindataFileSystem) GetNameSpace() string {
	return fs.nameSpace
}

func (fs *BindataFileSystem) GetParent() Interface {
	return fs.parent
}

func (fs *BindataFileSystem) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if fs.HttpHandler == nil {
		fs.HttpHandler = HTTPStaticHandler(fs)
	}
	fs.HttpHandler.ServeHTTP(w, r)
}
