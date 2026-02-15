package db

import (
	"database/sql"
	"fmt"
)

type AnswerSetRow struct {
	AnswerSetID int64
	UniqueCode  string
	Date        sql.NullString
	Name        sql.NullString
	Email       sql.NullString
}

func ListAnswerSets(db *sql.DB, surveyID int64) ([]AnswerSetRow, error) {
	// PA_ID=0 has two rows per answer set: name (first inserted) and email (second).
	// We use MIN/MAX on ANSWER_ID to reliably distinguish them.
	query := `
		SELECT a_set.ANSWER_SET_ID, a_set.UNIQUECODE, a_set.ANSWER_SET_DATE,
		       a_name.VALUE as name, a_email.VALUE as email
		FROM ANSWERS_SET a_set
		LEFT JOIN ANSWERS a_name ON a_name.AS_ID = a_set.ANSWER_SET_ID
		    AND a_name.ANSWER_ID = (SELECT MIN(ANSWER_ID) FROM ANSWERS WHERE AS_ID = a_set.ANSWER_SET_ID AND PA_ID = 0)
		LEFT JOIN ANSWERS a_email ON a_email.AS_ID = a_set.ANSWER_SET_ID
		    AND a_email.ANSWER_ID = (SELECT MAX(ANSWER_ID) FROM ANSWERS WHERE AS_ID = a_set.ANSWER_SET_ID AND PA_ID = 0)
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
		if err := rows.Scan(&a.AnswerSetID, &a.UniqueCode, &a.Date, &a.Name, &a.Email); err != nil {
			return nil, fmt.Errorf("scanning answer set row: %w", err)
		}
		answers = append(answers, a)
	}
	return answers, rows.Err()
}

type ResponseRow struct {
	PA_ID    int
	Question sql.NullString
	Value    sql.NullString
}

func GetResponses(db *sql.DB, answerSetID int64) ([]ResponseRow, error) {
	query := `
		SELECT a.PA_ID, e.ETITLE as question, a.VALUE
		FROM ANSWERS a
		LEFT JOIN ELEMENTS e ON e.ELEM_UID = a.PA_UID
		WHERE a.AS_ID = ?
		ORDER BY a.PA_ID, a.ANSWER_ID`

	rows, err := db.Query(query, answerSetID)
	if err != nil {
		return nil, fmt.Errorf("getting responses: %w", err)
	}
	defer rows.Close()

	var responses []ResponseRow
	for rows.Next() {
		var r ResponseRow
		if err := rows.Scan(&r.PA_ID, &r.Question, &r.Value); err != nil {
			return nil, fmt.Errorf("scanning response row: %w", err)
		}
		responses = append(responses, r)
	}
	return responses, rows.Err()
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