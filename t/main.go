package main

import (
	"fmt"
	"github.com/moisespsena-go/assetfs"
	"github.com/moisespsena-go/assetfs/assetfsapi"
)

func main() {
	fs := assetfs.NewAssetFileSystem()
	ns := fs.NameSpace("z")
	ns.RegisterPath("t/ns")
	fs.RegisterPath("t/data")
	fs.RegisterPath("t/data2")
	/*fmt.Println("------walk info from NS 'z' -------")
	ns.WalkInfo(".", func(info api.FileInfo) error {
		fmt.Println(info, "->", info.RealPath())
		return nil
	})
	fmt.Println("------walk info from FS '.' -------")
	fs.WalkInfo(".", func(info api.FileInfo) error {
		fmt.Println(info, "->", info.RealPath())
		return nil
	})
	fmt.Println("------ readir from FS '.' -------")
	fs.ReadDir(".", func(info api.FileInfo) error {
		fmt.Println(info, "->", info.RealPath())
		return nil
	}, false)
	fmt.Println("------ glob string from FS '.' -------")
	matches, _ := fs.NewGlobString(">*.txt").Names()
	fmt.Println(matches)
	fmt.Println("------ FS '.' DUMP -------")*/
	fs.Dump(func(info assetfsapi.FileInfo) error {
		fmt.Println(info, "->", info.RealPath())
		return nil
	})
	/*fmt.Println("---- paths from z/a ---------")
	fs.PathsFrom("z/a", func(pth string) error {
		fmt.Println(pth)
		return nil
	})*/
	return
	fmt.Println("---- paths from z/a/x.txt ---------")
	fs.PathsFrom("z/a/x.txt", func(pth string) error {
		fmt.Println(pth)
		return nil
	})
/*	fmt.Println("-------------")
	asset, err := fs.Asset("z/a/x.txt")
	fmt.Println(string(asset.GetData()))
	fmt.Println(err)*/

}
