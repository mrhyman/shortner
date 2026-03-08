package handler

import (
	"bufio"
	"bytes"

	"github.com/mrhyman/shortner/internal/pool"
)

var bufferPool = pool.New(func() *bytes.Buffer {
	return new(bytes.Buffer)
})

type resettableReader struct {
	*bufio.Reader
}

func (r *resettableReader) Reset() {
	r.Reader.Reset(nil)
}
