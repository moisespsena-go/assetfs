package api

import (
	"github.com/moisespsena/go-path-helpers"
	"os"
	"io"
)


type PreSourceAddFunc = func(repo Interface, src interface{}) interface{}
type AfterSourceAddFunc = func(repo Interface, src interface{})

type Template struct {
	Name string
	Data string
}

type Plugin interface {
	GetTemplates() []*Template
	Init(repo Interface)
}

type Interface interface {
	AddSource(sources ...interface{})
	AddSourcePath(sources ...*path_helpers.Path)
	BinFile() string
	AbsPath(create ...bool) string
	DataDir(create ...bool) string
	Init()
	Sync()
	Clean()
	RegisterPlugin(plugins ...Plugin)
	PreSourceAdd(cbcs ...PreSourceAddFunc)
	AfterSourceAdd(cbcs ...AfterSourceAddFunc)
	PreSync(cbcs ...func(repo Interface))
	AfterSync(cbcs ...func(repo Interface))
	PreClean(cbcs ...func(repo Interface))
	AfterClean(cbcs ...func(repo Interface))
	Dumper(dumpers ...Dumper)
}

type Dumper func(cb func(pth string, stat os.FileInfo, reader io.Reader) error) error
