package client

import (
	"eusurveymgr/log"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

var csrfRe = regexp.MustCompile(`<meta\s+name="_csrf"\s+content="([^"]+)"`)

func (c *Client) Login() error {
	if c.loggedIn {
		return nil
	}

	// Step 1: GET /auth/login to obtain CSRF token
	loginURL := c.BaseURL + "/auth/login"
	resp, err := c.HTTPClient.Do(mustNewRequest("GET", loginURL))
	if err != nil {
		return fmt.Errorf("fetching login page: %w", err)
	}
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()

	matches := csrfRe.FindSubmatch(body)
	if matches == nil {
		return fmt.Errorf("CSRF token not found in login page")
	}
	csrf := string(matches[1])
	log.Debugf("CSRF token: %s", csrf)

	// Step 2: POST /login with credentials
	form := url.Values{
		"username": {c.Username},
		"password": {c.Password},
		"_csrf":    {csrf},
	}
	req, _ := http.NewRequest("POST", c.BaseURL+"/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err = c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("login POST: %w", err)
	}
	io.ReadAll(resp.Body)
	resp.Body.Close()

	if resp.StatusCode != http.StatusFound && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("login failed: HTTP %d", resp.StatusCode)
	}

	c.loggedIn = true
	log.Infof("Logged in to EUSurvey")
	return nil
}

func mustNewRequest(method, url string) *http.Request {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		panic(err)
	}
	return req
}