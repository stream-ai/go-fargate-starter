package loggermw

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"
)

type rw struct {
	http.ResponseWriter
	statusCode int
	resp       bytes.Buffer
}

func (r *rw) Write(b []byte) (int, error) {
	// copy b to r.resp using io.Copy
	_, err := io.Copy(&r.resp, bytes.NewReader(b))
	if err != nil {
		return 0, err
	}
	return r.ResponseWriter.Write(b)
}

func (r *rw) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.statusCode = statusCode
}

func (r *rw) Header() http.Header {
	return r.ResponseWriter.Header()
}

func New(logger *slog.Logger, h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rw := &rw{ResponseWriter: w}
		h.ServeHTTP(rw, r)
		if logger != nil {
			logger.Info("request",
				"method", r.Method,
				"path", r.URL.Path,
				"status", rw.statusCode,
				"response", string(rw.resp.String()),
			)
		}
	})
}
