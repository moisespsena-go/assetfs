package assetfs

import (
	"context"

	"github.com/moisespsena-go/assetfs/assetfsapi"
	"github.com/moisespsena-go/os-common"
)

type AssetGetter struct {
	fs            Interface
	AssetFunc     func(ctx context.Context, path string) ([]byte, error)
	AssetInfoFunc func(ctx context.Context, path string) (assetfsapi.FileInfo, error)
	providers     []Interface
}

func (f *AssetGetter) Provider(providers ...Interface) {
	f.providers = append(f.providers, providers...)
}

func (f *AssetGetter) Providers() []Interface {
	return f.providers
}

func (f *AssetGetter) AssetC(ctx context.Context, path string) (asset assetfsapi.AssetInterface, err error) {
	info, err := f.AssetInfoC(ctx, path)
	if err != nil {
		if oscommon.IsNotFound(err) {
			var err2 error
			for _, provider := range f.providers {
				if asset, err2 = provider.AssetC(ctx, path); err2 != nil {
					if oscommon.IsNotFound(err2) {
						continue
					}
					return nil, &oscommon.PathError{path, err2}
				}
			}
		}
		return nil, &oscommon.PathError{path, err}
	}

	return info.(assetfsapi.AssetInterface), nil
}

func (f *AssetGetter) Asset(path string) (asset assetfsapi.AssetInterface, err error) {
	return f.AssetC(nil, path)
}

func (f *AssetGetter) MustAssetC(ctx context.Context, path string) assetfsapi.AssetInterface {
	asset, err := f.AssetC(ctx, path)
	if err != nil {
		panic(err)
	}
	return asset
}

func (f *AssetGetter) MustAsset(path string) assetfsapi.AssetInterface {
	asset, err := f.Asset(path)
	if err != nil {
		panic(err)
	}
	return asset
}

func (f *AssetGetter) AssetInfoC(ctx context.Context, path string) (assetfsapi.FileInfo, error) {
	return f.AssetInfoFunc(ctx, path)
}

func (f *AssetGetter) AssetInfo(path string) (assetfsapi.FileInfo, error) {
	return f.AssetInfoFunc(nil, path)
}

func (f *AssetGetter) MustAssetInfo(path string) assetfsapi.FileInfo {
	info, err := f.AssetInfo(path)
	if err != nil {
		panic(err)
	}
	return info
}

func (f *AssetGetter) MustAssetInfoC(ctx context.Context, path string) assetfsapi.FileInfo {
	info, err := f.AssetInfoC(ctx, path)
	if err != nil {
		panic(err)
	}
	return info
}

func (f *AssetGetter) AssetReaderC() assetfsapi.AssetReaderFuncC {
	return f.AssetFunc
}

func (f *AssetGetter) AssetReader() assetfsapi.AssetReaderFunc {
	return func(name string) (data []byte, err error) {
		return f.AssetReaderC()(nil, name)
	}
}
