package core

import (
	"context"
	"sort"
	"sync"
)

type Service struct {
	db    DB
	words Words

	mu    sync.RWMutex
	Index map[string][]ComicsInfo
}

func NewService(
	db DB, words Words) *Service {
	return &Service{
		db:    db,
		words: words,
		mu:    sync.RWMutex{},
		Index: make(map[string][]ComicsInfo),
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

func (s *Service) UpdateIndex() error {
	indexRows, err := s.db.GetIndexRows(context.Background())
	if err != nil {
		return err
	}

	newIndex := make(map[string][]ComicsInfo)

	for _, row := range indexRows {
		newIndex[row.Word] = append(newIndex[row.Word], ComicsInfo{
			ID:  row.ComicID,
			URL: row.URL,
		})
	}

	s.mu.Lock()
	s.Index = newIndex
	s.mu.Unlock()

	return nil
}

type SearchResult struct {
	ID  int
	URL string

	Matches int
}

func (s *Service) SearchIndex(ctx context.Context, phrase string, limit int) ([]ComicsInfo, error) {
	matches := make(map[int]SearchResult)
	words, err := s.words.Norm(ctx, phrase)

	if err != nil {
		return nil, err
	}

	s.mu.RLock()
	for _, word := range words {
		for _, comicsInfo := range s.Index[word] {
			res, ok := matches[comicsInfo.ID]
			if !ok {
				res = SearchResult{
					ID:  comicsInfo.ID,
					URL: comicsInfo.URL,
				}
			}
			res.Matches++
			matches[comicsInfo.ID] = res
		}
	}
	s.mu.RUnlock()

	results := make([]SearchResult, 0, len(matches))
	for _, ma := range matches {
		results = append(results, ma)
	}

	sort.Slice(results, func(i, j int) bool {
		if results[i].Matches != results[j].Matches {
			return results[i].Matches > results[j].Matches
		}

		return results[i].ID < results[j].ID
	})

	comicsInfos := make([]ComicsInfo, 0)
	for i := 0; i < min(len(results), limit); i++ {
		comicsInfos = append(comicsInfos, ComicsInfo{ID: results[i].ID, URL: results[i].URL})
	}

	return comicsInfos, nil
}
