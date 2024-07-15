package todo

import (
	"context"

	"github.com/fjnkt98/todo-go/repository"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/goark/errs"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PutResponse struct {
	Item Item `json:"item"`
}

type PutParams struct {
	ID    int64  `param:"id"`
	Title string `json:"title"`
}

func (p PutParams) Validate() error {
	return validation.ValidateStruct(
		&p,
		validation.Field(
			&p.ID,
			validation.Required,
		),
		validation.Field(
			&p.Title,
			validation.Required,
			validation.RuneLength(0, 200),
		),
	)
}

func UpdateItem(ctx context.Context, pool *pgxpool.Pool, p PutParams) (*PutResponse, error) {
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
	item, err := q.UpdateItem(ctx, repository.UpdateItemParams{ID: p.ID, Title: p.Title})
	if err != nil {
		return nil, errs.New(
			"failed to update item",
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

	return &PutResponse{
		Item: NewItem(item),
	}, nil
}
