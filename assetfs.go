package assetfs

import (
	"github.com/moisespsena-go/assetfs/assetfsapi"
)

type Asset struct {
	Path string
	Name string
	Data []byte
}

func NewAsset(path, name string, data []byte) assetfsapi.AssetInterface {
	return &Asset{path, name, data}
}

func (a *Asset) GetData() []byte {
	return a.Data
}

func (a *Asset) GetString() string {
	return string(a.Data)
}

func (a *Asset) GetName() string {
	return a.Name
}

func (a *Asset) GetPath() string {
	return a.Path
}