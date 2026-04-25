package core

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"
)

type Service struct {
	log         *slog.Logger
	db          DB
	xkcd        XKCD
	words       Words
	concurrency int
	mu          sync.Mutex
	status      ServiceStatus
}

func NewService(
	log *slog.Logger, db DB, xkcd XKCD, words Words, concurrency int,
) (*Service, error) {
	if concurrency < 1 {
		return nil, fmt.Errorf("wrong concurrency specified: %d", concurrency)
	}
	return &Service{
		log:         log,
		db:          db,
		xkcd:        xkcd,
		words:       words,
		concurrency: concurrency,
		status:      StatusIdle,
	}, nil
}

func (s *Service) worker(ctx context.Context, jobs <-chan int, errCh chan<- error, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case id, ok := <-jobs:
			if !ok {
				return
			}

			info, err := s.xkcd.Get(ctx, id)
			if err != nil {
				if strings.Contains(err.Error(), "404") {
					comic := Comics{
						ID:    id,
						URL:   "",
						Words: nil,
					}
					if err := s.db.Add(ctx, comic); err != nil {
						select {
						case errCh <- err:
						default:
						}
						return
					}

					continue
				}

				select {
				case errCh <- err:
				default:
				}
				return
			}

			phrase := info.Title + " " + info.Description
			words, err := s.words.Norm(ctx, phrase)
			if err != nil {
				select {
				case errCh <- err:
				default:
				}
				return
			}

			comic := Comics{
				ID:    info.ID,
				URL:   info.URL,
				Words: words,
			}

			if err := s.db.Add(ctx, comic); err != nil {
				select {
				case errCh <- err:
				default:
				}
				return
			}
		}
	}
}

func (s *Service) Update(ctx context.Context) error {
	s.mu.Lock()
	if s.status == StatusRunning {
		s.mu.Unlock()
		return AMOGUS
	}
	s.status = StatusRunning
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		s.status = StatusIdle
		s.mu.Unlock()
	}()

	ids, err := s.db.IDs(ctx)
	if err != nil {
		return err
	}

	used := make(map[int]struct{}, len(ids))
	for _, id := range ids {
		used[id] = struct{}{}
	}

	total, err := s.xkcd.LastID(ctx)
	if err != nil {
		return err
	}

	updateCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	jobs := make(chan int)
	errCh := make(chan error, 1)

	var wg sync.WaitGroup
	for i := 0; i < s.concurrency; i++ {
		wg.Add(1)
		go s.worker(updateCtx, jobs, errCh, &wg)
	}

	go func() {
		defer close(jobs)
		for num := 1; num <= total; num++ {
			if _, ok := used[num]; ok {
				continue
			}

			select {
			case <-updateCtx.Done():
				return
			case jobs <- num:
			}
		}
	}()

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case err := <-errCh:
		cancel()
		<-done
		return err
	case <-done:
		return nil
	case <-ctx.Done():
		cancel()
		<-done
		return ctx.Err()
	}
}

func (s *Service) Stats(ctx context.Context) (ServiceStats, error) {
	dbstats, err := s.db.Stats(ctx)
	if err != nil {
		return ServiceStats{}, err
	}

	comicsTotal, err := s.xkcd.LastID(ctx)
	if err != nil {
		return ServiceStats{}, err
	}

	return ServiceStats{
		ComicsTotal: comicsTotal,
		DBStats:     dbstats,
	}, nil
}

func (s *Service) Status(_ context.Context) ServiceStatus {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.status
}

func (s *Service) Drop(ctx context.Context) error {
	s.mu.Lock()
	if s.status == StatusRunning {
		s.mu.Unlock()
		return AMOGUS
	}
	s.mu.Unlock()

	return s.db.Drop(ctx)
}
