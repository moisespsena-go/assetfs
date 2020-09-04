package assetfsapi

import (
	"errors"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	iocommon "github.com/moisespsena-go/io-common"
	oscommon "github.com/moisespsena-go/os-common"

	httpcommon "github.com/moisespsena-go/http-common"
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

	f = httpcommon.NewFile(asset, func() (rsc iocommon.ReadSeekCloser, err error) {
		var r io.ReadCloser
		if r, err = asset.Reader(); err != nil {
			return
		}
		var ok bool
		if rsc, ok = r.(iocommon.ReadSeekCloser); ok {
			return
		}
		return readCloseUnsupportSeeker{r}, nil
	})
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

type readCloseUnsupportSeeker struct {
	io.ReadCloser
}

func (this readCloseUnsupportSeeker) RawReader() io.ReadCloser {
	if rr, ok := this.ReadCloser.(RawReadGetter); ok {
		return rr.RawReader()
	}
	return this.ReadCloser
}

func (readCloseUnsupportSeeker) Seek(int64, int) (int64, error) {
	return 0, errors.New("seek is not supported")
}
