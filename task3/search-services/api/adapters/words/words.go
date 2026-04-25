package words

import (
	"context"
	"log/slog"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"
	wordspb "yadro.com/course/proto/words"
)

type Client struct {
	log    *slog.Logger
	client wordspb.WordsClient
}

func NewClient(address string, log *slog.Logger) (*Client, error) {
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		log.Error("Can not create client")
		return nil, err
	}
	client := Client{
		log:    log,
		client: wordspb.NewWordsClient(conn),
	}

	return &client, nil
}

func (c Client) Norm(ctx context.Context, phrase string) ([]string, error) {
	request := wordspb.WordsRequest{
		Phrase: phrase,
	}
	response, err := c.client.Norm(ctx, &request)

	if err != nil {
		return nil, err
	}

	return response.GetWords(), nil
}

func (c Client) Ping(ctx context.Context) error {
	empty := emptypb.Empty{}
	_, err := c.client.Ping(ctx, &empty)
	return err
}
