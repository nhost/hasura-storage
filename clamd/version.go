package clamd

import (
	"fmt"
	"strings"
)

type Version struct {
	Version string
}

func parseVersion(response []byte) (Version, error) {
	parts := strings.Split(string(response), " ")
	if len(parts) != 2 { //nolint:gomnd
		return Version{}, fmt.Errorf("failed to parse version: %s", string(response)) //nolint:goerr113
	}

	return Version{
		Version: parts[1],
	}, nil
}

func (c *Client) Version() (Version, error) {
	conn, err := c.Dial()
	if err != nil {
		return Version{}, fmt.Errorf("failed to dial: %w", err)
	}
	defer conn.Close()

	if err := sendCommand(conn, "VERSION"); err != nil {
		return Version{}, fmt.Errorf("failed to send VERSION command: %w", err)
	}

	response, err := readResponse(conn)
	if err != nil {
		return Version{}, fmt.Errorf("failed to read response: %w", err)
	}

	return parseVersion(response)
}
