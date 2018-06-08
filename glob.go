package assetfs

import (
	"github.com/gobwas/glob"
	"github.com/gobwas/glob/syntax"
	"github.com/moisespsena/go-assetfs/api"
	"path/filepath"
)

type GlobPatter = api.GlobPattern

type GlobBase struct {
	dir           string
	pattern       string
	recursive     bool
	files         bool
	dirs          bool
	pathFormatter api.PathFormatterFunc
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
func (gp *GlobBase) GetPathFormatter() api.PathFormatterFunc {
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
func (gp DefaultGlobPattern) Recursive() api.GlobPattern {
	gp.recursive = true
	return &gp
}

func (gp DefaultGlobPattern) Wrap(dir ...string) api.GlobPattern {
	gp.dir = filepath.Join(append(dir, gp.dir)...)
	return &gp
}
func (gp DefaultGlobPattern) PathFormatter(formatter api.PathFormatterFunc) api.GlobPattern {
	gp.pathFormatter = formatter
	return &gp
}

type NormalPattern struct {
	GlobBase
}

func (gp *NormalPattern) Recursive() api.GlobPattern {
	clone := *gp
	clone.recursive = true
	return &clone
}

func (gp *NormalPattern) Match(value string) bool {
	return value == gp.pattern
}

func (gp NormalPattern) Wrap(dir ...string) api.GlobPattern {
	gp.dir = filepath.Join(append(dir, gp.dir)...)
	return &gp
}
func (gp NormalPattern) PathFormatter(formatter api.PathFormatterFunc) api.GlobPattern {
	gp.pathFormatter = formatter
	return &gp
}

// pattern: \f Files, \r dirs
func NewGlobPattern(pattern string) api.GlobPattern {
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
	dir, pattern := filepath.Split(pattern)
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
	fs      api.Interface
	pattern api.GlobPattern
}

func NewGlob(fs api.Interface, pattern api.GlobPattern) *Glob {
	return &Glob{fs, pattern}
}

func (g *Glob) GetPattern() api.GlobPattern {
	return g.pattern
}

func (g *Glob) SetPattern(pattern api.GlobPattern) {
	g.pattern = pattern
}

func (g *Glob) FS() api.Interface {
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

func (g *Glob) NamesOrPanic() []string {
	items, err := g.Names()
	if err != nil {
		panic(err)
	}
	return items
}

func (g *Glob) Info(cb func(info api.FileInfo) error) error {
	return g.fs.GlobInfo(g.pattern, cb)
}

func (g *Glob) InfoOrPanic(cb func(info api.FileInfo) error) {
	err := g.Info(cb)
	if err != nil {
		panic(err)
	}
}

func (g *Glob) Infos() (items []api.FileInfo, err error) {
	err = g.Info(func(info api.FileInfo) error {
		items = append(items, info)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return
}

func (g *Glob) InfosOrPanic() []api.FileInfo {
	items, err := g.Infos()
	if err != nil {
		panic(err)
	}
	return items
}
