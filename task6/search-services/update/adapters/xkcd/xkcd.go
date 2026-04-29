package xkcd

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"yadro.com/course/update/core"
)

type Client struct {
	log    *slog.Logger
	client http.Client
	url    string
}

func NewClient(url string, timeout time.Duration, log *slog.Logger) (*Client, error) {
	if url == "" {
		return nil, fmt.Errorf("empty base url specified")
	}
	return &Client{
		client: http.Client{Timeout: timeout},
		log:    log,
		url:    url,
	}, nil
}

type Response struct {
	Id         int    `json:"num"`
	Url        string `json:"img"`
	SafeTitle  string `json:"safe_title"`
	Transcript string `json:"transcript"`
	Alt        string `json:"alt"`
	Title      string `json:"title"`
}

func (c *Client) Get(ctx context.Context, id int) (core.XKCDInfo, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/%d/info.0.json", c.url, id), nil)

	if err != nil {
		return core.XKCDInfo{}, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return core.XKCDInfo{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return core.XKCDInfo{}, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var response Response

	err = json.NewDecoder(resp.Body).Decode(&response)

	return core.XKCDInfo{
		ID:         response.Id,
		URL:        response.Url,
		Title:      response.Title,
		SafeTitle:  response.SafeTitle,
		Alt:        response.Alt,
		Transcript: response.Transcript,
	}, err

}

func (c *Client) LastID(ctx context.Context) (int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/info.0.json", c.url), nil)

	if err != nil {
		return 0, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var response Response
	err = json.NewDecoder(resp.Body).Decode(&response)

	return response.Id, err
}
