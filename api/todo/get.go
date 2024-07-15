package todo

import (
	"context"

	"github.com/fjnkt98/todo-go/repository"
	"github.com/goark/errs"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type GetResponse struct {
	Items []Item `json:"items"`
}

func GetItems(ctx context.Context, pool *pgxpool.Pool) (*GetResponse, error) {
	q := repository.New(pool)

	items, err := q.GetItems(ctx)
	if err != nil && errs.Is(err, pgx.ErrNoRows) {
		return nil, errs.New("failed to get items", errs.WithCause(err))
	}

	return &GetResponse{
		Items: NewItems(items),
	}, nil
}
