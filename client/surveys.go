package client

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"unicode/utf8"
)

// XML response types for /webservice/getMySurveys
type SurveyList struct {
	XMLName xml.Name `xml:"Surveys"`
	User    string   `xml:"user,attr" json:"user"`
	Surveys []Survey `xml:"Survey"`
}

type Survey struct {
	UID   string `xml:"uid,attr" json:"uid"`
	Alias string `xml:"alias,attr" json:"alias"`
	Title string `xml:"Title" json:"title"`
}

// XML response for /webservice/getSurveyMetadata/{alias}
type SurveyMetadata struct {
	XMLName    xml.Name `xml:"Survey"`
	ID         string   `xml:"id,attr" json:"id"`
	Alias      string   `xml:"alias,attr" json:"alias"`
	SurveyType string   `xml:"SurveyType" json:"survey_type"`
	Title      string   `xml:"Title" json:"title"`
	Language   string   `xml:"PivotLanguage" json:"language"`
	Contact    string   `xml:"Contact" json:"contact"`
	Status     string   `xml:"Status" json:"status"`
	Start      string   `xml:"Start" json:"start"`
	End        string   `xml:"End" json:"end"`
	Results    int      `xml:"Results" json:"results"`
	Security   string   `xml:"Security" json:"security"`
	Visibility string   `xml:"Visibility" json:"visibility"`
}

// sanitizeXML replaces invalid UTF-8 bytes with the Unicode replacement character.
// EUSurvey sometimes returns Latin-1 data inside UTF-8 declared XML.
func sanitizeXML(data []byte) []byte {
	if utf8.Valid(data) {
		return data
	}
	var buf bytes.Buffer
	buf.Grow(len(data))
	for len(data) > 0 {
		r, size := utf8.DecodeRune(data)
		if r == utf8.RuneError && size == 1 {
			buf.WriteRune('\uFFFD')
		} else {
			buf.WriteRune(r)
		}
		data = data[size:]
	}
	return buf.Bytes()
}

func (c *Client) GetSurveys() (*SurveyList, error) {
	data, err := c.doBasicGet("/webservice/getMySurveys")
	if err != nil {
		return nil, fmt.Errorf("getMySurveys: %w", err)
	}
	var list SurveyList
	if err := xml.Unmarshal(sanitizeXML(data), &list); err != nil {
		return nil, fmt.Errorf("parsing getMySurveys XML: %w", err)
	}
	return &list, nil
}

func (c *Client) GetSurveyMetadata(alias string) (*SurveyMetadata, error) {
	data, err := c.doBasicGet("/webservice/getSurveyMetadata/" + alias)
	if err != nil {
		return nil, fmt.Errorf("getSurveyMetadata: %w", err)
	}
	var meta SurveyMetadata
	if err := xml.Unmarshal(sanitizeXML(data), &meta); err != nil {
		return nil, fmt.Errorf("parsing getSurveyMetadata XML: %w", err)
	}
	return &meta, nil
}