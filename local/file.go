package local

import (
	"crypto/sha256"
	"io"
	"io/ioutil"
	"os"
	"time"

	api "github.com/moisespsena-go/assetfs/assetfsapi"

	"fmt"

	"github.com/moisespsena-go/io-common"
)

type FileInfo = api.BasicFileInfo

type File struct {
	FileInfo
	realPath string
	reader   func() (iocommon.ReadSeekCloser, error)
	digest   *[sha256.Size]byte
}

func (f File) GetFileInfo() os.FileInfo {
	return f.FileInfo
}

func (f *File) Writer() (io.WriteCloser, error) {
	return os.Open(f.realPath)
}

func (f *File) Appender() (io.WriteCloser, error) {
	return os.OpenFile(f.realPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, f.Mode())
}

func (f *File) Type() api.FileType {
	return api.FileTypeReal
}

func (f *File) RealPath() string {
	return f.realPath
}

func NewFile(fileInfo FileInfo, reader func() (iocommon.ReadSeekCloser, error), digest *[sha256.Size]byte) *File {
	return &File{FileInfo: fileInfo, reader: reader, digest: digest}
}

func (f *File) ImportLocal(localPath, name string, info os.FileInfo) (err error) {
	if f.digest, err = Digest(localPath); err != nil {
		return
	}
	if info == nil {
		if info, err = os.Stat(localPath); err != nil {
			return
		}
	}
	f.FileInfo = api.NewBasicFileInfo(name, info.Size(), info.Mode(), info.ModTime(), time.Time{})
	f.realPath = localPath
	return
}

func (f *File) Reader() (iocommon.ReadSeekCloser, error) {
	if f.reader == nil {
		if i, ok := f.FileInfo.(api.RFileInfo); ok {
			return i.Reader()
		}
		return os.Open(f.realPath)
	}
	return f.reader()
}

func (f *File) Digest() (d [sha256.Size]byte) {
	if f.digest == nil {
		return
	}
	return *f.digest
}

func (f *File) Data() ([]byte, error) {
	if r, err := f.Reader(); err != nil {
		return nil, err
	} else {
		defer r.Close()
		return ioutil.ReadAll(r)
	}
}

func (f *File) DataS() (string, error) {
	if b, err := f.Data(); err != nil {
		return "", err
	} else {
		return string(b), nil
	}
}

func (f *File) MustData() []byte {
	if b, err := f.Data(); err != nil {
		panic(fmt.Errorf("[local file %q] MustaData: %v", f.realPath, err))
	} else {
		return b
	}
}

func (f *File) MustDataS() string {
	return string(f.MustData())
}
