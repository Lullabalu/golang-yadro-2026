package update

import (
	"context"
	"log/slog"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"
	"yadro.com/course/api/core"
	updatepb "yadro.com/course/proto/update"
)

type Client struct {
	log    *slog.Logger
	client updatepb.UpdateClient
}

func NewClient(address string, log *slog.Logger) (*Client, error) {
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	return &Client{
		client: updatepb.NewUpdateClient(conn),
		log:    log,
	}, nil
}

func statusToString(s updatepb.Status) core.UpdateStatus {
	switch s {
	case updatepb.Status_STATUS_IDLE:
		return core.StatusUpdateIdle
	case updatepb.Status_STATUS_RUNNING:
		return core.StatusUpdateRunning
	default:
		return core.StatusUpdateUnknown
	}
}

func (c Client) Ping(ctx context.Context) error {
	_, err := c.client.Ping(ctx, &emptypb.Empty{})
	return err
}

func (c Client) Status(ctx context.Context) (core.UpdateStatus, error) {
	resp, err := c.client.Status(ctx, &emptypb.Empty{})
	if err != nil {
		return "", err
	}
	return statusToString(resp.GetStatus()), nil
}

func (c Client) Stats(ctx context.Context) (core.UpdateStats, error) {
	resp, err := c.client.Stats(ctx, &emptypb.Empty{})

	if err != nil {
		return core.UpdateStats{}, err
	}

	updateStats := core.UpdateStats{
		WordsTotal:    int(resp.GetWordsTotal()),
		WordsUnique:   int(resp.GetWordsUnique()),
		ComicsFetched: int(resp.GetComicsFetched()),
		ComicsTotal:   int(resp.GetComicsTotal()),
	}

	return updateStats, nil
}

func (c Client) Update(ctx context.Context) error {
	_, err := c.client.Update(ctx, &emptypb.Empty{})
	return err
}

func (c Client) Drop(ctx context.Context) error {
	_, err := c.client.Drop(ctx, &emptypb.Empty{})

	return err
}
