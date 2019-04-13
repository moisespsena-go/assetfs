package assetfsapi

import (
	"os"
	"path"
	"time"

	"gopkg.in/djherbis/times.v1"
)

type basicFileInfo struct {
	path       string
	name       string
	size       int64
	mode       os.FileMode
	modTime    time.Time
	changeTime time.Time
}

func NewBasicFileInfo(pth string, size int64, mode os.FileMode, modTime, changeTime time.Time) BasicFileInfo {
	return &basicFileInfo{path: pth, name: path.Base(pth), size: size, mode: mode, modTime: modTime, changeTime: changeTime}
}

func NewCleanedBasicFileInfo(pth string, name ...string) BasicFileInfo {
	b := &basicFileInfo{path: pth}
	if len(name) == 0 || name[0] == "" {
		b.name = path.Base(pth)
	}
	return b
}

func OsFileInfoToBasic(pth string, info os.FileInfo) BasicFileInfo {
	b := &basicFileInfo{path: pth, name: path.Base(pth), size: info.Size(), mode: info.Mode(), modTime: info.ModTime()}
	t := times.Get(info)
	if t.HasChangeTime() {
		b.changeTime = t.ChangeTime()
	}
	return b
}

func (fi *basicFileInfo) Path() string {
	return fi.path
}

func (fi *basicFileInfo) Name() string {
	return fi.name
}

func (fi *basicFileInfo) Size() int64 {
	return fi.size
}

func (fi *basicFileInfo) Mode() os.FileMode {
	return fi.mode
}

func (fi *basicFileInfo) ModTime() time.Time {
	return fi.modTime
}

func (fi *basicFileInfo) ChangeTime() time.Time {
	return fi.changeTime
}

func (fi basicFileInfo) IsDir() bool {
	return false
}

func (fi basicFileInfo) Sys() interface{} {
	return nil
}

func SetBasicFileInfoPath(b BasicFileInfo, path string) BasicFileInfo {
	b.(*basicFileInfo).path = path
	return b
}
