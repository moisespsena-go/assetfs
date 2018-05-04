package assetfs

import (
	"os"
	"fmt"
)

type AssetGetter struct {
	AssetFunc     func(path string) ([]byte, error)
	AssetInfoFunc func(path string) (os.FileInfo, error)
	GlobFunc      GlobFunc
}

func (f *AssetGetter) Asset(path string) (AssetInterface, error) {
	data, err := f.AssetFunc(path)
	if err != nil {
		return nil, &AssetError{path, err}
	}
	return NewAsset(path, data), nil
}

func (f *AssetGetter) AssetOrPanic(path string) AssetInterface {
	asset, err := f.Asset(path)
	if err != nil {
		panic(err)
	}
	return asset
}

func (f *AssetGetter) AssetInfo(path string) (os.FileInfo, error) {
	return f.AssetInfoFunc(path)
}

func (f *AssetGetter) AssetInfoOrPanic(path string) os.FileInfo {
	info, err := f.AssetInfo(path)
	if err != nil {
		panic(err)
	}
	return info
}

func (f *AssetGetter) Glob(pattern string, recursive ...bool) (matches []string, err error) {
	return f.GlobFunc(pattern, recursive...)
}

func (f *AssetGetter) GlobOrPanic(pattern string, recursive ...bool) []string {
	matches, err := f.GlobFunc(pattern, recursive...)
	if err != nil {
		panic(fmt.Errorf("Glob(%q) error: %v", pattern, err))
	}
	return matches
}

func (f *AssetGetter) AssetReader() AssetReaderFunc {
	return f.AssetFunc
}
