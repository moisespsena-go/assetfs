package templates

func PreCompile() string {
	return `// +build {{.PreCompileTag}}

package {{.Package}}

import "fmt"

func init() {
	AddCallback(func() {
		fmt.Println("Pre Compiling '{{.Package}}'") 
		Repository.AddSourcePath(FileSystem.GetPaths(true)...)
		Repository.Sync()
		fmt.Println("Pre Compiling '{{.Package}}' done.")
	})
}
`
}