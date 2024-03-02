package server

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	"gitlab.con/stream-ai/huddle/backend/service/middleware/loggermw"
)

func new(logger *slog.Logger) http.Handler {
	mux := http.NewServeMux()

	addRoutes(mux, logger)

	var handler http.Handler = mux
	handler = loggermw.New(logger, handler)

	return handler
}

func Run(ctx context.Context, logger *slog.Logger, addr string) error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	srv := new(logger)
	httpServer := http.Server{
		Addr:    addr,
		Handler: srv,
	}
	go func() {
		log.Printf("server listening on %s", httpServer.Addr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Fprintf(os.Stderr, "error listening and serving: %s\n", err)
		}
		log.Println("server stopped")
	}()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-ctx.Done()
		log.Println("shutting down server")
		shutdownCtx := context.Background()
		shutdownCtx, cancel := context.WithTimeout(shutdownCtx, 10*time.Second)
		defer cancel()
		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			fmt.Fprintf(os.Stderr, "error shutting down http server: %s\n", err)
		}
	}()
	wg.Wait()

	return nil
}
