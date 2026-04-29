package core

import "context"

type Service struct {
	db    DB
	words Words
}

func NewService(
	db DB, words Words) *Service {
	return &Service{
		db:    db,
		words: words,
	}
}

func (s *Service) Search(ctx context.Context, phrase string, limit int) ([]ComicsInfo, error) {
	words, err := s.words.Norm(ctx, phrase)
	if err != nil {
		return nil, err
	}

	if len(words) == 0 {
		return []ComicsInfo{}, nil
	}

	return s.db.Search(ctx, words, limit)
}
