package words

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	wordspb "yadro.com/course/proto/words"
)

type Client struct {
	client wordspb.WordsClient
}

func NewClient(address string) (*Client, error) {
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		return nil, err
	}

	return &Client{
		client: wordspb.NewWordsClient(conn),
	}, nil
}

func (c *Client) Norm(ctx context.Context, phrase string) ([]string, error) {
	req := wordspb.WordsRequest{
		Phrase: phrase,
	}

	resp, err := c.client.Norm(ctx, &req)
	if err != nil {
		return nil, err
	}

	return resp.GetWords(), nil
}
