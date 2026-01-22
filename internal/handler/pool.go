package handler

import (
	"bufio"
	"bytes"
	"io"
	"sync"
)

var (
	bufferPool = sync.Pool{
		New: func() any {
			return new(bytes.Buffer)
		},
	}
	readerPool = sync.Pool{
		New: func() any {
			return bufio.NewReader(nil)
		},
	}
)

func GetBuffer() *bytes.Buffer {
	return bufferPool.Get().(*bytes.Buffer)
}

func PutBuffer(buf *bytes.Buffer) {
	buf.Reset()
	bufferPool.Put(buf)
}

func GetReader(r io.Reader) *bufio.Reader {
	reader := readerPool.Get().(*bufio.Reader)
	reader.Reset(r)
	return reader
}

func PutReader(r *bufio.Reader) {
	r.Reset(nil)
	readerPool.Put(r)
}
