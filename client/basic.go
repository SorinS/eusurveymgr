package client

import (
	"fmt"
	"io"
	"net/http"
)

func (c *Client) doBasicGet(path string) ([]byte, error) {
	body, status, err := c.doBasicGetStatus(path)
	if err != nil {
		return body, err
	}
	_ = status
	return body, nil
}

func (c *Client) doBasicGetStatus(path string) ([]byte, int, error) {
	url := c.BaseURL + path
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, 0, fmt.Errorf("creating request: %w", err)
	}
	req.SetBasicAuth(c.Username, c.Password)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("reading response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return body, resp.StatusCode, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}
	return body, resp.StatusCode, nil
}