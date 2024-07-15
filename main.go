package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/goark/errs"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/urfave/cli/v2"
)

type Validator struct{}

func (v *Validator) Validate(i any) error {
	if c, ok := i.(validation.Validatable); ok {
		return c.Validate()
	}
	return nil
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	app := cli.NewApp()
	app.Flags = []cli.Flag{
		&cli.IntFlag{
			Name:  "port",
			Value: 8000,
		},
	}
	app.Action = func(ctx *cli.Context) error {
		e := echo.New()
		e.Use(middleware.RequestID())
		e.Use(middleware.Logger())
		e.Use(middleware.Recover())
		e.HideBanner = true
		e.HidePort = true
		e.Validator = new(Validator)

		e.GET("/api/todo", func(ctx echo.Context) error {
			body := []map[string]any{
				{
					"id":    1,
					"title": "first todo",
				},
				{
					"id":    2,
					"title": "second todo",
				},
			}
			return ctx.JSON(http.StatusOK, body)
		})

		port := ctx.Int("port")

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
