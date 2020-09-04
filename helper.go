package assetfs

import (
	"compress/gzip"
	"io"
	"io/ioutil"
)


// IsGzip returns true if the string is gzipped data
func IsGzip(b []byte) bool {
	if len(b) < 2 {
		return false
	}
	return b[0] == 0x1f && b[1] == 0x8b
}

func Data(info interface{ Reader() (io.ReadCloser, error) }) (b []byte, err error) {
	var r io.ReadCloser
	if r, err = info.Reader(); err != nil {
		return
	}
	defer r.Close()
	if c, ok := r.(Compresseder); ok && c.Compressed() {
		if r, err = gzip.NewReader(r); err != nil {
			return
		}
	}
	b, err = ioutil.ReadAll(r)
	return
}

func DataS(info interface{ Reader() (io.ReadCloser, error) }) (s string, err error) {
	var b []byte
	if b, err = Data(info); err != nil {
		return
	}
	s = string(b)
	return
}

func MustData(info interface{ Reader() (io.ReadCloser, error) }) (b []byte) {
	var err error
	if b, err = Data(info); err != nil {
		panic(err)
	}
	return
}

func MustDataS(info interface{ Reader() (io.ReadCloser, error) }) (s string) {
	var err error
	if s, err = DataS(info); err != nil {
		panic(err)
	}
	return
}