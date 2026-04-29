package rest

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"yadro.com/course/api/core"
)

type PingResponse struct {
	Replies map[string]string `json:"replies"`
}

func NewPingHandler(log *slog.Logger, pingers map[string]core.Pinger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		response := PingResponse{
			Replies: make(map[string]string, len(pingers)),
		}

		for service, pinger := range pingers {
			err := pinger.Ping(r.Context())
			if err == nil {
				response.Replies[service] = "ok"
			} else {
				response.Replies[service] = "unavailable"
			}
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(&response)
	}
}

func NewUpdateHandler(log *slog.Logger, updater core.Updater) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := updater.Update(r.Context())

		if err != nil {
			if status.Code(err) == codes.AlreadyExists {
				w.WriteHeader(http.StatusAccepted)
				return
			}
			log.Error("update failed", "err", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

type StatsResponse struct {
	WordsTotal    int `json:"words_total"`
	WordsUnique   int `json:"words_unique"`
	ComicsFetched int `json:"comics_fetched"`
	ComicsTotal   int `json:"comics_total"`
}

func NewUpdateStatsHandler(log *slog.Logger, updater core.Updater) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		stats, err := updater.Stats(context.Background())

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		resp := StatsResponse{
			WordsTotal:    stats.WordsTotal,
			WordsUnique:   stats.WordsUnique,
			ComicsFetched: stats.ComicsFetched,
			ComicsTotal:   stats.ComicsTotal,
		}

		_ = json.NewEncoder(w).Encode(&resp)
	}
}

type StatusResponse struct {
	Status core.UpdateStatus `json:"status"`
}

func NewUpdateStatusHandler(log *slog.Logger, updater core.Updater) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		status, err := updater.Status(context.Background())

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		resp := StatusResponse{Status: status}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(&resp)
	}
}

func NewDropHandler(log *slog.Logger, updater core.Updater) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := updater.Drop(context.Background())

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

type SearchResponse struct {
	Comics []core.Comics `json:"comics"`
	Total  int           `json:"total"`
}

func GetLimit(limits string) (int, error) {
	if limits == "" {
		return 10, nil
	}

	limit, err := strconv.Atoi(limits)
	if err != nil {
		return 0, err
	}

	if limit <= 0 {
		return 0, errors.New("limit must be positive")
	}

	return limit, nil
}

func NewSearchHandler(log *slog.Logger, searcher core.Searcher) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		phrase := strings.TrimSpace(r.URL.Query().Get("phrase"))
		limits := r.URL.Query().Get("limit")

		limit, err := GetLimit(limits)
		if phrase == "" || err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		comics, err := searcher.Search(r.Context(), phrase, limit)
		if err != nil {
			log.Error("search failed", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		resp := SearchResponse{
			Comics: comics,
			Total:  len(comics),
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(resp)
	}
}
func NewSearchIndexHandler(log *slog.Logger, searcher core.Searcher) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		phrase := strings.TrimSpace(r.URL.Query().Get("phrase"))
		limits := r.URL.Query().Get("limit")

		limit, err := GetLimit(limits)
		if phrase == "" || err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		comics, err := searcher.SearchIndex(r.Context(), phrase, limit)
		if err != nil {
			log.Error("search failed", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		resp := SearchResponse{
			Comics: comics,
			Total:  len(comics),
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(resp)
	}
}

type AuthData struct {
	Name     string `json:"name"`
	Password string `json:"password"`
}

func NewLoginHandler(log *slog.Logger, authenticator core.Authenticator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var authData AuthData
		defer r.Body.Close()
		err := json.NewDecoder(r.Body).Decode(&authData)

		if err != nil {
			log.Warn("login failed")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		token, err := authenticator.Login(authData.Name, authData.Password)

		if err != nil {
			log.Warn("login failed", "user", authData.Name)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "text/plain")
		_, _ = w.Write([]byte(token))
	}
}
