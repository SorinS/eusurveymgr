package client

import (
	"eusurveymgr/log"
	"fmt"
	"io"
	"net/http"
)

// GetSurveyPDF downloads the survey form PDF via Basic Auth.
func (c *Client) GetSurveyPDF(alias string) ([]byte, error) {
	data, err := c.doBasicGet("/webservice/getSurveyPDF/" + alias)
	if err != nil {
		return nil, fmt.Errorf("getSurveyPDF: %w", err)
	}
	return data, nil
}

// CreateAnswerPDF triggers server-side PDF generation for an answer.
// Requires prior Login().
func (c *Client) CreateAnswerPDF(uniqueCode string) error {
	if err := c.Login(); err != nil {
		return err
	}

	url := c.BaseURL + "/worker/createanswerpdf/" + uniqueCode
	resp, err := c.HTTPClient.Do(mustNewRequest("GET", url))
	if err != nil {
		return fmt.Errorf("createanswerpdf: %w", err)
	}
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()

	result := string(body)
	if result != "OK" {
		return fmt.Errorf("createanswerpdf returned %q (expected OK)", result)
	}
	log.Infof("PDF generation triggered for %s", uniqueCode)
	return nil
}

// DownloadAnswerPDF downloads a previously generated answer PDF.
// Requires prior Login().
func (c *Client) DownloadAnswerPDF(uniqueCode string) ([]byte, error) {
	if err := c.Login(); err != nil {
		return nil, err
	}

	url := c.BaseURL + "/pdf/answer/" + uniqueCode
	req, _ := http.NewRequest("GET", url, nil)
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("download answer PDF: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("download answer PDF: HTTP %d: %s", resp.StatusCode, string(body))
	}
	return io.ReadAll(resp.Body)
}

// IsAnswerPDFReady checks if a PDF has been generated.
func (c *Client) IsAnswerPDFReady(uniqueCode string) (bool, error) {
	if err := c.Login(); err != nil {
		return false, err
	}

	url := c.BaseURL + "/pdf/answerready/" + uniqueCode
	resp, err := c.HTTPClient.Do(mustNewRequest("GET", url))
	if err != nil {
		return false, fmt.Errorf("answerready: %w", err)
	}
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()

	result := string(body)
	return result == "exists" || result == "OK", nil
}