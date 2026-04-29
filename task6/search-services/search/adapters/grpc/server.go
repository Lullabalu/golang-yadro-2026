package grpc

import (
	"context"

	"google.golang.org/protobuf/types/known/emptypb"
	searchpb "yadro.com/course/proto/search"
	"yadro.com/course/search/core"
)

type Server struct {
	searchpb.UnimplementedSearchServer
	service core.Searcher
}

func NewServer(service core.Searcher) *Server {
	return &Server{service: service}
}

func (s *Server) Ping(ctx context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

func (s *Server) Search(ctx context.Context, req *searchpb.SearchRequest) (*searchpb.SearchResponse, error) {
	limit := req.GetLimit()
	phrase := req.GetPhrase()

	comicsInfos, err := s.service.Search(ctx, phrase, int(limit))
	if err != nil {
		return nil, err
	}

	comics := make([]*searchpb.SearchComic, 0)

	for _, comicInfo := range comicsInfos {
		comics = append(comics, &searchpb.SearchComic{
			Id:  uint64(comicInfo.ID),
			Url: comicInfo.URL,
		})
	}

	response := searchpb.SearchResponse{
		Comics: comics,
		Total:  uint32(len(comics)),
	}

	return &response, nil
}

func (s *Server) SearchIndex(ctx context.Context, req *searchpb.SearchRequest) (*searchpb.SearchResponse, error) {
	limit := req.GetLimit()
	phrase := req.GetPhrase()

	comicsInfos, err := s.service.SearchIndex(ctx, phrase, int(limit))
	if err != nil {
		return nil, err
	}

	comics := make([]*searchpb.SearchComic, 0)

	for _, comicInfo := range comicsInfos {
		comics = append(comics, &searchpb.SearchComic{
			Id:  uint64(comicInfo.ID),
			Url: comicInfo.URL,
		})
	}

	response := searchpb.SearchResponse{
		Comics: comics,
		Total:  uint32(len(comics)),
	}

	return &response, nil
}
