package repository

import (
	"os"
	"fmt"
	"bytes"
	"path/filepath"
	"text/template"
	"github.com/jteeuwen/go-bindata"
	"github.com/moisespsena/go-path-helpers"
	"github.com/moisespsena/go-assetfs/repository/templates"
)

type PreSourceAddFunc = func(repo Interface, src interface{}) interface{}
type AfterSourceAddFunc = func(repo Interface, src interface{})

type PluginInterface interface {
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
	RegisterPlugin(plugins ...PluginInterface)
	PreSourceAdd(cbcs ...PreSourceAddFunc)
	AfterSourceAdd(cbcs ...AfterSourceAddFunc)
	PreSync(cbcs ...func(repo Interface))
	AfterSync(cbcs ...func(repo Interface))
	PreClean(cbcs ...func(repo Interface))
	AfterClean(cbcs ...func(repo Interface))
}

type Repository struct {
	PackagePath          string
	Package              string
	PreCompileTag        string
	CleanTag             string
	BindataTag           string
	Templates            []*Template
	PrepareConfig        func(config *bindata.Config)
	Sources              []interface{}
	sources              map[interface{}]int
	absPath              string
	Plugins              []PluginInterface
	preSourceAdd         []PreSourceAddFunc
	afterSourceAdd       []AfterSourceAddFunc
	preClean             []func(repo Interface)
	afterClean           []func(repo Interface)
	preSync              []func(repo Interface)
	afterSync            []func(repo Interface)
}

type Template struct {
	Name string
	Data string
}

func NewRepository(packagePath string) *Repository {
	return &Repository{PackagePath: packagePath, Package: filepath.Base(packagePath),
		PreCompileTag: "pre_compile", BindataTag: "bindata", CleanTag: "bindata_clean"}
}

func (r *Repository) RegisterPlugin(plugins ...PluginInterface) {
	r.Plugins = append(r.Plugins, plugins...)
}

func (r *Repository) AddSourcePath(sources ...*path_helpers.Path) {
	for _, p := range sources {
		r.AddSource(p)
	}
}

func (r *Repository) AddSource(sources ...interface{}) {
	if r.sources == nil {
		r.sources = make(map[interface{}]int)
	}
	for _, s := range sources {
		ss := fmt.Sprint(s)
		if _, ok := r.sources[ss]; !ok {
			for _, p := range r.preSourceAdd {
				s = p(r, s)
			}
			r.sources[ss] = len(r.Sources)
			r.Sources = append(r.Sources, s)
			for _, p := range r.afterSourceAdd {
				p(r, s)
			}
		}
	}
}

func (r *Repository) BinFile() string {
	return filepath.Join(r.AbsPath(), "data.go")
}

func (r *Repository) AbsPath(create ...bool) string {
	if r.absPath == "" {
		if absPath := path_helpers.ResolveGoSrcPath(filepath.Dir(r.PackagePath)); absPath != "" {
			absPath = filepath.Join(absPath, filepath.Base(r.PackagePath))
			if (len(create) != 0 && create[0]) && !path_helpers.IsExistingDir(absPath) {
				if err := os.MkdirAll(absPath, os.ModePerm); err != nil {
					panic(err)
				}
			}
			r.absPath = absPath
		}
	}
	return r.absPath
}

func (r *Repository) DataDir(create ...bool) string {
	absPath := filepath.Join(r.AbsPath(create...), "data")
	if (len(create) != 0 && create[0]) && !path_helpers.IsExistingDir(absPath) {
		if err := os.MkdirAll(absPath, os.ModePerm); err != nil {
			panic(err)
		}
	}
	return absPath
}

func (r *Repository) renderTemplate(tpl string) []byte {
	t, err := template.New("-").Parse(tpl)
	if err != nil {
		panic(err)
	}
	var out bytes.Buffer
	err = t.Execute(&out, r)
	if err != nil {
		panic(err)
	}
	return out.Bytes()
}

func (r *Repository) template(tpl *Template) interface{} {
	return &path_helpers.Path{Alias: tpl.Name, Data: r.renderTemplate(tpl.Data)}
}

func (r *Repository) GetInitTempĺates() []*Template {
	tpls := r.Templates[:]
	tpls = append(tpls,
		&Template{"repository.go", templates.Repository()},
		&Template{"assetfsfs.go", templates.FS()},
		&Template{"assetfsfs_bindata.go", templates.FSBindata()},
		&Template{"assetfsprecompile.go", templates.PreCompile()},
		&Template{"assetfsclean.go", templates.Clean()})

	for _, p := range r.Plugins {
		tpls = append(tpls, p.GetTemplates()...)
	}

	return tpls
}

func (r *Repository) InitWithTemplates(tpls []*Template) {
	absPath := r.AbsPath(true)
	if absPath == "" {
		panic("Invalid absPath.")
	}

	tplsi := make([]interface{}, len(tpls))
	gitIgnore := "data\ndata.go\n"

	for i, t := range tpls {
		tplsi[i] = r.template(t)
		gitIgnore += t.Name + "\n"
	}

	tplsi = append(tplsi, &path_helpers.Path{Alias: ".gitignore", Data: []byte(gitIgnore)})

	if err := path_helpers.CopyTree(absPath, tplsi); err != nil {
		panic(err)
	}
}

func (r *Repository) Init() {
	r.InitWithTemplates(r.GetInitTempĺates())
	for _,p := range r.Plugins {
		p.Init(r)
	}
}

func (r *Repository) Sync() {
	for _, cb := range r.preSync {
		cb(r)
	}
	sdest := r.DataDir(true)
	os.RemoveAll(sdest)
	if err := path_helpers.CopyTree(sdest, r.Sources); err != nil {
		panic(err)
	}
	config := bindata.NewConfig()
	config.Input = []bindata.InputConfig{
		{
			Path:      sdest,
			Recursive: true,
		},
	}
	config.Package = filepath.Base(r.absPath)
	config.Tags = r.BindataTag
	config.Output = r.BinFile()
	config.Prefix = sdest
	config.NoMetadata = true

	if r.PrepareConfig != nil {
		r.PrepareConfig(config)
	}

	if err := bindata.Translate(config); err != nil {
		panic(err)
	}
	for _, cb := range r.afterSync {
		cb(r)
	}
}

func (r *Repository) Clean() {
	for _, cb := range r.preClean {
		cb(r)
	}
	sdest := r.DataDir(true)
	os.RemoveAll(sdest)
	binFile := r.BinFile()
	if path_helpers.IsExistingRegularFile(binFile) {
		os.Remove(binFile)
	}
	for _, cb := range r.afterClean {
		cb(r)
	}
}

func (r *Repository) PreSourceAdd(cbcs ...PreSourceAddFunc) {
	r.preSourceAdd = append(r.preSourceAdd, cbcs...)
}
func (r *Repository) AfterSourceAdd(cbcs ...AfterSourceAddFunc) {
	r.afterSourceAdd = append(r.afterSourceAdd, cbcs...)
}
func (r *Repository) PreSync(cbcs ...func(repo Interface)) {
	r.preSync = append(r.preSync, cbcs...)
}
func (r *Repository) AfterSync(cbcs ...func(repo Interface)) {
	r.afterSync = append(r.afterSync, cbcs...)
}
func (r *Repository) PreClean(cbcs ...func(repo Interface)) {
	r.preClean = append(r.preClean, cbcs...)
}
func (r *Repository) AfterClean(cbcs ...func(repo Interface)) {
	r.afterClean = append(r.afterClean, cbcs...)
}
