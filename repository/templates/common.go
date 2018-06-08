package templates

func Common() string {
	return `package {{.Package}}

import "github.com/moisespsena/go-path-helpers"

var (
	DIR = path_helpers.GetCalledDir(true)
	Repository = AssetFS.NewRepository(path_helpers.GetCalledDir())	
	callbacks  []func()
)

func AddCallback(f ...func()) {
	callbacks = append(callbacks, f...) 
}

func CallCallbacks() {
	for _, f := range callbacks {
		f()
	}
}
`
}