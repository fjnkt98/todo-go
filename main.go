package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func serve(ctx context.Context) error {
	port, err := strconv.Atoi(os.Getenv("PORT"))
	if err != nil {
		return fmt.Errorf("parse port: %w", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		v := map[string]string{
			"message": "ok",
		}

		if err := json.NewEncoder(w).Encode(v); err != nil {
			slog.ErrorContext(r.Context(), "write resuponse body", slog.Any("error", err))
		}
	})

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	ctx, stop := signal.NotifyContext(ctx, os.Interrupt)
	defer stop()

	go func() {
		slog.InfoContext(ctx, "start server", slog.Int("port", port))
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			slog.ErrorContext(ctx, "server error", slog.Any("error", err))
			return
		}
	}()

	<-ctx.Done()
	if err := server.Shutdown(ctx); err != nil {
		return fmt.Errorf("shutdown server: %w", err)
	}
	slog.InfoContext(ctx, "shutdown server", slog.Int("port", port))
	return nil
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	ctx := context.Background()

	if err := serve(ctx); err != nil {
		slog.ErrorContext(ctx, "comnand failed", slog.Any("error", err))
		os.Exit(1)
	}
}
