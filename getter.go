package assetfs

import (
	"github.com/moisespsena/go-assetfs/api"
	"github.com/moisespsena/go-error-wrap"
)

type AssetGetter struct {
	fs            Interface
	AssetFunc     func(path string) ([]byte, error)
	AssetInfoFunc func(path string) (api.FileInfo, error)
	providers     []Interface
}

func (f *AssetGetter) Provider(providers ...Interface) {
	f.providers = append(f.providers, providers...)
}

func (f *AssetGetter) Providers() []Interface {
	return f.providers
}

func (f *AssetGetter) Asset(path string) (asset api.AssetInterface, err error) {
	info, err := f.AssetInfo(path)
	if err != nil {
		if api.IsNotFound(err) {
			var err2 error
			for _, provider := range f.providers {
				if asset, err2 = provider.Asset(path); err2 != nil {
					if api.IsNotFound(err2) {
						continue
					}
					return nil, &api.AssetError{path, err2}
				}
			}
		}
		return nil, &api.AssetError{path, err}
	}
	data, err := info.Data()
	if err != nil {
		return nil, errwrap.Wrap(err, "Read data")
	}
	return &FileInfoAsset{info, path, data}, nil
}

func (f *AssetGetter) AssetOrPanic(path string) api.AssetInterface {
	asset, err := f.Asset(path)
	if err != nil {
		panic(err)
	}
	return asset
}

func (f *AssetGetter) AssetInfo(path string) (api.FileInfo, error) {
	return f.AssetInfoFunc(path)
}

func (f *AssetGetter) AssetInfoOrPanic(path string) api.FileInfo {
	info, err := f.AssetInfo(path)
	if err != nil {
		panic(err)
	}
	return info
}

func (f *AssetGetter) AssetReader() api.AssetReaderFunc {
	return f.AssetFunc
}
