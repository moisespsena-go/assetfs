package assetfs

import (
	"context"
	"net/http"

	oscommon "github.com/moisespsena-go/os-common"

	"github.com/moisespsena-go/assetfs/assetfsapi"
	"github.com/moisespsena-go/assetfs/local"
)

// fakeFileSystem fake AssetFS
type fakeFileSystem struct {
	assetfsapi.AssetGetterInterface
	assetfsapi.TraversableInterface
	local.LocalSourcesAttribute
}

func (f *fakeFileSystem) GetName() string {
	return ""
}

func (f *fakeFileSystem) Compile() error {
	return nil
}

func (f *fakeFileSystem) ServeHTTP(http.ResponseWriter, *http.Request) {
}

func (f *fakeFileSystem) GetNameSpace(string) (assetfsapi.NameSpacedInterface, error) {
	return f, nil
}

func (f *fakeFileSystem) NameSpaces() []assetfsapi.NameSpacedInterface {
	return nil
}

func (f *fakeFileSystem) NameSpace(string) assetfsapi.NameSpacedInterface {
	return f
}

func (f *fakeFileSystem) GetPath() string {
	return "fake_fs:/"
}

func (f *fakeFileSystem) GetParent() assetfsapi.Interface {
	return nil
}

func (f *fakeFileSystem) RegisterPlugin(...assetfsapi.Plugin) {
}

func (f *fakeFileSystem) DumpFiles(func(assetfsapi.FileInfo) error) error {
	return nil
}

func (f *fakeFileSystem) Dump(func(assetfsapi.FileInfo) error, ...func(pth string) bool) error {
	return nil
}

func FakeFileSystem() Interface {
	fs := &fakeFileSystem{}
	fs.AssetGetterInterface = &AssetGetter{
		fs: fs,
		AssetFunc: func(ctx context.Context, name string) (data []byte, err error) {
			return nil, oscommon.ErrNotFound("fake_fs:" + name)
		},
		AssetInfoFunc: func(ctx context.Context, path string) (assetfsapi.FileInfo, error) {
			return nil, oscommon.ErrNotFound("fake_fs:" + path)
		},
	}
	fs.TraversableInterface = &Traversable{
		FS: fs,
		WalkFunc: func(dir string, cb assetfsapi.CbWalkFunc, mode assetfsapi.WalkMode) error {
			return nil
		},
		WalkInfoFunc: func(dir string, cb assetfsapi.CbWalkInfoFunc, mode assetfsapi.WalkMode) error {
			return nil
		},
		GlobFunc: func(pattern assetfsapi.GlobPattern, cb func(pth string, isDir bool) error) (err error) {
			return nil
		},
		GlobInfoFunc: func(pattern assetfsapi.GlobPattern, cb func(info assetfsapi.FileInfo) error) (err error) {
			return nil
		},
	}
	return fs
}
