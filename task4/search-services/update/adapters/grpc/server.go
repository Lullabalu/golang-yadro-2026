package grpc

import (
	"context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	updatepb "yadro.com/course/proto/update"
	"yadro.com/course/update/core"
)

func NewServer(service core.Updater) *Server {
	return &Server{service: service}
}

type Server struct {
	updatepb.UnimplementedUpdateServer
	service core.Updater
}

func (s *Server) Ping(ctx context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

func toPBStatus(status core.ServiceStatus) updatepb.Status {
	switch status {
	case core.StatusIdle:
		return updatepb.Status_STATUS_IDLE
	case core.StatusRunning:
		return updatepb.Status_STATUS_RUNNING
	default:
		return updatepb.Status_STATUS_UNSPECIFIED
	}
}

func (s *Server) Status(ctx context.Context, _ *emptypb.Empty) (*updatepb.StatusReply, error) {
	status := s.service.Status(ctx)

	return &updatepb.StatusReply{
		Status: toPBStatus(status),
	}, nil
}

func (s *Server) Update(ctx context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
	err := s.service.Update(ctx)
	if err == core.ErrUpdateAlreadyRunning {
		return nil, status.Error(codes.AlreadyExists, "Обновление уже запущено")
	}
	return &emptypb.Empty{}, err
}

func (s *Server) Stats(ctx context.Context, _ *emptypb.Empty) (*updatepb.StatsReply, error) {
	stats, err := s.service.Stats(ctx)

	if err != nil {
		return nil, err
	}

	statsReply := updatepb.StatsReply{
		WordsTotal:    int64(stats.WordsTotal),
		WordsUnique:   int64(stats.WordsUnique),
		ComicsTotal:   int64(stats.ComicsTotal),
		ComicsFetched: int64(stats.ComicsFetched),
	}

	return &statsReply, nil
}

func (s *Server) Drop(ctx context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
	err := s.service.Drop(ctx)
	return &emptypb.Empty{}, err
}
