package serverapi

import (
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type compressWriter struct {
	w  http.ResponseWriter
	zw *gzip.Writer
}

func newCompressWriter(w http.ResponseWriter) *compressWriter {
	return &compressWriter{
		w:  w,
		zw: gzip.NewWriter(w),
	}
}

func (c *compressWriter) Header() http.Header {
	return c.w.Header()
}

func (c *compressWriter) Write(p []byte) (int, error) {
	return c.zw.Write(p)
}

func (c *compressWriter) WriteHeader(statusCode int) {
	c.w.Header().Set("Content-Encoding", "gzip")
	c.w.WriteHeader(statusCode)
}

func (c *compressWriter) Close() error {
	return c.zw.Close()
}

type compressReader struct {
	r  io.ReadCloser
	zr *gzip.Reader
}

func newCompressReader(r io.ReadCloser) (*compressReader, error) {
	zr, err := gzip.NewReader(r)
	if err != nil {
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

func gzipMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextW := w
		if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
			cr, err := newCompressReader(r.Body)
			if err != nil {
				responseError(w, fmt.Errorf("gzip: read compressed body: %w", err))
				return
			}
			r.Body = cr
			defer cr.Close()
		}
		if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			w.Header()
			cw := newCompressWriter(w)
			nextW = cw
			defer cw.Close()
		}

		h.ServeHTTP(nextW, r)
	})
}
