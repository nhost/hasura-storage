package client

import (
	"fmt"
	"net/http"

	"golang.org/x/net/context"
)

// nolint: cyclop
func (c *Client) GetFileByName(
	ctx context.Context, bucketID, filename string, opts ...GetFileInformationOpt,
) (*FileInformationHeaderWithReader, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		"GET",
		fmt.Sprintf("%s/buckets/%s/file", c.baseURL, bucketID),
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("problem creating request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.jwt)

	q := req.URL.Query()
	q.Add("filename", filename)
	req.URL.RawQuery = q.Encode()

	for _, o := range opts {
		o(req)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("problem executing request: %w", err)
	}

	if resp.StatusCode != http.StatusOK &&
		resp.StatusCode != http.StatusNotModified &&
		resp.StatusCode != http.StatusPreconditionFailed {
		defer resp.Body.Close()
		return nil, unmarshalGetFileError(resp)
	}

	info, err := fileInformationFromResponse(resp)
	if err != nil {
		return nil, err
	}

	if info.StatusCode == http.StatusPreconditionFailed || info.StatusCode == http.StatusNotModified {
		defer resp.Body.Close()
		return &FileInformationHeaderWithReader{info, "", nil}, nil
	}

	return &FileInformationHeaderWithReader{
		info,
		filename,
		resp.Body,
	}, nil
}
