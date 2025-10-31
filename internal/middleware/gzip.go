package middleware

import (
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/mrhyman/shortner/internal/logger"
	"github.com/mrhyman/shortner/internal/model"
)

type gzipWriter struct {
	w  http.ResponseWriter
	zw *gzip.Writer
}

func newCompressWriter(w http.ResponseWriter) *gzipWriter {
	return &gzipWriter{
		w:  w,
		zw: gzip.NewWriter(w),
	}
}

func (c *gzipWriter) Header() http.Header {
	return c.w.Header()
}

func (c *gzipWriter) Write(p []byte) (int, error) {
	return c.zw.Write(p)
}

func (c *gzipWriter) WriteHeader(statusCode int) {
	if statusCode < http.StatusMultipleChoices {
		c.w.Header().Set("Content-Encoding", "gzip")
	}
	c.w.WriteHeader(statusCode)
}

func (c *gzipWriter) Close() error {
	return c.zw.Close()
}

type compressReader struct {
	r  io.ReadCloser
	zr *gzip.Reader
}

func newCompressReader(ctx context.Context, r io.ReadCloser) (*compressReader, error) {
	zr, err := gzip.NewReader(r)
	if err != nil {
		logger.FromContext(ctx).With("err", err.Error()).Error()
		return nil, err
	}

	return &compressReader{
		r:  r,
		zr: zr,
	}, nil
}

func (c compressReader) Read(p []byte) (n int, err error) {
	return c.zr.Read(p)
}

func (c *compressReader) Close() error {
	if err := c.r.Close(); err != nil {
		return err
	}
	return c.zr.Close()
}

func WithGzip(next http.HandlerFunc) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		log := logger.FromContext(req.Context())
		ow := res

		acceptEncoding := req.Header.Get("Accept-Encoding")
		contentType := req.Header.Get("Content-Type")
		if strings.HasPrefix(contentType, "application/json") || strings.HasPrefix(contentType, "text/plain") {
			supportsGzip := strings.Contains(acceptEncoding, "gzip")
			if supportsGzip {
				cw := newCompressWriter(res)
				ow = cw
				defer cw.Close()
			}
		}

		contentEncoding := req.Header.Get("Content-Encoding")
		sendsGzip := strings.Contains(contentEncoding, "gzip")
		if sendsGzip {
			cr, err := newCompressReader(req.Context(), req.Body)
			if err != nil {
				log.With("err", model.ErrCompressReading.Error(), "trace", err.Error())
				res.WriteHeader(http.StatusInternalServerError)
				fmt.Fprint(res, model.ErrCompressReading.Error())
				return
			}
			req.Body = cr
			defer cr.Close()
		}

		next.ServeHTTP(ow, req)
	}
}
