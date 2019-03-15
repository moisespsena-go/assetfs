package assetfs

import (
	"bytes"
	"io"
)

type BytesReadCloser struct {
	*bytes.Reader
}

func NewBytesReadCloser(b []byte) io.ReadCloser {
	return &BytesReadCloser{bytes.NewReader(b)}
}

func (b *BytesReadCloser) Close() error {
	b.Reader = nil
	return nil
}
