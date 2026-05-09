package metricsserver

import (
	"context"
	"net/http"
	"strings"
	"time"

	platformmetrics "cityhawk/backend/internal/platform/metrics"
)

func Start(addr string) (func(), error) {
	if strings.TrimSpace(addr) == "" {
		return func() {}, nil
	}

	mux := http.NewServeMux()
	mux.Handle("/metrics", platformmetrics.Handler())
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	server := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
		close(errCh)
	}()

	select {
	case err := <-errCh:
		return nil, err
	default:
	}

	cleanup := func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = server.Shutdown(ctx)
	}
	return cleanup, nil
}
