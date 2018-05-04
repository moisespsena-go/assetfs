package assetfs

import (
	"os"
	"strings"
	"io/ioutil"
	"path/filepath"
)

// Glob list matched files from assetfs
func filesystemGlob(fs *AssetFileSystem, pattern string, recursive ...bool) (matches []string, err error) {
	rec := len(recursive) > 0 && recursive[0]
	set := make(map[string]bool)
	var name string
	if rec {
		err = fs.WalkFilesInfo("", func(prefix, pth string, info os.FileInfo, err error) error {
			if err != nil || !info.IsDir() {
				return err
			}
			if results, err := filepath.Glob(filepath.Join(pth, pattern)); err == nil {
				for _, result := range results {
					name = strings.TrimPrefix(result, prefix)
					if _, ok := set[name]; !ok {
						matches = append(matches, name)
						set[name] = true
					}
				}
			}
			return nil
		})
		if err != nil {
			return
		}
	} else {
		for _, pth := range fs.paths {
			prefix := pth + string(filepath.Separator)
			if results, err := filepath.Glob(filepath.Join(pth, pattern)); err == nil {
				for _, result := range results {
					name = strings.TrimPrefix(result, prefix)
					if _, ok := set[name]; !ok {
						matches = append(matches, name)
						set[name] = true
					}
				}
			}
		}
	}

	if err == nil && fs.parent != nil {
		supers, err := filesystemGlob(fs.parent.(*AssetFileSystem), filepath.Join(fs.nameSpace, pattern), rec)
		if err != nil {
			return []string{}, err
		}

		prefix := fs.nameSpace + string(os.PathSeparator)
		for _, m := range supers {
			name = strings.TrimPrefix(m, prefix)
			if _, ok := set[name]; !ok {
				matches = append(matches, name)
				set[name] = true
			}
		}
	}

	return
}

// Asset get content with name from assetfs
func filesystemAsset(fs *AssetFileSystem, name string) (AssetInterface, error) {
	var path string
	for _, pth := range fs.paths {
		path = filepath.Join(pth, name)
		if _, err := os.Stat(path); err == nil {
			data, err := ioutil.ReadFile(path)
			if err != nil {
				return nil, err
			}
			return NewAsset(path, data), nil
		}
	}
	if fs.parent != nil {
		return filesystemAsset(fs.parent.(*AssetFileSystem), filepath.Join(fs.nameSpace, name))
	}
	return nil, NotFound(name)
}

func filesystemAssetInfo(fs *AssetFileSystem, path string) (info os.FileInfo, err error) {
	for _, pth := range fs.paths {
		info, err = os.Stat(filepath.Join(pth, path))
		if err == nil {
			return info, nil
		}
		if !os.IsNotExist(err) {
			return nil, err
		}
	}
	if fs.parent != nil {
		return filesystemAssetInfo(fs.parent.(*AssetFileSystem), filepath.Join(fs.nameSpace, path))
	}
	return nil, NotFound(path)
}
