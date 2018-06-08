package main

import (
	"github.com/moisespsena/go-assetfs"
	"fmt"
	"github.com/moisespsena/go-assetfs/api"
)

func main() {
	fs := assetfs.NewAssetFileSystem()
	ns := fs.NameSpace("z")
	ns.RegisterPath("t/ns")
	fs.RegisterPath("t/data")
	fs.RegisterPath("t/data2")
	ns.WalkInfo(".", func(info api.FileInfo) error {
		fmt.Println(info, "->", info.RealPath())
		return nil
	})
	fmt.Println("-------------")
	fs.WalkInfo(".", func(info api.FileInfo) error {
		fmt.Println(info, "->", info.RealPath())
		return nil
	})
	fmt.Println("-------------")
	fs.ReadDir(".", func(info api.FileInfo) error {
		fmt.Println(info, "->", info.RealPath())
		return nil
	}, false)
	fmt.Println("-------------")
	matches, _ := fs.NewGlobString(">*.txt").Names()
	fmt.Println(matches)
	fmt.Println("-------------")
	fs.Dump(func(info api.FileInfo) error {
		fmt.Println(info, "->", info.RealPath())
		return nil
	})
	fmt.Println("-------------")
	fs.PathsFrom("z/a", func(pth string) error {
		fmt.Println(pth)
		return nil
	})
	fmt.Println("-------------")
	asset, err := fs.Asset("z/a/x.txt")
	fmt.Println(string(asset.GetData()))
	fmt.Println(err)

}
