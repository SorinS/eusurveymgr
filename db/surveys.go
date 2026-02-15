package db

import (
	"database/sql"
	"fmt"
)

type SurveyRow struct {
	SurveyID  int64
	Title     string
	Alias     string
	SurveyUID string
	Created   sql.NullString
	Published bool
	NumAnswers int
}

func ListSurveys(db *sql.DB) ([]SurveyRow, error) {
	// Only show the latest version of each survey (max SURVEY_ID per SURVEY_UID)
	query := `
		SELECT s.SURVEY_ID, COALESCE(s.TITLE,'') as TITLE,
		       s.SURVEYNAME as ALIAS, s.SURVEY_UID,
		       s.SURVEY_CREATED, COALESCE(s.ISPUBLISHED, 0),
		       (SELECT COUNT(*) FROM ANSWERS_SET a WHERE a.SURVEY_ID = s.SURVEY_ID) as num_answers
		FROM SURVEYS s
		WHERE s.SURVEY_ID = (
		    SELECT MAX(s2.SURVEY_ID) FROM SURVEYS s2 WHERE s2.SURVEY_UID = s.SURVEY_UID
		)
		ORDER BY s.SURVEY_CREATED DESC`

	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("listing surveys: %w", err)
	}
	defer rows.Close()

	var surveys []SurveyRow
	for rows.Next() {
		var s SurveyRow
		if err := rows.Scan(&s.SurveyID, &s.Title, &s.Alias, &s.SurveyUID,
			&s.Created, &s.Published, &s.NumAnswers); err != nil {
			return nil, fmt.Errorf("scanning survey row: %w", err)
		}
		surveys = append(surveys, s)
	}
	return surveys, rows.Err()
}