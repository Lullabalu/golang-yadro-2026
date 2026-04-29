package core

import (
	"context"
)

type Words interface {
	Norm(ctx context.Context, phrase string) ([]string, error)
}

type DB interface {
	Search(ctx context.Context, words []string, limit int) ([]ComicsInfo, error)
	GetIndexRows(ctx context.Context) ([]IndexRow, error)
	Get(ctx context.Context, id int) (ComicsInfo, error)
}

type Searcher interface {
	Search(ctx context.Context, phrase string, limit int) ([]ComicsInfo, error)
	SearchIndex(ctx context.Context, phrase string, limit int) ([]ComicsInfo, error)
	UpdateIndex() error
}
