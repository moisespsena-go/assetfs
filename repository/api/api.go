package api

import (
	"io"
	"os"

	"github.com/moisespsena-go/file-utils"
	"github.com/moisespsena-go/sortvalues"
	"github.com/moisespsena-go/xbindata"
)

type PreSourceAddFunc = func(repo Interface, src fileutils.Copier) fileutils.Copier
type AfterSourceAddFunc = func(repo Interface, src fileutils.Copier)

type Template struct {
	Name string
	Data string
}

type Plugin interface {
	GetTemplates() []*Template
	Init(repo Interface)
}

type PrepareConfigCallbackValueInterface interface {
	sortvalues.ValueInterface
	Callback() func(config *xbindata.Config)
}

type Interface interface {
	AddSource(sources ...fileutils.Copier)
	AddSourcePath(sources ...*fileutils.Dir)
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
	PrepareConfig(f func(config *xbindata.Config), name ...string) PrepareConfigCallbackValueInterface
	IgnorePath(f ...func(pth string) bool)
}

type Dumper func(cb func(pth string, stat os.FileInfo, reader io.Reader) error, ignore ...func(pth string) bool) error
