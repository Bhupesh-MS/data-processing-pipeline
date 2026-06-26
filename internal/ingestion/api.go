package ingestion

import (
	"context"
	"net/http"

	"data-processing-pipeline/internal/models"
)

type APIClient struct {
	client *http.Client
}

func NewAPIClient(client *http.Client) *APIClient {
	if client == nil {
		client = http.DefaultClient
	}
	return &APIClient{client: client}
}

func (c *APIClient) Fetch(ctx context.Context, endpoint string) ([]models.Record, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return ReadJSON(ctx, resp.Body)
}
