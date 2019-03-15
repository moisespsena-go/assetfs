package assetfsapi

import (
	"errors"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/moisespsena-go/os-common"

	"github.com/moisespsena-go/http-common"
)

type HttpFileSystem struct {
	FS       Interface
	DirIndex bool
}

func NewHttpFileSystem(fs Interface) *HttpFileSystem {
	return &HttpFileSystem{FS: fs}
}

func (fs *HttpFileSystem) Open(name string) (f http.File, err error) {
	if filepath.Separator != '/' && strings.ContainsRune(name, filepath.Separator) {
		return nil, errors.New("http: invalid character in file path")
	}
	fullName := filepath.FromSlash(path.Clean("/" + name))
	if fullName[0] == filepath.Separator {
		fullName = fullName[1:]
	}

	if fullName == "." {
		return httpcommon.NewDir(fullName, httpAssetDirReader(fs.FS, fullName)), nil
	}

	var asset FileInfo
	if asset, err = fs.FS.AssetInfo(fullName); err != nil {
		if oscommon.IsNotFound(err) {
			err = os.ErrNotExist
		}
		return nil, err
	}
	if asset.IsDir() {
		return httpcommon.NewDir(fullName, httpAssetDirReader(fs.FS, fullName)), nil
	}
	f = httpcommon.NewFile(asset, asset.Reader)
	return
}

func httpAssetDirReader(fs Interface, path string) func(count int) (items []os.FileInfo, err error) {
	return func(count int) (items []os.FileInfo, err error) {
		if count > 0 {
			var i int
			err = fs.ReadDir(path, func(info FileInfo) error {
				if i == count {
					return io.EOF
				}
				items = append(items, info)
				i++
				return nil
			}, false)
		} else {
			err = fs.ReadDir(path, func(info FileInfo) error {
				items = append(items, info)
				return nil
			}, false)
		}

		if err == io.EOF {
			err = nil
		}

		if err != nil {
			return nil, err
		}
		return
	}
}
