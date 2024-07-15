package todo

import (
	"context"

	"github.com/fjnkt98/todo-go/repository"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/goark/errs"
	"github.com/jackc/pgx/v5/pgxpool"
)

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

type PostResponse struct {
	Item Item `json:"item"`
}

func CreateItem(ctx context.Context, pool *pgxpool.Pool, p PostParams) (*PostResponse, error) {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return nil, errs.New(
			"failed to begin transaction",
			errs.WithCause(err),
			errs.WithContext("params", p),
		)
	}
	defer tx.Rollback(ctx)

	q := repository.New(pool).WithTx(tx)
	item, err := q.CreateItem(ctx, p.Title)
	if err != nil {
		return nil, errs.New(
			"failed to create item",
			errs.WithCause(err),
			errs.WithContext("params", p),
		)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, errs.New(
			"failed to commit the transaction",
			errs.WithCause(err),
			errs.WithContext("params", p),
		)
	}

	return &PostResponse{
		Item: NewItem(item),
	}, nil
}
