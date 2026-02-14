package db

import (
	"database/sql"
	"fmt"
)

type AnswerSetRow struct {
	AnswerSetID int64
	UniqueCode  string
	Date        sql.NullString
	Email       sql.NullString
}

func ListAnswerSets(db *sql.DB, surveyID int64) ([]AnswerSetRow, error) {
	query := `
		SELECT a_set.ANSWER_SET_ID, a_set.UNIQUECODE, a_set.ANSWER_SET_DATE,
		       a.VALUE as email
		FROM ANSWERS_SET a_set
		LEFT JOIN ANSWERS a ON a.AS_ID = a_set.ANSWER_SET_ID AND a.PA_ID = 0
		WHERE a_set.SURVEY_ID = ?
		ORDER BY a_set.ANSWER_SET_DATE DESC`

	rows, err := db.Query(query, surveyID)
	if err != nil {
		return nil, fmt.Errorf("listing answer sets: %w", err)
	}
	defer rows.Close()

	var answers []AnswerSetRow
	for rows.Next() {
		var a AnswerSetRow
		if err := rows.Scan(&a.AnswerSetID, &a.UniqueCode, &a.Date, &a.Email); err != nil {
			return nil, fmt.Errorf("scanning answer set row: %w", err)
		}
		answers = append(answers, a)
	}
	return answers, rows.Err()
}

func LookupUniqueCode(db *sql.DB, email string, surveyID int64) (int64, string, error) {
	query := `
		SELECT a_set.ANSWER_SET_ID, a_set.UNIQUECODE
		FROM ANSWERS_SET a_set
		JOIN ANSWERS a ON a.AS_ID = a_set.ANSWER_SET_ID
		WHERE a_set.SURVEY_ID = ?
		  AND a.PA_ID = 0
		  AND a.VALUE = ?
		ORDER BY a_set.ANSWER_SET_DATE DESC
		LIMIT 1`

	var answerSetID int64
	var uniqueCode string
	err := db.QueryRow(query, surveyID, email).Scan(&answerSetID, &uniqueCode)
	if err == sql.ErrNoRows {
		return 0, "", fmt.Errorf("no answer set found for email=%q survey=%d", email, surveyID)
	}
	if err != nil {
		return 0, "", fmt.Errorf("looking up uniquecode: %w", err)
	}
	return answerSetID, uniqueCode, nil
}