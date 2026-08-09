package main

import (
	"context"
	"embed"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

//go:embed templates
var templates embed.FS

func serve(ctx context.Context) error {
	port, err := strconv.Atoi(os.Getenv("PORT"))
	if err != nil {
		return fmt.Errorf("parse port: %w", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		t, err := template.ParseFS(templates, "templates/top.html", "templates/base.html")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "server error") //nolint: errcheck
			return
		}

		type Data struct {
			Title   string
			Content string
		}

		data := Data{
			Title:   "ToDo Go",
			Content: "This is my first html/template content.",
		}

		w.WriteHeader(http.StatusOK)
		if err := t.Execute(w, &data); err != nil {
			slog.ErrorContext(ctx, "write ResponseWriter", slog.Any("error", err))
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
