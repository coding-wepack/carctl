package ioutils

import (
	"io"
)

func QuiteClose(r io.ReadCloser) {
	_ = r.Close()
}
