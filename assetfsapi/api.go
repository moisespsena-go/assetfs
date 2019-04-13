package assetfsapi

import (
	"context"
	"io"
	"net/http"
	"os"

	"github.com/moisespsena-go/io-common"

	"github.com/moisespsena-go/assetfs/repository/api"
)

type FileInfo interface {
	os.FileInfo
	RealPath() string
	Path() string
	Reader() (iocommon.ReadSeekCloser, error)
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
	AssetC(ctx context.Context, path string) (asset AssetInterface, err error)
	AssetOrPanic(path string) AssetInterface
	AssetOrPanicC(ctx context.Context, path string) AssetInterface
	AssetInfo(path string) (FileInfo, error)
	AssetInfoC(ctx context.Context, path string) (FileInfo, error)
	AssetInfoOrPanic(path string) FileInfo
	AssetInfoOrPanicC(ctx context.Context, path string) FileInfo
	AssetReader() AssetReaderFunc
	AssetReaderC() AssetReaderFuncC
	Provider(providers ...Interface)
	Providers() []Interface
}

type AssetCompilerInterface interface {
	Compile() error
}

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
	http.Handler
	GetNameSpace(nameSpace string) (NameSpacedInterface, error)
	NameSpaces() []NameSpacedInterface
	NameSpace(nameSpace string) NameSpacedInterface
	GetPath() string
	GetParent() Interface
	NewRepository(pkg string) api.Interface
	RegisterPlugin(plugins ...Plugin)
	DumpFiles(cb func(info FileInfo) error) error
	Dump(cb func(info FileInfo) error, ignore ...func(pth string) bool) error
}

type PathRegistrator interface {
	Interface
	OnPathRegister(cb ...PathRegisterCallback)
	PrependPath(path string, ignoreExists ...bool) error
	RegisterPath(path string, ignoreExists ...bool) error
}

type NameSpacedInterface interface {
	Interface
	GetName() string
}
