package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"

	"github.com/fjnkt98/todo-go/server"
	_ "github.com/mattn/go-sqlite3"
)

func serve(ctx context.Context) error {
	port, err := strconv.Atoi(os.Getenv("PORT"))
	if err != nil {
		return fmt.Errorf("parse port: %w", err)
	}

	s, err := server.NewServer(port)
	if err != nil {
		return fmt.Errorf("create server: %w", err)
	}

	ctx, stop := signal.NotifyContext(ctx, os.Interrupt)
	defer stop()

	shutdown, err := server.SetupOTel(ctx)
	if err != nil {
		return err
	}
	defer func() {
		err = errors.Join(err, shutdown(ctx))
	}()

	errs := make(chan error, 1)
	go func() {
		slog.InfoContext(ctx, "start server", slog.Int("port", port))
		if err := s.ListenAndServe(); err != http.ErrServerClosed {
			errs <- err
		}
	}()

	select {
	case err = <-errs:
		return fmt.Errorf("server error: %w", err)
	case <-ctx.Done():
		if err := s.Shutdown(ctx); err != nil {
			return fmt.Errorf("shutdown server: %w", err)
		}
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
