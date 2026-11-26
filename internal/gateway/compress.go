package gateway

import (
	"compress/gzip"
	"net/http"
	"path/filepath"
	"strings"
)

// gzipAssets compresses text assets on the fly. Video segments under /hls and
// API responses use their own handlers and stay untouched.
func gzipAssets(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodHead || r.Header.Get("Range") != "" ||
			!acceptsGzip(r) || !compressible(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}
		writer := &gzipResponseWriter{ResponseWriter: w}
		next.ServeHTTP(writer, r)
		_ = writer.finish()
	})
}

func acceptsGzip(r *http.Request) bool {
	for _, value := range strings.Split(r.Header.Get("Accept-Encoding"), ",") {
		if strings.TrimSpace(strings.SplitN(value, ";", 2)[0]) == "gzip" {
			return true
		}
	}
	return false
}

func compressible(path string) bool {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".css", ".html", ".js", ".json", ".svg", ".txt":
		return true
	}
	return false
}

// gzipResponseWriter delays the header until the first body byte, so a 304 or
// a redirect keeps its status without a Content-Encoding it does not have.
type gzipResponseWriter struct {
	http.ResponseWriter
	status  int
	started bool
	writer  *gzip.Writer
}

func (w *gzipResponseWriter) WriteHeader(status int) {
	if w.started || w.status != 0 {
		return
	}
	w.status = status
}

func (w *gzipResponseWriter) Write(data []byte) (int, error) {
	if !w.started {
		w.start()
	}
	return w.writer.Write(data)
}

func (w *gzipResponseWriter) start() {
	w.started = true
	if w.status == 0 {
		w.status = http.StatusOK
	}
	header := w.Header()
	header.Del("Content-Length")
	header.Set("Content-Encoding", "gzip")
	header.Add("Vary", "Accept-Encoding")
	w.ResponseWriter.WriteHeader(w.status)
	w.writer = gzip.NewWriter(w.ResponseWriter)
}

// finish writes the gzip trailer, or forwards a status that produced no body.
func (w *gzipResponseWriter) finish() error {
	if w.writer != nil {
		return w.writer.Close()
	}
	if !w.started && w.status != 0 {
		w.ResponseWriter.WriteHeader(w.status)
	}
	return nil
}
