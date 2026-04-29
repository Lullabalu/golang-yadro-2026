package core

import (
	"context"
	"fmt"
	"log/slog"
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

func GetById(ctx context.Context, id int, xkcd XKCD) (XKCDInfo, error) {
	if id == 404 {
		return XKCDInfo{
			ID:         404,
			URL:        "",
			Alt:        "",
			SafeTitle:  "",
			Transcript: "",
			Title:      "",
		}, nil
	}

	return xkcd.Get(ctx, id)
}

func worker(ctx context.Context, jobs <-chan int, db DB, words Words, errors chan<- error, xkcd XKCD, wg *sync.WaitGroup, log *slog.Logger) {
	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case id, ok := <-jobs:
			if !ok {
				return
			}

			info, err := GetById(ctx, id, xkcd)

			if err != nil {
				log.Error("Не получилось скачать комикс $1", "id", id, "error", err)
				select {
				case errors <- err:
				case <-ctx.Done():
				}
				return
			}
			curPhrase := info.Alt + " " + info.SafeTitle + " " + info.Title + " " + info.Transcript
			curWords, err := words.Norm(ctx, curPhrase)

			if err != nil {
				log.Error("Сервис words не отработал корректно", "error", err)
				select {
				case errors <- err:
				case <-ctx.Done():
				}
				return
			}

			err = db.Add(ctx, Comics{ID: info.ID, Words: curWords, URL: info.URL})
			if err != nil {
				log.Error("База данных не отработало корректно", "error", err)
				select {
				case errors <- err:
				case <-ctx.Done():
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
		return ErrUpdateAlreadyRunning
	}
	s.status = StatusRunning
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		s.status = StatusIdle
		s.mu.Unlock()
	}()

	count, err := s.xkcd.LastID(ctx)
	if err != nil {
		s.log.Error("Не удалось получить общее число комиксов", "error", err)
		return err
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	jobs := make(chan int, s.concurrency)
	errs := make(chan error, 1)

	var wg sync.WaitGroup

	for i := 0; i < s.concurrency; i++ {
		wg.Add(1)
		go worker(ctx, jobs, s.db, s.words, errs, s.xkcd, &wg, s.log)
	}

	go func() {
		defer close(jobs)

		for id := 1; id <= count; id++ {
			select {
			case jobs <- id:
			case <-ctx.Done():
				return
			}
		}
	}()

	done := make(chan struct{})

	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil

	case err := <-errs:
		cancel()
		<-done
		return err

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
		return ErrUpdateAlreadyRunning
	}
	s.mu.Unlock()

	return s.db.Drop(ctx)
}
