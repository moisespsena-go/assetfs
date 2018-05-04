package templates

func Repository() string {
	return `package {{.Package}}

import (
	"github.com/moisespsena/go-path-helpers"
	"github.com/moisespsena/go-assetfs/repository"
)

var (
	DIR = path_helpers.GetCalledDir(true)
	Repository repository.Interface = repository.NewRepository(path_helpers.GetCalledDir())
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