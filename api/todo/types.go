package todo

import (
	"time"

	"github.com/fjnkt98/todo-go/repository"
)

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
