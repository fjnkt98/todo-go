package main

import (
	"context"
	"database/sql"
	"log/slog"
	"os"

	_ "github.com/tursodatabase/go-libsql"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	ctx := context.Background()

	db, err := sql.Open("libsql", os.Getenv("DATABASE_URL"))
	if err != nil {
		slog.LogAttrs(ctx, slog.LevelError, "open database", slog.Any("error", err))
		return
	}
	defer db.Close() //nolint: errcheck

	slog.LogAttrs(ctx, slog.LevelInfo, "todo-go")

	rows, err := db.QueryContext(ctx, "SELECT 1;")
	if err != nil {
		slog.LogAttrs(ctx, slog.LevelError, "fetch rows", slog.Any("error", err))
		return
	}
	defer rows.Close() //nolint: errcheck

	for rows.Next() {
		var num int
		if err := rows.Scan(&num); err != nil {
			slog.LogAttrs(ctx, slog.LevelError, "fetch rows", slog.Any("error", err))
			return
		}
		slog.LogAttrs(ctx, slog.LevelInfo, "row fetched", slog.Int("value", num))
	}

	if err := rows.Err(); err != nil {
		slog.LogAttrs(ctx, slog.LevelError, "fetch rows", slog.Any("error", err))
		return
	}
}
