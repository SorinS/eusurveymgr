package client

import (
	"encoding/xml"
	"fmt"
)

// XML response types for /webservice/getSurveys
type SurveyList struct {
	XMLName xml.Name `xml:"surveys"`
	Surveys []Survey `xml:"survey"`
}

type Survey struct {
	Alias       string `xml:"alias" json:"alias"`
	Title       string `xml:"title" json:"title"`
	State       string `xml:"state" json:"state"`
	NumAnswers  int    `xml:"numberOfAnswers" json:"number_of_answers"`
	Start       string `xml:"start" json:"start"`
	End         string `xml:"end" json:"end"`
	CreatedDate string `xml:"createdDate" json:"created_date"`
}

// XML response for /webservice/getSurveyMetadata
type SurveyMetadata struct {
	XMLName    xml.Name `xml:"survey"`
	Alias      string   `xml:"alias"`
	Title      string   `xml:"title"`
	State      string   `xml:"state"`
	NumAnswers int      `xml:"numberOfAnswers"`
	Start      string   `xml:"start"`
	End        string   `xml:"end"`
	Created    string   `xml:"createdDate"`
	Updated    string   `xml:"updatedDate"`
	Published  string   `xml:"publishedDate"`
	Contact    string   `xml:"contact"`
	Language   string   `xml:"language"`
	Security   string   `xml:"security"`
}

func (c *Client) GetSurveys() (*SurveyList, error) {
	data, err := c.doBasicGet("/webservice/getSurveys")
	if err != nil {
		return nil, fmt.Errorf("getSurveys: %w", err)
	}
	var list SurveyList
	if err := xml.Unmarshal(data, &list); err != nil {
		return nil, fmt.Errorf("parsing getSurveys XML: %w", err)
	}
	return &list, nil
}

func (c *Client) GetSurveyMetadata(alias string) (*SurveyMetadata, error) {
	data, err := c.doBasicGet("/webservice/getSurveyMetadata/" + alias)
	if err != nil {
		return nil, fmt.Errorf("getSurveyMetadata: %w", err)
	}
	var meta SurveyMetadata
	if err := xml.Unmarshal(data, &meta); err != nil {
		return nil, fmt.Errorf("parsing getSurveyMetadata XML: %w", err)
	}
	return &meta, nil
}