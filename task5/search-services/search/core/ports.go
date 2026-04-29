package core

import (
	"context"
)

type Words interface {
	Norm(ctx context.Context, phrase string) ([]string, error)
}

type DB interface {
	Search(ctx context.Context, words []string, limit int) ([]ComicsInfo, error)
}

type Searcher interface {
	Search(ctx context.Context, phrase string, limit int) ([]ComicsInfo, error)
}
