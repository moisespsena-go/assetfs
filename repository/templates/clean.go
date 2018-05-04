package templates

func Clean() string {
	return `// +build {{.CleanTag}}

package {{.Package}}

func init() {
	AddCallback(Repository.Clean)
}
`
}