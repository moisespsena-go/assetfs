package assetfsapi

import "os"

type LocalSourceInfo interface {
	os.FileInfo
	Path() string
}

type LocalSource interface {
	Dir() string
	Get(name string) (info LocalSourceInfo, err error)
}

type LocalSourceRegister interface {
	Register(name string, src LocalSource)
	Get(name string) (src LocalSource)
}
