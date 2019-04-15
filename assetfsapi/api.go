package assetfsapi

import (
	"context"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/moisespsena-go/io-common"
)

type BasicFileInfo interface {
	os.FileInfo
	Path() string
}

type BasicFileInfoWithChangedTime interface {
	BasicFileInfo
	ChangeTime() time.Time
}

type FileContents interface {
	Data() ([]byte, error)
	String() (string, error)
}

type RFileInfo interface {
	BasicFileInfo
	Reader() (iocommon.ReadSeekCloser, error)
	RealPath() string
}

type FileInfo interface {
	RFileInfo
	Writer() (io.WriteCloser, error)
	Appender() (io.WriteCloser, error)
	Type() FileType
}

type DirFileInfo interface {
	FileInfo
	ReadDir(func(child FileInfo) error) error
}

type AssetInterface interface {
	Name() string
	Data() ([]byte, error)
	DataS() (string, error)
	MustData() []byte
	MustDataS() string
	Path() string
}

type AssetGetterInterface interface {
	Asset(path string) (AssetInterface, error)
	AssetC(ctx context.Context, path string) (asset AssetInterface, err error)
	MustAsset(path string) AssetInterface
	MustAssetC(ctx context.Context, path string) AssetInterface
	AssetInfo(path string) (FileInfo, error)
	AssetInfoC(ctx context.Context, path string) (FileInfo, error)
	MustAssetInfo(path string) FileInfo
	MustAssetInfoC(ctx context.Context, path string) FileInfo
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

type LocalSourcesGetter interface {
	LocalSources() LocalSourceRegister
}

type LocalSourcesAttribute interface {
	LocalSourcesGetter
	SetLocalSources(sources LocalSourceRegister)
}

type Interface interface {
	AssetGetterInterface
	AssetCompilerInterface
	TraversableInterface
	LocalSourcesAttribute
	http.Handler
	GetNameSpace(nameSpace string) (NameSpacedInterface, error)
	NameSpaces() []NameSpacedInterface
	NameSpace(nameSpace string) NameSpacedInterface
	GetPath() string
	GetParent() Interface
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
