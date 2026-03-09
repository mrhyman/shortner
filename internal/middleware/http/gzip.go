// Package middleware реализует промежуточное ПО для обработки HTTP запросов.
package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"

	"github.com/mrhyman/shortner/internal/logger"
	"github.com/mrhyman/shortner/internal/model"
)

// gzipWriter оборачивает http.ResponseWriter для обеспечения GZIP сжатия.
type gzipWriter struct {
	w  http.ResponseWriter
	zw *gzip.Writer
}

// newCompressWriter создает новый экземпляр gzipWriter.
func newCompressWriter(w http.ResponseWriter) *gzipWriter {
	return &gzipWriter{
		w:  w,
		zw: gzip.NewWriter(w),
	}
}

// Header возвращает заголовки ответа.
func (c *gzipWriter) Header() http.Header {
	return c.w.Header()
}

// Write записывает данные в ответ с применением GZIP сжатия.
func (c *gzipWriter) Write(p []byte) (int, error) {
	return c.zw.Write(p)
}

// WriteHeader записывает код состояния HTTP в ответ.
// Устанавливает заголовок Content-Encoding в "gzip" для успешных ответов.
func (c *gzipWriter) WriteHeader(statusCode int) {
	if statusCode < http.StatusMultipleChoices {
		c.w.Header().Set("Content-Encoding", "gzip")
	}
	c.w.WriteHeader(statusCode)
}

// Close закрывает GZIP writer и освобождает ресурсы.
func (c *gzipWriter) Close() error {
	return c.zw.Close()
}

// compressReader оборачивает io.ReadCloser для обеспечения декомпрессии GZIP данных.
type compressReader struct {
	r  io.ReadCloser
	zr *gzip.Reader
}

// newCompressReader создает новый экземпляр compressReader для декомпрессии GZIP данных.
func newCompressReader(r io.ReadCloser) (*compressReader, error) {
	zr, err := gzip.NewReader(r)
	if err != nil {
		logger.Get().With("err", err.Error()).Error()
		return nil, err
	}

	return &compressReader{
		r:  r,
		zr: zr,
	}, nil
}

// Read читает и декомпрессирует данные из запроса.
func (c compressReader) Read(p []byte) (n int, err error) {
	return c.zr.Read(p)
}

// Close закрывает reader и освобождает ресурсы.
func (c *compressReader) Close() error {
	if err := c.r.Close(); err != nil {
		return err
	}
	return c.zr.Close()
}

// Flush сбрасывает буферизованные данные в underlying writer.
func (c *gzipWriter) Flush() {
	c.zw.Flush()
	if f, ok := c.w.(http.Flusher); ok {
		f.Flush()
	}
}

// WithGzip это middleware для сжатия и декомпрессии данных с использованием GZIP.
// Он автоматически сжимает ответы для клиентов, которые поддерживают GZIP,
// и декомпрессирует тела запросов, отправленные с GZIP сжатием.
func WithGzip(next http.HandlerFunc) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		log := logger.Get()
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
			cr, err := newCompressReader(req.Body)
			if err != nil {
				log.With("err", err.Error()).Error()
				res.WriteHeader(http.StatusBadRequest)
				res.Write([]byte(model.ErrCompressReading.Error()))
				return
			}
			req.Body = cr
			defer cr.Close()
		}

		next.ServeHTTP(ow, req)
	}
}
