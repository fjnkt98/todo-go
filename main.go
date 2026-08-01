package main

import (
	"context"
	"database/sql"
	"log/slog"
	"os"

	_ "github.com/tursodatabase/libsql-client-go/libsql"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	ctx := context.Background()

	db, err := sql.Open("libsql", os.Getenv("DATABASE_URL"))
	if err != nil {
		slog.LogAttrs(ctx, slog.LevelError, "open database", slog.Any("error", err))
		os.Exit(1)
	}
	defer db.Close()

	slog.LogAttrs(ctx, slog.LevelInfo, "todo-go")
}
