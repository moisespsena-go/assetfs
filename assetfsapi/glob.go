package assetfsapi

type GlobPattern interface {
	Dir() string
	Pattern() string
	IsRecursive() bool
	AllowDirs() bool
	AllowFiles() bool
	Recursive() GlobPattern
	Match(value string) bool
	Wrap(dir ...string) GlobPattern
	GetPathFormatter() PathFormatterFunc
	PathFormatter(formatter PathFormatterFunc) GlobPattern
}

type Glob interface {
	GetPattern() GlobPattern
	SetPattern(pattern GlobPattern)
	FS() Interface
	Name(cb func(pth string, isDir bool) error) error
	NameOrPanic(cb func(pth string, isDir bool) error)
	Names() ([]string, error)
	SortedNames() ([]string, error)
	NamesOrPanic() []string
	Info(cb func(info FileInfo) error) error
	Infos() ([]FileInfo, error)
	SortedInfos() (items []FileInfo, err error)
	InfoOrPanic(cb func(info FileInfo) error)
	InfosOrPanic() []FileInfo
}

type GlobError struct {
	Err error
}

func (ge GlobError) Error() string {
	return ge.Err.Error()
}
