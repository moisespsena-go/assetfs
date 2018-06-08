package assetfs

import (
	"github.com/moisespsena/go-assetfs/api"
)

type Asset struct {
	Name string
	Data []byte
}

func NewAsset(name string, data []byte) api.AssetInterface {
	return &Asset{name, data}
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
