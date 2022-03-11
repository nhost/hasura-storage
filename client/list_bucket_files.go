package client

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/nhost/hasura-storage/controller"
	"golang.org/x/net/context"
)

func (c *Client) ListBucketFiles(
	ctx context.Context,
	id string,
	filter string,
) (*controller.ListBucketFilesResponse, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/buckets/"+id+"/files/list", nil)
	if err != nil {
		return nil, fmt.Errorf("problem creating request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.jwt)

	q := req.URL.Query()
	q.Add("filter_files_regex", filter)
	req.URL.RawQuery = q.Encode()

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("problem executing request: %w", err)
	}
	defer resp.Body.Close()

	response := &controller.ListBucketFilesResponse{}
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(response); err != nil {
		return nil, fmt.Errorf("problem unmarshaling response: %w", err)
	}

	return response, nil
}
