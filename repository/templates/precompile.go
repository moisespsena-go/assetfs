package templates

func PreCompile() string {
	return `// +build {{.BindataCompileTag}}

package {{.Package}}

import "fmt"

func init() {
	AddCallback(func() {
		fmt.Println("Pre Compiling '{{.Package}}'")
		Repository.Sync()
		fmt.Println("Pre Compiling '{{.Package}}' done.")
	})
}
`
}