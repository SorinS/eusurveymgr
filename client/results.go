package client

import (
	"fmt"
	"eusurveymgr/log"
	"strings"
	"time"
)

func (c *Client) PrepareResults(formID string, showIDs bool) (string, error) {
	ids := "false"
	if showIDs {
		ids = "true"
	}
	data, err := c.doBasicGet("/webservice/prepareResults/" + formID + "/" + ids)
	if err != nil {
		return "", fmt.Errorf("prepareResults: %w", err)
	}
	taskID := strings.TrimSpace(string(data))
	if taskID == "" {
		return "", fmt.Errorf("prepareResults returned empty task ID")
	}
	return taskID, nil
}

func (c *Client) GetResults(taskID string, timeoutSeconds int) ([]byte, error) {
	deadline := time.Now().Add(time.Duration(timeoutSeconds) * time.Second)
	delay := 1 * time.Second

	for {
		data, err := c.doBasicGet("/webservice/getResults/" + taskID)
		if err != nil {
			if time.Now().After(deadline) {
				return nil, fmt.Errorf("getResults timed out after %ds: %w", timeoutSeconds, err)
			}
			log.Debugf("Results not ready yet, retrying in %v...", delay)
			time.Sleep(delay)
			if delay < 5*time.Second {
				delay += time.Second
			}
			continue
		}
		return data, nil
	}
}