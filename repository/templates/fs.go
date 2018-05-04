package templates

func FS() string {
	return `// +build !{{.BindataTag}}

package {{.Package}}

import (
	"github.com/moisespsena/go-assetfs"
)

var (
	FileSystem                   = assetfs.NewAssetFileSystem()
	AssetFS    assetfs.Interface = FileSystem
)
`
}

func FSBindata() string {
	return `// +build {{.BindataTag}}

package {{.Package}}

import (
	"C"
	"os"
	"fmt"
	"time"
	"strings"
	_ "unsafe"
	"path/filepath"
	"github.com/moisespsena/go-assetfs"

)

var (
	now = time.Now()
	AssetFS assetfs.Interface = assetfs.NewBindataFileSystem(Asset, AssetInfo, AssetWalk, AssetWalkInfo, AssetGlob)
)

type bindataDirInfo struct {
	name    string
}

func (fi bindataDirInfo) Name() string {
	return fi.name
}

func (fi bindataDirInfo) Size() int64 {
	return -1
}

func (fi bindataDirInfo) Mode() os.FileMode {
	return os.ModeDir
}

func (fi bindataDirInfo) ModTime() time.Time {
	return now
}

func (fi bindataDirInfo) IsDir() bool {
	return true
}

func (fi bindataDirInfo) Sys() interface{} {
	return nil
}

func AssetWalk(name string, cb assetfs.WalkFunc) (error) {
	node, err := AssetGetDir(name)
	if err != nil {
		return err
	}

	var walk func(name string, node *bintree) error
	walk = func(pth string, node *bintree) (err error) {
		for childName, child := range node.Children {
			if child.Func == nil {
				err = walk(filepath.Join(pth, childName), child)
			} else {
				err = cb("", filepath.Join(pth, childName), nil)
			}
			if err != nil {
				return err
			}
		}
		return
	}
	return walk(name, node)
}

func AssetWalkInfo(name string, cb assetfs.WalkInfoFunc) (error) {
	node, err := AssetGetDir(name)
	if err != nil {
		return err
	}

	var walk func(name string, node *bintree) error
	walk = func(pth string, node *bintree) (err error) {
		for childName, child := range node.Children {
			if child.Func == nil {
				info := &bindataDirInfo{childName}
				cpth := filepath.Join(pth, childName)
				err = cb("", cpth, info, nil)
				if err != nil {
					return
				}
				err = walk(cpth, child)
			} else {
				var asset *asset
				asset, err = child.Func()
				if err != nil {
					return
				}
				err = cb("", filepath.Join(pth, childName), asset.info, nil)
			}
			if err != nil {
				return err
			}
		}
		return
	}
	return walk(name, node)
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

func AssetGlob(pattern string, recursive ...bool) (matches []string, err error) {
	rec := len(recursive) > 0 && recursive[0]
	if !filepath_hasMeta(pattern) {
		if _, ok := _bindata[pattern]; !ok {
			_, err = AssetGetDir(pattern)
			if err != nil {
				return nil, nil
			}
		}
		return []string{pattern}, nil
	}

	dir, file := filepath.Split(pattern)
	dir = cleanGlobPath(dir)

	if !filepath_hasMeta(dir) {
		return glob(dir, file, nil, rec)
	}

	// Prevent infinite recursion. See issue 15879.
	if dir == pattern {
		return nil, filepath.ErrBadPattern
	}

	var m []string
	m, err = AssetGlob(dir)
	if err != nil {
		return
	}
	for _, d := range m {
		matches, err = glob(d, file, matches, rec)
		if err != nil {
			return
		}
	}
	return
}

func glob(dir, pattern string, matches []string, recursive bool) (m []string, e error) {
	m = matches
	if recursive {
		if dir == "" {
			e = AssetWalk(dir, func(prefix, name string, err error) error {
				matched, err := filepath.Match(pattern, filepath.Base(name))
				if err != nil {
					return err
				}
				if matched {
					m = append(m, name)
				}
				return nil
			})
		} else {
			e = AssetWalk(dir, func(prefix, name string, err error) error {
				if strings.HasPrefix(name, dir + "/") {
					matched, err := filepath.Match(pattern, filepath.Base(name))
					if err != nil {
						return err
					}
					if matched {
						m = append(m, name)
					}
				}
				return nil
			})
		}
	} else {
		e = AssetWalk(dir, func(prefix, name string, err error) error {
			if filepath.Dir(name) == dir {
				matched, err := filepath.Match(pattern, filepath.Base(name))
				if err != nil {
					return err
				}
				if matched {
					m = append(m, name)
				}
			}
			return nil
		})
	}
	return
}

//go:linkname filepath_cleanGlobPath path/filepath.cleanGlobPath
func filepath_cleanGlobPath(path string) string

//go:linkname filepath_hasMeta path/filepath.hasMeta
func filepath_hasMeta(path string) bool

func cleanGlobPath(path string) string {
	if path == "" {
		return ""
	}
	return filepath_cleanGlobPath(path)
}
`
}
