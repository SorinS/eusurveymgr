package client

import (
	"crypto/tls"
	"eusurveymgr/config"
	"net/http"
	"net/http/cookiejar"
	"time"
)

type Client struct {
	BaseURL     string
	Username    string
	Password    string
	HTTPClient  *http.Client
	loggedIn    bool
}

func New(cfg *config.Configuration) *Client {
	jar, _ := cookiejar.New(nil)
	transport := &http.Transport{}
	if cfg.InsecureTLS {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}
	return &Client{
		BaseURL:  cfg.BaseURL,
		Username: cfg.WebUser,
		Password: cfg.WebPassword,
		HTTPClient: &http.Client{
			Jar:       jar,
			Timeout:   time.Duration(cfg.TimeoutSeconds) * time.Second,
			Transport: transport,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
	}
}