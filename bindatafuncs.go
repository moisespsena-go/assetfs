package assetfs

import (
	"path/filepath"
	"github.com/moisespsena/go-assetfs/api"
)

// Names list matched files from assetfs
func bindataGlob(fs *BindataFileSystem, glob api.GlobFunc, pattern api.GlobPattern, cb func(pth string, isDir bool) error) error {
	if fs.root != nil {
		l := len(fs.path)
		pattern = pattern.Wrap(fs.path)
		return glob(pattern, func(pth string, isDir bool) error {
			return cb(pth[l+1:], isDir)
		})
	}
	return glob(pattern, cb)
}
// Names list matched files from assetfs
func bindataGlobInfo(fs *BindataFileSystem, glob api.GlobInfoFunc, pattern api.GlobPattern, cb func(info api.FileInfo) error) error {
	if fs.root != nil {
		pattern = pattern.Wrap(fs.path)
		return glob(pattern, func(info api.FileInfo) error {
			if i, ok := info.(interface{TrimPrefix(prefix string)}); ok {
				i.TrimPrefix(fs.path)
			}
			return cb(info)
		})
	}
	return glob(pattern, cb)
}

func bindataAsset(fs *BindataFileSystem, asset api.AssetReaderFunc, pth string) ([]byte, error) {
	if fs.root != nil {
		pth = filepath.Join(fs.path, pth)
	}
	return asset(pth)
}

func bindataAssetInfo(fs *BindataFileSystem, assetInfo api.GetAssetInfoFunc, pth string) (api.FileInfo, error) {
	if fs.root != nil {
		pth = filepath.Join(fs.path, pth)
	}
	return assetInfo(pth)
}
