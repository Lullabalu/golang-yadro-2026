package rest

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"yadro.com/course/api/core"
)

type PingResponse struct {
	Replies map[string]string `json:"replies"`
}

func NewPingHandler(log *slog.Logger, pingers map[string]core.Pinger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		replies := PingResponse{
			Replies: make(map[string]string),
		}

		replies.Replies["other service"] = "unavailable"

		wordsPinger := pingers["words"]

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		err := wordsPinger.Ping(ctx)

		if err == nil {
			replies.Replies["words"] = "ok"
		} else {
			replies.Replies["words"] = "unavailable"
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(replies)

	}
}

type WordsResponse struct {
	Words []string `json:"words"`
	Total int      `json:"total"`
}

func NewWordsHandler(log *slog.Logger, normalizer core.Normalizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		phrase := r.URL.Query().Get("phrase")
		if phrase == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		words, err := normalizer.Norm(r.Context(), phrase)

		if err != nil {
			st, ok := status.FromError(err)

			if ok {
				if st.Code() == codes.ResourceExhausted {
					http.Error(w, "too large", http.StatusBadRequest)
					return
				} else {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
			}
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")

		wordsResponse := WordsResponse{
			Words: words,
			Total: len(words),
		}
		_ = json.NewEncoder(w).Encode(wordsResponse)
	}
}
