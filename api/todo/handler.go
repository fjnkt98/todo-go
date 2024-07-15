package todo

import (
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
)

type Handler struct {
	pool *pgxpool.Pool
}

func NewHandler(pool *pgxpool.Pool) *Handler {
	return &Handler{
		pool: pool,
	}
}

func (h *Handler) GET(ctx echo.Context) error {
	res, err := GetItems(ctx.Request().Context(), h.pool)
	if err != nil {
		return &echo.HTTPError{
			Code:     http.StatusInternalServerError,
			Message:  "internal server error",
			Internal: err,
		}
	}

	return ctx.JSON(http.StatusOK, res)
}

func (h *Handler) POST(ctx echo.Context) error {
	var p PostParams
	if err := ctx.Bind(&p); err != nil {
		return &echo.HTTPError{
			Code:     http.StatusBadRequest,
			Message:  "bad request",
			Internal: err,
		}
	}

	if err := ctx.Validate(p); err != nil {
		return &echo.HTTPError{
			Code:     http.StatusBadRequest,
			Message:  err.Error(),
			Internal: err,
		}
	}

	res, err := CreateItem(ctx.Request().Context(), h.pool, p)
	if err != nil {
		return &echo.HTTPError{
			Code:     http.StatusInternalServerError,
			Message:  "internal server error",
			Internal: err,
		}
	}

	return ctx.JSON(http.StatusOK, res)
}

func (h *Handler) PUT(ctx echo.Context) error {
	var p PutParams
	if err := ctx.Bind(&p); err != nil {
		return &echo.HTTPError{
			Code:     http.StatusBadRequest,
			Message:  "bad request",
			Internal: err,
		}
	}

	if err := ctx.Validate(p); err != nil {
		return &echo.HTTPError{
			Code:     http.StatusBadRequest,
			Message:  err.Error(),
			Internal: err,
		}
	}

	res, err := UpdateItem(ctx.Request().Context(), h.pool, p)
	if err != nil {
		return &echo.HTTPError{
			Code:     http.StatusInternalServerError,
			Message:  "internal server error",
			Internal: err,
		}
	}

	return ctx.JSON(http.StatusOK, res)
}
