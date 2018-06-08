package templates

func Clean() string {
	return `// +build {{.BindataCleanTag}}

package {{.Package}}

func init() {
	AddCallback(Repository.Clean)
}
`
}