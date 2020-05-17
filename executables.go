package assetfs

import (
	"io"
	"os"
	"path/filepath"

	path_helpers "github.com/moisespsena-go/path-helpers"
	"github.com/pkg/errors"

	"github.com/moisespsena-go/assetfs/assetfsapi"
	fileutils "github.com/moisespsena-go/file-utils"
)

func SaveExecutable(assetName string, fs assetfsapi.Interface, dest string, force ...bool) (err error) {
	name := filepath.Base(assetName)

	var (
		s     os.FileInfo
		asset assetfsapi.FileInfo
	)

	if asset, err = fs.AssetInfo(assetName); err != nil {
		err = errors.Wrapf(err, "Get Asset %q", assetName)
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
			err = errors.Wrapf(err, "Stat of %q", dest)
			return
		}
	}

	if err = path_helpers.MkdirAllIfNotExists(filepath.Dir(dest)); err != nil {
		return
	}

	var in io.ReadCloser

	if in, err = asset.Reader(); err != nil {
		err = errors.Wrapf(err, "Get Reader of %q", assetName)
		return
	}

	defer in.Close()

	if err = fileutils.CopyReaderInfo(in, asset, dest); err != nil {
		err = errors.Wrapf(err, "`%s` executable %q creation failed", name, dest)
	}

	return
}
