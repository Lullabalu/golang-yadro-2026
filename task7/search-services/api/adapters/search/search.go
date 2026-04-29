package search

import (
	"context"
	"log/slog"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"
	"yadro.com/course/api/core"
	searchpb "yadro.com/course/proto/search"
)

type Client struct {
	log    *slog.Logger
	client searchpb.SearchClient
}

func NewClient(address string, log *slog.Logger) (*Client, error) {
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	return &Client{
		client: searchpb.NewSearchClient(conn),
		log:    log,
	}, nil
}

func (c *Client) Search(ctx context.Context, phrase string, limit int) ([]core.Comics, error) {
	req := searchpb.SearchRequest{
		Phrase: phrase,
		Limit:  uint64(limit),
	}

	resp, err := c.client.Search(ctx, &req)
	if err != nil {
		return nil, err
	}

	comics := make([]core.Comics, 0, len(resp.GetComics()))
	for _, comic := range resp.GetComics() {
		curComic := core.Comics{
			ID:  int(comic.GetId()),
			URL: comic.GetUrl(),
		}
		comics = append(comics, curComic)
	}

	return comics, nil
}

func (c *Client) SearchIndex(ctx context.Context, phrase string, limit int) ([]core.Comics, error) {
	req := searchpb.SearchRequest{
		Phrase: phrase,
		Limit:  uint64(limit),
	}

	resp, err := c.client.SearchIndex(ctx, &req)
	if err != nil {
		return nil, err
	}

	comics := make([]core.Comics, 0, len(resp.GetComics()))
	for _, comic := range resp.GetComics() {
		curComic := core.Comics{
			ID:  int(comic.GetId()),
			URL: comic.GetUrl(),
		}
		comics = append(comics, curComic)
	}

	return comics, nil
}

func (c *Client) Ping(ctx context.Context) error {
	_, err := c.client.Ping(ctx, &emptypb.Empty{})
	return err
}
