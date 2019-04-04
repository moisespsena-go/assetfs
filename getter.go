package assetfs

import (
	"github.com/moisespsena-go/os-common"
	"github.com/moisespsena-go/assetfs/assetfsapi"
	"github.com/moisespsena-go/error-wrap"
)

type AssetGetter struct {
	fs            Interface
	AssetFunc     func(path string) ([]byte, error)
	AssetInfoFunc func(path string) (assetfsapi.FileInfo, error)
	providers     []Interface
}

func (f *AssetGetter) Provider(providers ...Interface) {
	f.providers = append(f.providers, providers...)
}

func (f *AssetGetter) Providers() []Interface {
	return f.providers
}

func (f *AssetGetter) Asset(path string) (asset assetfsapi.AssetInterface, err error) {
	info, err := f.AssetInfo(path)
	if err != nil {
		if oscommon.IsNotFound(err) {
			var err2 error
			for _, provider := range f.providers {
				if asset, err2 = provider.Asset(path); err2 != nil {
					if oscommon.IsNotFound(err2) {
						continue
					}
					return nil, &oscommon.PathError{path, err2}
				}
			}
		}
		return nil, &oscommon.PathError{path, err}
	}
	data, err := info.Data()
	if err != nil {
		return nil, errwrap.Wrap(err, "Read data")
	}
	return &FileInfoAsset{info, path, data}, nil
}

func (f *AssetGetter) AssetOrPanic(path string) assetfsapi.AssetInterface {
	asset, err := f.Asset(path)
	if err != nil {
		panic(err)
	}
	return asset
}

func (f *AssetGetter) AssetInfo(path string) (assetfsapi.FileInfo, error) {
	return f.AssetInfoFunc(path)
}

func (f *AssetGetter) AssetInfoOrPanic(path string) assetfsapi.FileInfo {
	info, err := f.AssetInfo(path)
	if err != nil {
		panic(err)
	}
	return info
}

func (f *AssetGetter) AssetReader() assetfsapi.AssetReaderFunc {
	return f.AssetFunc
}
