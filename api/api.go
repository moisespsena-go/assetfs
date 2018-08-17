package api

import (
	"os"
	"io"
	"net/http"
	"github.com/moisespsena/go-assetfs/repository/api"
)

const (
	FileTypeNameSpace FileType = 1 << iota
	FileTypeNormal
	FileTypeDir
	FileTypeReal
	FileTypeBindata
)

type FileType int

func (f FileType) IsNameSpace() bool {
	return (f & FileTypeNameSpace) != 0
}

func (f FileType) IsNormal() bool {
	return (f & FileTypeNormal) != 0
}

func (f FileType) IsDir() bool {
	return (f & FileTypeDir) != 0
}

func (f FileType) IsReal() bool {
	return (f & FileTypeReal) != 0
}

func (f FileType) IsBindata() bool {
	return (f & FileTypeBindata) != 0
}
const (
	WalkDirs WalkMode = 1 << iota
	WalkFiles
	WalkNameSpaces
	WalkNameSpacesLookUp
	WalkParentLookUp
	WalkReverse

	WalkAll = WalkFiles | WalkDirs | WalkNameSpaces | WalkNameSpacesLookUp | WalkParentLookUp
)

type WalkMode int

func (f WalkMode) IsDirs() bool {
	return (f & WalkDirs) != 0
}

func (f WalkMode) IsFiles() bool {
	return (f & WalkFiles) != 0
}

func (f WalkMode) IsNameSpaces() bool {
	return (f & WalkNameSpaces) != 0
}

func (f WalkMode) IsNameSpacesLookUp() bool {
	return (f & WalkNameSpacesLookUp) != 0
}

func (f WalkMode) IsParentLookUp() bool {
	return (f & WalkParentLookUp) != 0
}
func (f WalkMode) IsReverse() bool {
	return (f & WalkReverse) != 0
}


type FileInfo interface {
	os.FileInfo
	RealPath() string
	Path() string
	Reader() (io.ReadCloser, error)
	Writer() (io.WriteCloser, error)
	Appender() (io.WriteCloser, error)
	Data() ([]byte, error)
	Type() FileType
}



type DirFileInfo interface {
	FileInfo
	ReadDir(func(child FileInfo) error) error
}

type AssetInterface interface {
	GetName() string
	GetData() []byte
	GetString() string
	GetPath() string
}

type AssetGetterInterface interface {
	Asset(path string) (AssetInterface, error)
	AssetOrPanic(path string) AssetInterface
	AssetInfo(path string) (FileInfo, error)
	AssetInfoOrPanic(path string) FileInfo
	AssetReader() AssetReaderFunc
	Provider(providers ...Interface)
	Providers() []Interface
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
type CbWalkFunc = func(name string, isDir bool) error
type CbWalkInfoFunc = func(info FileInfo) error
type AssetReaderFunc = func(name string) (data []byte, err error)
type GetAssetInfoFunc = func(name string) (info FileInfo, err error)
type PathFormatterFunc = func(pth *string)

type WalkFunc = func(pth string, cb CbWalkFunc, mode WalkMode) error
type WalkInfoFunc = func(path string, cb CbWalkInfoFunc, mode WalkMode) error
type GlobFunc = func(pattern GlobPattern, cb func(pth string, isDir bool) error) error
type GlobInfoFunc = func(pattern GlobPattern, cb func(info FileInfo) error) error

type TraversableInterface interface {
	Walk(dir string, cb CbWalkFunc, mode ...WalkMode) error
	WalkInfo(dir string, cb CbWalkInfoFunc, mode ...WalkMode) error
	ReadDir(dir string, cb CbWalkInfoFunc, skipDir bool) (err error)
	Glob(pattern GlobPattern, cb func(pth string, isDir bool) error) error
	GlobInfo(pattern GlobPattern, cb func(info FileInfo) error) error
	NewGlob(pattern GlobPattern) Glob
	NewGlobString(pattern string) Glob
}

type Plugin interface {
	Init(fs Interface)
	PathRegisterCallback(fs Interface)
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
	NewRepository(pkg string) api.Interface
	RegisterPlugin(plugins ...Plugin)
	DumpFiles(cb func(info FileInfo) error) error
	Dump(cb func(info FileInfo) error) error
}

type NameSpacedInterface interface {
	Interface
	GetNameSpace() string
}
