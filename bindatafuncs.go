package assetfs

import (
	"os"
	"path/filepath"
)

// Glob list matched files from assetfs
func bindataGlob(fs *BindataFileSystem, glob GlobFunc, pattern string, recursive ...bool) (matches []string, err error) {
	if fs.root != nil {
		pattern = filepath.Join(fs.path, pattern)
		matches, err = glob(pattern, recursive...)
		if err != nil {
			return
		}
		l := len(fs.path)
		for i, mach := range matches {
			matches[i] = mach[l+1:]
		}
		return
	}
	return glob(pattern, recursive...)
}

func bindataAsset(fs *BindataFileSystem, asset AssetReaderFunc, pth string) ([]byte, error) {
	if fs.root != nil {
		pth = filepath.Join(fs.path, pth)
	}
	return asset(pth)
}

func bindataAssetInfo(fs *BindataFileSystem, assetInfo GetAssetInfoFunc, pth string) (os.FileInfo, error) {
	if fs.root != nil {
		pth = filepath.Join(fs.path, pth)
	}
	return assetInfo(pth)
}
