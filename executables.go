package assetfs

import (
	"io"
	"os"
	"path/filepath"

	"github.com/moisespsena-go/file-utils"
	"github.com/moisespsena/go-assetfs/assetfsapi"
	"github.com/moisespsena/go-error-wrap"
	"github.com/moisespsena-go/path-helpers"
)

func SaveExecutable(assetName string, fs assetfsapi.Interface, dest string, force ...bool) (err error) {
	name := filepath.Base(assetName)

	var (
		s     os.FileInfo
		asset assetfsapi.FileInfo
	)

	if asset, err = fs.AssetInfo(assetName); err != nil {
		err = errwrap.Wrap(err, "Get Asset %q", assetName)
		return
	}

	if len(force) == 0 || !force[0] {
		if s, err = os.Stat(dest); err == nil {
			if os.SameFile(s, asset) {
				return
			}
		} else if os.IsNotExist(err) {
			err = nil
		} else {
			err = errwrap.Wrap(err, "Stat of %q", dest)
			return
		}
	}

	if err = path_helpers.MkdirAllIfNotExists(filepath.Dir(dest)); err != nil {
		return
	}

	var in io.ReadCloser

	if in, err = asset.Reader(); err != nil {
		err = errwrap.Wrap(err, "Get Reader of %q", assetName)
		return
	}

	defer in.Close()

	if err = fileutils.CopyReaderInfo(in, asset, dest); err != nil {
		err = errwrap.Wrap(err, "`%s` executable %q creation failed", name, dest)
	}

	return
}
