package templates

func FS() string {
	return `// +build !{{.BindataTag}}

package {{.Package}}

import (
	"github.com/moisespsena-go/assetfs"
	"github.com/moisespsena-go/assetfs/assetfsapi"
)

var (
	FileSystem               = assetfs.NewAssetFileSystem()
	AssetFS    assetfsapi.Interface = FileSystem
)
`
}

func FSBindata() string {
	return `// +build {{.BindataTag}}

package {{.Package}}

import (
	"os"
	"fmt"
	"time"
	"strings"
	"path/filepath"
	"github.com/moisespsena-go/assetfs"
	"github.com/moisespsena-go/assetfs/assetfsapi"
)

var (
	now                   = time.Now()
	AssetFS assetfsapi.Interface = assetfs.NewBindataFileSystem(Asset, GetAssetInfo, AssetWalk, AssetWalkInfo, AssetGlob, AssetGlobInfo)
)

func AssetWalk(root string, cb assetfsapi.CbWalkFunc) (error) {
	node, err := AssetGetDir(root)
	if err != nil {
		return err
	}

	var walk func(name string, node *bintree) error
	walk = func(pth string, node *bintree) (err error) {
		for childName, child := range node.Children {
			if child.Func == nil {
				err = cb(filepath.Join(pth, childName), true)
				if err != nil {
					return err
				}
				err = walk(filepath.Join(pth, childName), child)
			} else {
				err = cb(filepath.Join(pth, childName), false)
			}
			if err != nil {
				return err
			}
		}
		return
	}
	return walk(root, node)
}

func newAssetFileInfo(assetFunc func() (*asset, error), pth string) assetfsapi.FileInfo {
	info := bindataFileInfo{name: filepath.Base(pth), size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	return assetfs.NewBindataFileInfo(info, pth, func() ([]byte, error) {
		asset, err := assetFunc()
		if err != nil {
			return nil, err
		}
		return asset.bytes, nil
	})
}

func assetFileInfo(pth string) assetfsapi.FileInfo {
	if assetFunc, ok := _bindata[pth]; ok {
		return newAssetFileInfo(assetFunc, pth)
	}
	return nil
}

func GetAssetInfo(pth string) (assetfsapi.FileInfo, error) {
	return getAssetInfo(pth, nil)
}

func getAssetInfo(pth string, node *bintree) (info assetfsapi.FileInfo, err error) {
	if node == nil {
		info := assetFileInfo(pth)
		if info != nil {
			return info, nil
		}
		node, err = AssetGetDir(pth)
		if err != nil {
			return nil, err
		}
	} else if node.Func != nil {
		info := assetFileInfo(pth)
		if info == nil {
			return nil, assetfsapi.NotFound(pth)
		}
		return info, nil
	}
	return assetfs.NewBindataDirInfo(pth, childrenInfo(node, pth)), nil
}

func childrenInfo(node *bintree, pth string) func(cb func(info assetfsapi.FileInfo) error) error {
	return func(cb func(info assetfsapi.FileInfo) error) error {
		var pth string
		for childName, child := range node.Children {
			info, err := getAssetInfo(filepath.Join(pth, childName), child)
			if err != nil {
				return err
			}
			err = cb(info)
			if err != nil {
				return err
			}
		}
		return nil
	}
}

func AssetWalkInfo(root string, cb assetfsapi.CbWalkInfoFunc) (error) {
	node, err := AssetGetDir(root)
	if err != nil {
		return err
	}

	var walk func(name string, node *bintree) error
	walk = func(pth string, node *bintree) (err error) {
		for childName, child := range node.Children {
			cpth := filepath.Join(pth, childName)
			var info assetfsapi.FileInfo
			if child.Func == nil {
				info = assetfs.NewBindataDirInfo(cpth, childrenInfo(child, cpth))
				err = cb(info)
				if err != nil {
					return
				}
				err = walk(cpth, child)
			} else {
				info = newAssetFileInfo(child.Func, cpth)
				err = cb(info)
			}
			if err != nil {
				return err
			}
		}
		return
	}
	return walk(root, node)
}

func AssetGetDir(name string) (*bintree, error) {
	node := _bintree
	if len(name) != 0 {
		cannonicalName := strings.Replace(name, "\\", "/", -1)
		pathList := strings.Split(cannonicalName, "/")
		for _, p := range pathList {
			node = node.Children[p]
			if node == nil {
				return nil, fmt.Errorf("Asset %s not found", name)
			}
		}
	}

	if node.Func != nil {
		return nil, fmt.Errorf("Asset %s not found", name)
	}
	return node, nil
}

func AssetGlob(pattern assetfsapi.GlobPattern, cb func(pth string, isDir bool) error) (err error) {
	if pattern.IsRecursive() {
		err = AssetWalk(pattern.Dir(), func(pth string, isDir bool) error {
			if isDir {
				if !pattern.AllowDirs() {
					return err
				}
			} else {
				if !pattern.AllowFiles() {
					return err
				}
			}
			if pattern.Match(filepath.Base(pth)) {
				return cb(pth, false)
			}
			return nil
		})
	} else {
		node, err := AssetGetDir(pattern.Dir())
		if err != nil {
			return err
		}
		info := assetfs.NewBindataDirInfo(pattern.Dir(), childrenInfo(node, pattern.Dir()))
		err = info.ReadDir(func(info assetfsapi.FileInfo) error {
			if info.IsDir() {
				if !pattern.AllowDirs() {
					return err
				}
			} else {
				if !pattern.AllowFiles() {
					return err
				}
			}
			if pattern.Match(info.Name()) {
				return cb(info.Path(), info.IsDir())
			}
			return nil
		})
	}
	return
}

func AssetGlobInfo(pattern assetfsapi.GlobPattern, cb func(info assetfsapi.FileInfo) error) (err error) {
	if pattern.IsRecursive() {
		err = AssetWalkInfo(pattern.Dir(), func(info assetfsapi.FileInfo) error {
			if info.IsDir() {
				if !pattern.AllowDirs() {
					return err
				}
			} else {
				if !pattern.AllowFiles() {
					return err
				}
			}
			if pattern.Match(info.Name()) {
				return cb(info)
			}
			return nil
		})
	} else {
		node, err := AssetGetDir(pattern.Dir())
		if err != nil {
			return err
		}
		info := assetfs.NewBindataDirInfo(pattern.Dir(), childrenInfo(node, pattern.Dir()))
		err = info.ReadDir(func(info assetfsapi.FileInfo) error {
			if info.IsDir() {
				if !pattern.AllowDirs() {
					return err
				}
			} else {
				if !pattern.AllowFiles() {
					return err
				}
			}
			if pattern.Match(info.Name()) {
				return cb(info)
			}
			return nil
		})
	}
	return
}`
}
