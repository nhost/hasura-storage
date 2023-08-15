package clamd

import (
	"fmt"
	"regexp"
)

type Version struct {
	Version string
	Date    string
}

func parseVersion(response []byte) (Version, error) {
	// parse ClamAV 1.1.0/26984/Sat Jul 29 07:26:39 2023
	regex := regexp.MustCompile(`^ClamAV (.*\/.*)\/(.*)\n$`)
	match := regex.FindStringSubmatch(string(response))
	if len(match) == 3 { //nolint:gomnd
		return Version{
			Version: match[1],
			Date:    match[2],
		}, nil
	}

	return Version{}, fmt.Errorf("failed to parse version: %s", string(response)) //nolint:goerr113
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
