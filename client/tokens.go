package client

import (
	"encoding/xml"
	"fmt"
)

type TokenList struct {
	XMLName xml.Name `xml:"tokens"`
	Tokens  []Token  `xml:"token"`
}

type Token struct {
	Value string `xml:"value" json:"value"`
}

func (c *Client) GetTokens(surveyName string) (*TokenList, error) {
	data, err := c.doBasicGet("/webservice/getTokens/" + surveyName)
	if err != nil {
		return nil, fmt.Errorf("getTokens: %w", err)
	}
	var list TokenList
	if err := xml.Unmarshal(data, &list); err != nil {
		return nil, fmt.Errorf("parsing getTokens XML: %w", err)
	}
	return &list, nil
}

func (c *Client) CreateToken(surveyName string) (string, error) {
	data, err := c.doBasicGet("/webservice/createToken/" + surveyName)
	if err != nil {
		return "", fmt.Errorf("createToken: %w", err)
	}
	return string(data), nil
}