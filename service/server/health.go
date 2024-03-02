package server

import (
	"io"
	"log/slog"
	"net/http"
)

func handleHealthZ() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slog.Debug("health check")
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, "OK")
	}
}
