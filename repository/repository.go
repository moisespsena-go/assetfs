package repository

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"text/template"

	"github.com/moisespsena-go/file-utils"
	"github.com/moisespsena-go/sortvalues"
	"github.com/moisespsena-go/xbindata"
	"github.com/moisespsena/go-assetfs/repository/api"
	"github.com/moisespsena/go-assetfs/repository/templates"
	"github.com/moisespsena/go-error-wrap"
	"github.com/moisespsena/go-path-helpers"
)

const DefaultBS = 1024 * 1024 * 20

var DefaultBufferSize = DefaultBS

type Template = api.Template

type PrepareConfigCallbackValue struct {
	sortvalues.ValueInterface
}

func (cb PrepareConfigCallbackValue) Callback() func(config *xbindata.Config) {
	return cb.Value().(func(config *xbindata.Config))
}

func PrepareConfigCallback(f func(config *xbindata.Config), name ...string) PrepareConfigCallbackValue {
	return PrepareConfigCallbackValue{sortvalues.NewValue(f, name...)}
}

type Repository struct {
	PackagePath       string
	Package           string
	BindataCompileTag string
	BindataCleanTag   string
	BindataTag        string
	Templates         []*Template
	prepareConfig     sortvalues.Sorter
	Sources           []fileutils.Copier
	sources           map[interface{}]int
	absPath           string
	Plugins           []api.Plugin
	preSourceAdd      []api.PreSourceAddFunc
	afterSourceAdd    []api.AfterSourceAddFunc
	preClean          []func(repo api.Interface)
	afterClean        []func(repo api.Interface)
	preSync           []func(repo api.Interface)
	afterSync         []func(repo api.Interface)
	dumpers           []api.Dumper
	ignorePaths        []func(pth string) bool
}

func NewRepository(packagePath string) *Repository {
	return &Repository{PackagePath: packagePath, Package: filepath.Base(packagePath),
		BindataCompileTag: "assetfs_bindataCompile", BindataTag: "assetfs_bindata", BindataCleanTag: "assetfs_bindataClean"}
}

func (r *Repository) PrepareConfig(f func(config *xbindata.Config), name ...string) (v api.PrepareConfigCallbackValueInterface) {
	v = PrepareConfigCallback(f, name...)
	if err := r.prepareConfig.Append(v); err != nil {
		panic(err)
	}
	return
}

func (r *Repository) RegisterPlugin(plugins ...api.Plugin) {
	r.Plugins = append(r.Plugins, plugins...)
}

func (r *Repository) AddSourcePath(sources ...*fileutils.Dir) {
	for _, p := range sources {
		r.AddSource(p)
	}
}

func (r *Repository) AddSource(sources ...fileutils.Copier) {
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
				perms, err := path_helpers.ResolvePerms(absPath)
				if err != nil {
					panic(fmt.Errorf("Error on resolv mode: %v", err))
				}
				if err := os.MkdirAll(absPath, os.FileMode(perms)); err != nil {
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
		perms, err := path_helpers.ResolvePerms(absPath)
		if err != nil {
			panic(fmt.Errorf("Error on resolv mode: %v", err))
		}
		if err := os.MkdirAll(absPath, os.FileMode(perms)); err != nil {
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

func (r *Repository) template(tpl *Template) fileutils.Copier {
	return &fileutils.SrcData{Destation: fileutils.Destation{tpl.Name}, Data: r.renderTemplate(tpl.Data)}
}

func (r *Repository) GetInitTempĺates() []*Template {
	tpls := r.Templates[:]
	tpls = append(tpls,
		&Template{"assetfs.go", templates.FS()},
		&Template{"assetfsCommon.go", templates.Common()},
		&Template{"assetfsBindata.go", templates.FSBindata()},
		&Template{"assetfsBindataCompile.go", templates.PreCompile()},
		&Template{"assetfsBindataClean.go", templates.Clean()})

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

	tplsi := make([]fileutils.Copier, len(tpls))
	gitIgnore := "data\ndata.go\n"

	for i, t := range tpls {
		tplsi[i] = r.template(t)
		gitIgnore += t.Name + "\n"
	}

	tplsi = append(tplsi, &fileutils.SrcData{Destation: fileutils.Destation{".gitignore"}, Data: []byte(gitIgnore)})

	if err := fileutils.CopyTree(absPath, tplsi); err != nil {
		panic(err)
	}
}

func (r *Repository) Init() {
	r.InitWithTemplates(r.GetInitTempĺates())
	for _, p := range r.Plugins {
		p.Init(r)
	}
}

func (r *Repository) IsIgnorePath(pth string) bool {
	for _, f := range r.ignorePaths {
		if f(pth) {
			return true
		}
	}
	return false
}

func (r *Repository) copy(sdest string) {
	var err error

	for i, dump := range r.dumpers {
		err = dump(func(pth string, stat os.FileInfo, reader io.Reader) error {
			opath := filepath.Join(sdest, pth)
			if stat.IsDir() {
				if !path_helpers.IsExistingDir(opath) {
					err := os.MkdirAll(opath, stat.Mode())
					if err != nil {
						return fmt.Errorf("Sync: Dump[%v] mkdir %q: %v", i, pth, err)
					}
				}
				return nil
			} else {
				dirName := filepath.Dir(opath)
				if _, err := os.Stat(dirName); os.IsNotExist(err) {
					mode, err := path_helpers.ResolvePerms(dirName)
					if err != nil {
						return errwrap.Wrap(err, "Resolve permissions of %q", dirName)
					}
					if err := os.MkdirAll(dirName, os.FileMode(mode)); err != nil {
						return fmt.Errorf("Sync: Dump[%v] mkdir %q: %v", i, pth, err)
					}
				}
			}

			if err := fileutils.CreateFileSync(opath, reader, stat); err != nil {
				return fmt.Errorf("Sync: Dump[%v] create %q: %v", i, pth, err)
			}
			return nil
		}, r.IsIgnorePath)
		if err != nil {
			panic(err)
		}
	}
}
func (r *Repository) Sync() {
	for _, cb := range r.preSync {
		cb(r)
	}
	sdest := r.DataDir(true)

	if err := fileutils.CopyTree(sdest, r.Sources); err != nil {
		panic(err)
	}

	r.copy(sdest)

	config := xbindata.NewConfig()
	config.Input = []xbindata.InputConfig{
		{
			Path:      sdest,
			Recursive: true,
		},
	}
	config.Package = filepath.Base(r.absPath)
	config.Tags = r.BindataTag
	config.Output = r.BinFile()
	config.Prefix = sdest
	config.FileSystem = true
	config.Embed = true

	prepareConfig, err := r.prepareConfig.Sort()
	if err != nil {
		panic(err)
	}
	for _, pc := range prepareConfig {
		pc.Value().(func(config *xbindata.Config))(config)
	}

	if err := xbindata.Translate(config); err != nil {
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

func (r *Repository) PreSourceAdd(cbcs ...api.PreSourceAddFunc) {
	r.preSourceAdd = append(r.preSourceAdd, cbcs...)
}
func (r *Repository) AfterSourceAdd(cbcs ...api.AfterSourceAddFunc) {
	r.afterSourceAdd = append(r.afterSourceAdd, cbcs...)
}
func (r *Repository) PreSync(cbcs ...func(repo api.Interface)) {
	r.preSync = append(r.preSync, cbcs...)
}
func (r *Repository) AfterSync(cbcs ...func(repo api.Interface)) {
	r.afterSync = append(r.afterSync, cbcs...)
}
func (r *Repository) PreClean(cbcs ...func(repo api.Interface)) {
	r.preClean = append(r.preClean, cbcs...)
}
func (r *Repository) AfterClean(cbcs ...func(repo api.Interface)) {
	r.afterClean = append(r.afterClean, cbcs...)
}
func (r *Repository) Dumper(dumpers ...api.Dumper) {
	r.dumpers = append(r.dumpers, dumpers...)
}

func (r *Repository) IgnorePath(f ...func(pth string) bool) {
	r.ignorePaths = append(r.ignorePaths, f...)
}