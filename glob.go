package assetfs

import (
	"path"
	"sort"

	"github.com/gobwas/glob"
	"github.com/gobwas/glob/syntax"
	"github.com/moisespsena-go/assetfs/assetfsapi"
)

type GlobPatter = assetfsapi.GlobPattern

type GlobBase struct {
	dir           string
	pattern       string
	recursive     bool
	files         bool
	dirs          bool
	pathFormatter assetfsapi.PathFormatterFunc
}

func (gp *GlobBase) Dir() string {
	return gp.dir
}
func (gp *GlobBase) Pattern() string {
	return gp.pattern
}
func (gp *GlobBase) IsRecursive() bool {
	return gp.recursive
}
func (gp *GlobBase) AllowDirs() bool {
	return gp.dirs
}
func (gp *GlobBase) AllowFiles() bool {
	return gp.files
}
func (gp *GlobBase) GetPathFormatter() assetfsapi.PathFormatterFunc {
	return gp.pathFormatter
}

type DefaultGlobPattern struct {
	GlobBase
	glob glob.Glob
}

func (gp *DefaultGlobPattern) Match(value string) bool {
	return gp.glob.Match(value)
}
func (gp *DefaultGlobPattern) Glob() glob.Glob {
	return gp.glob
}
func (gp DefaultGlobPattern) Recursive() assetfsapi.GlobPattern {
	gp.recursive = true
	return &gp
}

func (gp DefaultGlobPattern) Wrap(dir ...string) assetfsapi.GlobPattern {
	gp.dir = path.Join(append(dir, gp.dir)...)
	return &gp
}
func (gp DefaultGlobPattern) PathFormatter(formatter assetfsapi.PathFormatterFunc) assetfsapi.GlobPattern {
	gp.pathFormatter = formatter
	return &gp
}

type NormalPattern struct {
	GlobBase
}

func (gp *NormalPattern) Recursive() assetfsapi.GlobPattern {
	clone := *gp
	clone.recursive = true
	return &clone
}

func (gp *NormalPattern) Match(value string) bool {
	return value == gp.pattern
}

func (gp NormalPattern) Wrap(dir ...string) assetfsapi.GlobPattern {
	gp.dir = path.Join(append(dir, gp.dir)...)
	return &gp
}
func (gp NormalPattern) PathFormatter(formatter assetfsapi.PathFormatterFunc) assetfsapi.GlobPattern {
	gp.pathFormatter = formatter
	return &gp
}

// pattern: \f Files, \r dirs
func NewGlobPattern(pattern string) assetfsapi.GlobPattern {
	recursive := pattern[0] == '>'
	if recursive {
		pattern = pattern[1:]
	}
	dirs := true
	files := true

	if pattern[0] == '\r' {
		files = false
		pattern = pattern[1:]
	} else if pattern[0] == '\f' {
		dirs = false
		pattern = pattern[1:]
	}
	dir, pattern := path.Split(pattern)
	var hasSpecial bool
	for i := 0; i < len(pattern); i++ {
		if syntax.Special(pattern[i]) {
			hasSpecial = true
			break
		}
	}
	base := GlobBase{dir, pattern, recursive, files, dirs, func(pth *string) {}}
	if hasSpecial {
		return &DefaultGlobPattern{base, glob.MustCompile(pattern)}
	}
	return &NormalPattern{base}
}

var G = NewGlobPattern

type Glob struct {
	fs      assetfsapi.Interface
	pattern assetfsapi.GlobPattern
}

func NewGlob(fs assetfsapi.Interface, pattern assetfsapi.GlobPattern) *Glob {
	return &Glob{fs, pattern}
}

func (g *Glob) GetPattern() assetfsapi.GlobPattern {
	return g.pattern
}

func (g *Glob) SetPattern(pattern assetfsapi.GlobPattern) {
	g.pattern = pattern
}

func (g *Glob) FS() assetfsapi.Interface {
	return g.fs
}

func (g *Glob) Name(cb func(pth string, isDir bool) error) error {
	return g.fs.Glob(g.pattern, cb)
}

func (g *Glob) NameOrPanic(cb func(pth string, isDir bool) error) {
	err := g.Name(cb)
	if err != nil {
		panic(err)
	}
}

func (g *Glob) Names() (items []string, err error) {
	err = g.Name(func(pth string, isDir bool) error {
		items = append(items, pth)
		return nil
	})
	return
}

func (g *Glob) SortedNames() (items []string, err error) {
	if items, err = g.Names(); err == nil {
		sort.Strings(items)
	}
	return
}

func (g *Glob) NamesOrPanic() []string {
	items, err := g.Names()
	if err != nil {
		panic(err)
	}
	return items
}

func (g *Glob) Info(cb func(info assetfsapi.FileInfo) error) error {
	return g.fs.GlobInfo(g.pattern, cb)
}

func (g *Glob) InfoOrPanic(cb func(info assetfsapi.FileInfo) error) {
	err := g.Info(cb)
	if err != nil {
		panic(err)
	}
}

func (g *Glob) Infos() (items []assetfsapi.FileInfo, err error) {
	err = g.Info(func(info assetfsapi.FileInfo) error {
		items = append(items, info)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return
}

func (g *Glob) SortedInfos() (items []assetfsapi.FileInfo, err error) {
	if items, err = g.Infos(); err == nil {
		sort.Slice(items, func(i, j int) bool {
			return items[i].Path() < items[j].Path()
		})
	}
	return
}

func (g *Glob) InfosOrPanic() []assetfsapi.FileInfo {
	items, err := g.Infos()
	if err != nil {
		panic(err)
	}
	return items
}
