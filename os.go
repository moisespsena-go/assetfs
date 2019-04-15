package assetfs

import (
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/moisespsena-go/io-common"

	"github.com/go-errors/errors"
	"github.com/moisespsena-go/assetfs/assetfsapi"
)

var (
	now          = time.Now()
	IS_DIR_ERROR = errors.New("Is directory.")
	IS_NS_ERROR  = errors.New("Is name space.")
)

type RealFileInfo struct {
	assetfsapi.BasicFileInfo
	realPath string
}

func NewRealFileInfo(basicFileInfo assetfsapi.BasicFileInfo, realPath string) *RealFileInfo {
	return &RealFileInfo{BasicFileInfo: basicFileInfo, realPath: realPath}
}

func (RealFileInfo) Type() assetfsapi.FileType {
	return assetfsapi.FileTypeReal | assetfsapi.FileTypeNormal
}

func (rf *RealFileInfo) RealPath() string {
	return rf.realPath
}

func (rf *RealFileInfo) Reader() (iocommon.ReadSeekCloser, error) {
	return os.Open(rf.realPath)
}

func (rf *RealFileInfo) Writer() (io.WriteCloser, error) {
	return os.OpenFile(rf.realPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, rf.Mode())
}

func (rf *RealFileInfo) Appender() (io.WriteCloser, error) {
	return os.OpenFile(rf.realPath, os.O_APPEND|os.O_WRONLY, rf.Mode())
}

func (rf *RealFileInfo) String() string {
	return StringifyFileInfo(rf)
}

func (rf *RealFileInfo) Data() ([]byte, error) {
	reader, err := rf.Reader()
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	return ioutil.ReadAll(reader)
}

func (rf *RealFileInfo) DataS() (string, error) {
	b, err := rf.Data()
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (rf *RealFileInfo) MustData() []byte {
	if b, err := rf.Data(); err != nil {
		panic(err)
	} else {
		return b
	}
}

func (rf *RealFileInfo) MustDataS() string {
	if b, err := rf.Data(); err != nil {
		panic(err)
	} else {
		return string(b)
	}
}

type RealDirFileInfo struct {
	*RealFileInfo
}

func (RealDirFileInfo) Type() assetfsapi.FileType {
	return assetfsapi.FileTypeReal | assetfsapi.FileTypeDir
}

func (rf *RealDirFileInfo) Reader() (iocommon.ReadSeekCloser, error) {
	return nil, IS_DIR_ERROR
}

func (rf *RealDirFileInfo) Writer() (io.WriteCloser, error) {
	return nil, IS_DIR_ERROR
}

func (rf *RealDirFileInfo) Appender() (io.WriteCloser, error) {
	return nil, IS_DIR_ERROR
}

func (rf *RealDirFileInfo) Data() ([]byte, error) {
	return nil, IS_DIR_ERROR
}

func (rf *RealDirFileInfo) String() string {
	return StringifyFileInfo(rf)
}

func (d *RealDirFileInfo) ReadDir(cb func(child assetfsapi.FileInfo) error) (err error) {
	infos, err := ioutil.ReadDir(d.realPath)
	if err != nil {
		return err
	}

	for _, info := range infos {
		rinfo := &RealFileInfo{basicFileInfo(path.Join(d.Path(), info.Name()), info), filepath.Join(d.realPath, info.Name())}
		if info.IsDir() {
			err = cb(&RealDirFileInfo{rinfo})
		} else {
			err = cb(rinfo)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

type NameSpaceFileInfo struct {
	assetfsapi.BasicFileInfo
	ns assetfsapi.Interface
}

func (NameSpaceFileInfo) Type() assetfsapi.FileType {
	return assetfsapi.FileTypeNameSpace
}

func (ns *NameSpaceFileInfo) Size() int64 {
	return -1
}
func (ns *NameSpaceFileInfo) Mode() os.FileMode {
	return os.ModeDir
}
func (ns *NameSpaceFileInfo) ModTime() time.Time {
	return now
}
func (ns *NameSpaceFileInfo) IsDir() bool {
	return true
}
func (ns *NameSpaceFileInfo) Sys() interface{} {
	return nil
}

func (ns *NameSpaceFileInfo) Reader() (iocommon.ReadSeekCloser, error) {
	return nil, IS_NS_ERROR
}
func (ns *NameSpaceFileInfo) Writer() (io.WriteCloser, error) {
	return nil, IS_NS_ERROR
}

func (ns *NameSpaceFileInfo) Appender() (io.WriteCloser, error) {
	return nil, IS_NS_ERROR
}

func (ns *NameSpaceFileInfo) Data() ([]byte, error) {
	return nil, IS_NS_ERROR
}

func (ns *NameSpaceFileInfo) ReadDir(cb func(child assetfsapi.FileInfo) error) (err error) {
	return ns.ns.ReadDir(".", cb, false)
}
func (ns *NameSpaceFileInfo) RealPath() string {
	return ns.ns.GetPath()
}

func (ns *NameSpaceFileInfo) String() string {
	return StringifyFileInfo(ns)
}

func StringifyFileInfo(info assetfsapi.FileInfo) string {
	b := []byte("oo")
	typ := info.Type()
	if typ.IsDir() {
		b[0] = 'd'
	} else if typ.IsNormal() {
		b[0] = 'f'
	} else if typ.IsNameSpace() {
		b[0] = 'n'
	}
	if typ.IsBindata() {
		b[1] = 'b'
	} else if typ.IsReal() {
		b[1] = 'r'
	}
	return string(b) + "://" + info.Path()
}

func ParseFileType(typ string) (t assetfsapi.FileType) {
	if typ[2:5] == "://" {
		switch typ[0] {
		case 'd':
			t |= assetfsapi.FileTypeDir
		case 'f':
			t |= assetfsapi.FileTypeNormal
		case 'n':
			t |= assetfsapi.FileTypeNameSpace
		}
		switch typ[1] {
		case 'b':
			t |= assetfsapi.FileTypeBindata
		case 'r':
			t |= assetfsapi.FileTypeReal
		}
	}
	return t
}
