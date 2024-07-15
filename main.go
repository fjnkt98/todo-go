package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/fjnkt98/todo-go/repository"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/goark/errs"
	"github.com/jackc/pgx/v5"
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
		&cli.StringFlag{
			Name:    "database-url",
			EnvVars: []string{"DATABASE_URL"},
		},
	}
	app.Action = action

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	if err := app.RunContext(ctx, os.Args); err != nil {
		slog.Error("command failed", slog.Any("error", err))
		os.Exit(1)
	}
}

type GetResponse struct {
	Items   []Item `json:"items"`
	Message string `json:"message,omitempty"`
}

type PostResponse struct {
	Item    Item   `json:"item"`
	Message string `json:"message,omitempty"`
}

type PutResponse struct {
	Item    Item   `json:"item"`
	Message string `json:"message,omitempty"`
}

type Item struct {
	ID        int64     `json:"id"`
	Title     string    `json:"title"`
	UpdatedAt time.Time `json:"updated_at"`
}

func NewItem(item repository.Item) Item {
	return Item{item.ID, item.Title, item.UpdatedAt}
}

func NewItems(items []repository.Item) []Item {
	res := make([]Item, len(items))
	for i, item := range items {
		res[i] = NewItem(item)
	}
	return res
}

type PostParams struct {
	Title string `json:"title"`
}

func (p PostParams) Validate() error {
	return validation.ValidateStruct(
		&p,
		validation.Field(
			&p.Title,
			validation.Required,
			validation.RuneLength(0, 200),
		),
	)
}

type PutParams struct {
	ID    int64  `json:"id"`
	Title string `json:"title"`
}

func (p PutParams) Validate() error {
	return validation.ValidateStruct(
		&p,
		validation.Field(
			&p.Title,
			validation.Required,
			validation.RuneLength(0, 200),
		),
	)
}

func action(ctx *cli.Context) error {
	e := echo.New()
	e.Use(middleware.RequestID())
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.HideBanner = true
	e.HidePort = true
	e.Validator = new(Validator)

	pool, err := repository.NewPool(ctx.Context, ctx.String("database-url"))
	if err != nil {
		return errs.Wrap(err)
	}

	e.GET("/api/todo", func(ctx echo.Context) error {
		q := repository.New(pool)

		items, err := q.FetchItems(ctx.Request().Context())
		if err != nil && errs.Is(err, pgx.ErrNoRows) {
			return &echo.HTTPError{
				Code: http.StatusInternalServerError,
				Message: GetResponse{
					Items:   make([]Item, 0),
					Message: "internal server error",
				},
				Internal: err,
			}
		}
		return ctx.JSON(
			http.StatusOK,
			GetResponse{
				Items: NewItems(items),
			},
		)
	})

	e.POST("api/todo", func(ctx echo.Context) error {
		var p PostParams
		if err := ctx.Bind(&p); err != nil {
			return &echo.HTTPError{
				Code: http.StatusBadRequest,
				Message: PostResponse{
					Message: "bad request",
				},
				Internal: err,
			}
		}
		if err := ctx.Validate(p); err != nil {
			return &echo.HTTPError{
				Code: http.StatusBadRequest,
				Message: PostResponse{
					Message: err.Error(),
				},
				Internal: err,
			}
		}

		tx, err := pool.Begin(ctx.Request().Context())
		if err != nil {
			return &echo.HTTPError{
				Code: http.StatusInternalServerError,
				Message: PostResponse{
					Message: "internal server error",
				},
				Internal: err,
			}
		}
		defer tx.Rollback(ctx.Request().Context())

		q := repository.New(pool).WithTx(tx)
		item, err := q.CreateItem(ctx.Request().Context(), p.Title)
		if err != nil {
			return &echo.HTTPError{
				Code: http.StatusInternalServerError,
				Message: PostResponse{
					Message: "internal server error",
				},
				Internal: err,
			}
		}

		if err := tx.Commit(ctx.Request().Context()); err != nil {
			return &echo.HTTPError{
				Code: http.StatusInternalServerError,
				Message: PostResponse{
					Message: "internal server error",
				},
				Internal: err,
			}
		}

		return ctx.JSON(
			http.StatusOK,
			PostResponse{
				Item: NewItem(item),
			},
		)
	})

	e.PUT("api/todo", func(ctx echo.Context) error {
		var p PutParams
		if err := ctx.Bind(&p); err != nil {
			return &echo.HTTPError{
				Code: http.StatusBadRequest,
				Message: PutResponse{
					Message: "bad request",
				},
				Internal: err,
			}
		}
		if err := ctx.Validate(p); err != nil {
			return &echo.HTTPError{
				Code: http.StatusBadRequest,
				Message: PutResponse{
					Message: err.Error(),
				},
				Internal: err,
			}
		}

		tx, err := pool.Begin(ctx.Request().Context())
		if err != nil {
			return &echo.HTTPError{
				Code: http.StatusInternalServerError,
				Message: PutResponse{
					Message: "internal server error",
				},
				Internal: err,
			}
		}
		defer tx.Rollback(ctx.Request().Context())

		q := repository.New(pool).WithTx(tx)
		item, err := q.UpdateItem(ctx.Request().Context(), repository.UpdateItemParams{ID: p.ID, Title: p.Title})
		if err != nil {
			return &echo.HTTPError{
				Code: http.StatusInternalServerError,
				Message: PutResponse{
					Message: "internal server error",
				},
				Internal: err,
			}
		}

		if err := tx.Commit(ctx.Request().Context()); err != nil {
			return &echo.HTTPError{
				Code: http.StatusInternalServerError,
				Message: PutResponse{
					Message: "internal server error",
				},
				Internal: err,
			}
		}

		return ctx.JSON(
			http.StatusOK,
			PutResponse{
				Item: NewItem(item),
			},
		)
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
