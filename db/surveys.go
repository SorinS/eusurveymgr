package db

import (
	"database/sql"
	"fmt"
)

type SurveyRow struct {
	SurveyID   int64
	Title      string
	Shortname  string
	SurveyUID  string
	Created    sql.NullString
	StartDate  sql.NullString
	EndDate    sql.NullString
	Published  bool
	NumAnswers int
}

func ListSurveys(db *sql.DB) ([]SurveyRow, error) {
	query := `
		SELECT s.SURVEY_ID, s.TITLE, COALESCE(s.SHORTNAME,'') as SHORTNAME,
		       COALESCE(s.SURVEY_UID,'') as SURVEY_UID,
		       s.SURVEY_CREATED, s.SURVEY_START_DATE, s.SURVEY_END_DATE,
		       s.ISPUBLISHED,
		       COUNT(a.ANSWER_SET_ID) as num_answers
		FROM SURVEYS s
		LEFT JOIN ANSWERS_SET a ON a.SURVEY_ID = s.SURVEY_ID
		GROUP BY s.SURVEY_ID, s.TITLE, s.SHORTNAME, s.SURVEY_UID,
		         s.SURVEY_CREATED, s.SURVEY_START_DATE, s.SURVEY_END_DATE, s.ISPUBLISHED
		ORDER BY s.SURVEY_CREATED DESC`

	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("listing surveys: %w", err)
	}
	defer rows.Close()

	var surveys []SurveyRow
	for rows.Next() {
		var s SurveyRow
		if err := rows.Scan(&s.SurveyID, &s.Title, &s.Shortname, &s.SurveyUID,
			&s.Created, &s.StartDate, &s.EndDate, &s.Published, &s.NumAnswers); err != nil {
			return nil, fmt.Errorf("scanning survey row: %w", err)
		}
		surveys = append(surveys, s)
	}
	return surveys, rows.Err()
}