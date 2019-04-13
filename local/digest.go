package local

import (
	"crypto/sha256"
	"errors"
	"io"
	"os"
)

func Digest(pth string) (digest *[sha256.Size]byte, err error) {
	f, err := os.Open(pth)
	if err != nil {
		return nil, err
	}

	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return nil, err
	}

	h := sha256.New()
	var n int64
	if n, err = io.Copy(h, f); err != nil {
		return nil, err
	}
	if n != info.Size() {
		return nil, errors.New("Digest of `" + pth + "` failed: readed size is not eq to file size.")
	}
	var d [sha256.Size]byte
	copy(d[:], h.Sum(nil))
	return &d, nil
}
