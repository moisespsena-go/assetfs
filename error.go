package assetfs

import (
	"os"
	"fmt"
)

type AssetError struct {
	Name string
	Err  error
}

func (ar *AssetError) Error() string {
	return fmt.Sprintf("Asset %q: %v", ar.Name, ar.Err)
}

func NotFound(name string) error {
	return &AssetError{name, os.ErrNotExist}
}

func IsNotFound(err error) (ok bool) {
	if ae, ok := err.(*AssetError); ok && os.IsNotExist(ae.Err) {
		return true
	}
	return false
}
