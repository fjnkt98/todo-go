package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/urfave/cli/v3"
)

type HelloHandler struct {
	db *sql.DB
}

func NewHelloHandler(db *sql.DB) *HelloHandler {
	return &HelloHandler{
		db: db,
	}
}

func (h *HelloHandler) HandleGET(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")

	message := map[string]string{
		"message": "Hello!",
	}

	if err := json.NewEncoder(w).Encode(message); err != nil {
		slog.Error("write resuponse body", slog.Any("error", err))
	}
}

func (h *HelloHandler) GetUsers(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.QueryContext(r.Context(), "SELECT id, name FROM users ORDER BY id ASC")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Header().Set("Content-Type", "application/json")

		v := map[string]string{
			"message": "server error",
		}

		if err := json.NewEncoder(w).Encode(v); err != nil {
			slog.Error("write resuponse body", slog.Any("error", err))
		}
	}
	defer func() {
		if err := rows.Close(); err != nil {
			slog.ErrorContext(r.Context(), "close rows")
		}
	}()

	users := make([]map[string]any, 0)

	for rows.Next() {
		var id int
		var name string
		if err := rows.Scan(&id, &name); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Header().Set("Content-Type", "application/json")
			v := map[string]string{
				"message": "server error",
			}

			if err := json.NewEncoder(w).Encode(v); err != nil {
				slog.Error("write resuponse body", slog.Any("error", err))
			}
		}
		users = append(users, map[string]any{"id": id, "name": name})
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(users); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Header().Set("Content-Type", "application/json")
		v := map[string]string{
			"message": "server error",
		}

		if err := json.NewEncoder(w).Encode(v); err != nil {
			slog.Error("write resuponse body", slog.Any("error", err))
		}
	}
}

func NewDB(ctx context.Context, dsn string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	statements := []string{
		"PRAGMA journal_mode = WAL;",
		"PRAGMA foreign_keys = on;",
		"PRAGMA defer_foreign_keys = on;",
	}

	for _, stmt := range statements {
		if _, err := db.ExecContext(ctx, ""); err != nil {
			return db, fmt.Errorf("execute initializer: %s, %w", stmt, err)
		}
	}

	return db, nil
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	ctx := context.Background()

	cmd := &cli.Command{
		Name: "app",
		Flags: []cli.Flag{
			&cli.IntFlag{
				Name:  "port",
				Value: 8000,
			},
			&cli.StringFlag{
				Name:     "database-url",
				Required: true,
				Sources:  cli.EnvVars("DATABASE_URL"),
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			db, err := NewDB(ctx, cmd.String("database-url"))
			if err != nil {
				return fmt.Errorf("open database: %w", err)
			}

			port := cmd.Int("port")
			handler := NewHelloHandler(db)

			mux := http.NewServeMux()
			mux.HandleFunc("GET /", handler.HandleGET)
			mux.HandleFunc("GET /users", handler.GetUsers)

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
					slog.Error("server error", slog.Any("error", err))
					return
				}
			}()

			<-ctx.Done()
			if err := server.Shutdown(ctx); err != nil {
				return fmt.Errorf("shutdown server: %w", err)
			} else {
				slog.InfoContext(ctx, "shutdown server", slog.Int("port", port))
			}
			return nil
		},
	}

	if err := cmd.Run(ctx, os.Args); err != nil {
		slog.Error("command failed", slog.Any("error", err))
		os.Exit(1)
	}
}
