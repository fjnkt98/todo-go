package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"

	"github.com/fjnkt98/todo-go/api"
	"github.com/fjnkt98/todo-go/api/todo"
	"github.com/fjnkt98/todo-go/repository"
	"github.com/goark/errs"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/urfave/cli/v2"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	app := cli.NewApp()
	app.Flags = []cli.Flag{
		&cli.IntFlag{
			Name:  "port",
			Value: 8000,
		},
		&cli.StringFlag{
			Name:    "database-url",
			EnvVars: []string{"DATABASE_URL"},
		},
	}
	app.Action = func(ctx *cli.Context) error {
		pool, err := repository.NewPool(ctx.Context, ctx.String("database-url"))
		if err != nil {
			return errs.Wrap(err)
		}
		port := ctx.Int("port")

		e := echo.New()
		e.Use(middleware.RequestID())
		e.Use(middleware.Logger())
		e.Use(middleware.Recover())
		e.HideBanner = true
		e.HidePort = true
		e.Validator = new(api.Validator)

		handler := todo.NewHandler(pool)

		e.GET("/api/todo", handler.GET)
		e.POST("/api/todo", handler.POST)
		e.PUT("/api/todo/:id", handler.PUT)

		go func() {
			slog.Info("start server", slog.Int("port", port))
			if err := e.Start(fmt.Sprintf(":%d", port)); err != nil && err != http.ErrServerClosed {
				slog.Error("failed to start server", slog.Int("port", port), slog.Any("error", err))
				panic(fmt.Sprintf("failed to start server: %s", err.Error()))
			}
		}()

		<-ctx.Done()
		slog.Info("shutdown server")
		if err := e.Shutdown(ctx.Context); err != nil {
			return errs.Wrap(err)
		}
		return nil
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	if err := app.RunContext(ctx, os.Args); err != nil {
		slog.Error("command failed", slog.Any("error", err))
		os.Exit(1)
	}
}
