package main

import (
	"context"
	"database/sql"
	"log/slog"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	ctx := context.Background()

	db, err := sql.Open("sqlite3", os.Getenv("DATABASE_URL"))
	if err != nil {
		slog.LogAttrs(ctx, slog.LevelError, "open database", slog.Any("error", err))
		return
	}
	defer db.Close() //nolint: errcheck

	if _, err := db.ExecContext(ctx, "PRAGMA journal_mode = WAL;"); err != nil {
		slog.LogAttrs(ctx, slog.LevelError, "pragma journal mode", slog.Any("error", err))
		return
	}
	if _, err := db.ExecContext(ctx, "PRAGMA foreign_keys = on;"); err != nil {
		slog.LogAttrs(ctx, slog.LevelError, "pragma foreign keys", slog.Any("error", err))
		return
	}
	if _, err := db.ExecContext(ctx, "PRAGMA defer_foreign_keys = on;"); err != nil {
		slog.LogAttrs(ctx, slog.LevelError, "pragma defer foreign keys", slog.Any("error", err))
		return
	}

	slog.LogAttrs(ctx, slog.LevelInfo, "todo-go")

	rows, err := db.QueryContext(ctx, "SELECT id, name FROM users;")
	if err != nil {
		slog.LogAttrs(ctx, slog.LevelError, "fetch rows", slog.Any("error", err))
		return
	}
	defer rows.Close() //nolint: errcheck

	for rows.Next() {
		var id int
		var name string
		if err := rows.Scan(&id, &name); err != nil {
			slog.LogAttrs(ctx, slog.LevelError, "fetch rows", slog.Any("error", err))
			return
		}
		slog.LogAttrs(ctx, slog.LevelInfo, "row fetched", slog.Int("id", id), slog.String("name", name))
	}

	if err := rows.Err(); err != nil {
		slog.LogAttrs(ctx, slog.LevelError, "fetch rows", slog.Any("error", err))
		return
	}
}
