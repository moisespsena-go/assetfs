package assetfs

import (
	"os"
	"net/http"
)

type AssetInterface interface {
	GetData() []byte
	GetString() string
}

type AssetGetterInterface interface {
	Asset(path string) (AssetInterface, error)
	AssetOrPanic(path string) AssetInterface
	AssetInfo(path string) (os.FileInfo, error)
	AssetInfoOrPanic(path string) os.FileInfo
	Glob(pattern string, recursive ...bool) (matches []string, err error)
	GlobOrPanic(pattern string, recursive ...bool) []string
	AssetReader() AssetReaderFunc
}

type AssetCompilerInterface interface {
	Compile() error
}

type AssetRegisterInterface interface {
	OnPathRegister(cb ...PathRegisterCallback)
	PrependPath(path string) error
	RegisterPath(path string) error
}

type PathRegisterCallback = func(fs Interface)
type WalkFunc = func(prefix, name string, err error) error
type WalkInfoFunc = func(prefix, path string, info os.FileInfo, err error) error
type OnErrorFunc = func(path string, err error) (stop bool)
type AssetReaderFunc = func(name string) (data []byte, err error)
type GetAssetInfoFunc = func(name string) (info os.FileInfo, err error)
type FileWalkFunc = func(pth string, walkFunc WalkFunc) error
type FileWalkInfoFunc = func(path string, walkFunc WalkInfoFunc) error
type GlobFunc = func(pattern string, recursive ...bool) (matches []string, err error)

type TraversableInterface interface {
	WalkFiles(dir string, cb WalkFunc, onError ...OnErrorFunc) error
	WalkFilesInfo(dir string, cb WalkInfoFunc, onError ...OnErrorFunc) error
}

type Interface interface {
	AssetGetterInterface
	AssetCompilerInterface
	TraversableInterface
	AssetRegisterInterface
	http.Handler
	NameSpace(nameSpace string) NameSpacedInterface
	GetPath() string
	GetParent() Interface
}

type NameSpacedInterface interface {
	Interface
	GetNameSpace() string
}

type Asset struct {
	Name string
	Data []byte
}

func NewAsset(name string, data []byte) AssetInterface {
	return &Asset{name, data}
}

func (a *Asset) GetData() []byte {
	return a.Data
}

func (a *Asset) GetString() string {
	return string(a.Data)
}
