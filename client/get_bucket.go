package client

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/nhost/hasura-storage/controller"
	"golang.org/x/net/context"
)

func (c *Client) GetBucket(
	ctx context.Context,
	id string,
) (*controller.GetBucketResponse, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/buckets/"+id, nil)
	if err != nil {
		return nil, fmt.Errorf("problem creating request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.jwt)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("problem executing request: %w", err)
	}
	defer resp.Body.Close()

	response := &controller.GetBucketResponse{}
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(response); err != nil {
		return nil, fmt.Errorf("problem unmarshaling response: %w", err)
	}

	return response, nil
}