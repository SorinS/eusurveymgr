# EUSurvey Database Reference

## Connection Details

- **Host**: 127.0.0.1 (localhost, or via SSH port forward)
- **Port**: 3306
- **Database**: eusurveydb
- **User**: reportr
- **Password**: YE65MoSbiVAT7vnz
- **Engine**: MySQL 8.0

## Key Tables

### SURVEYS
Main survey definition table.

| Column | Type | Description |
|--------|------|-------------|
| SURVEY_ID | int | Primary key |
| TITLE | varchar | Survey title |
| SHORTNAME | varchar | URL-friendly alias |
| SURVEY_UID | varchar | Unique identifier (UUID-like) |
| OWNER | int | Owner user ID |
| ISDRAFT | bit | Draft status |
| ISPUBLISHED | bit | Published status |
| ISACTIVE | bit | Active status |
| SURVEY_CREATED | datetime | Creation date |
| SURVEY_START_DATE | datetime | Start date |
| SURVEY_END_DATE | datetime | End date |

### ANSWERS_SET
Links respondents to surveys. One row per survey response.

| Column | Type | Description |
|--------|------|-------------|
| ANSWER_SET_ID | int | Primary key |
| SURVEY_ID | int | FK to SURVEYS |
| UNIQUECODE | varchar | UUID identifying this response (used for PDF generation) |
| ANSWER_SET_DATE | datetime | Response date |
| ANSWER_SET_UPDATE | datetime | Last update date |
| ISDRAFT | bit | Draft response |

### ANSWERS
Individual answers within a response set.

| Column | Type | Description |
|--------|------|-------------|
| ANSWER_ID | int | Primary key |
| AS_ID | int | FK to ANSWERS_SET.ANSWER_SET_ID |
| PA_ID | int | 0 = identity/free-text field |
| VALUE | longtext | Answer value (text, email, or element ID) |
| QUESTION_ID | int | FK to question element |

### ELEMENTS
Survey elements (questions, sections, etc.)

| Column | Type | Description |
|--------|------|-------------|
| ID | int | Primary key |
| ETITLE | varchar | Element title/label |
| ETYPE | varchar | Element type |

### SURVEYS_ELEMENTS
Maps surveys to their elements (join table).

## Common Queries

### List all published surveys
```sql
SELECT SURVEY_ID, TITLE, SHORTNAME, SURVEY_UID,
       SURVEY_CREATED, SURVEY_START_DATE, SURVEY_END_DATE
FROM SURVEYS
WHERE ISPUBLISHED = 1
ORDER BY SURVEY_CREATED DESC;
```

### Count answers per survey
```sql
SELECT s.SURVEY_ID, s.TITLE, COUNT(a.ANSWER_SET_ID) as num_answers
FROM SURVEYS s
LEFT JOIN ANSWERS_SET a ON a.SURVEY_ID = s.SURVEY_ID
WHERE s.ISPUBLISHED = 1
GROUP BY s.SURVEY_ID, s.TITLE
ORDER BY num_answers DESC;
```

### List answer sets for a survey
```sql
SELECT a_set.ANSWER_SET_ID, a_set.UNIQUECODE, a_set.ANSWER_SET_DATE,
       a.VALUE as email
FROM ANSWERS_SET a_set
LEFT JOIN ANSWERS a ON a.AS_ID = a_set.ANSWER_SET_ID AND a.PA_ID = 0
WHERE a_set.SURVEY_ID = {SURVEY_ID}
ORDER BY a_set.ANSWER_SET_DATE DESC;
```

### Look up UNIQUECODE by email + survey ID
```sql
SELECT a_set.ANSWER_SET_ID, a_set.UNIQUECODE
FROM ANSWERS_SET a_set
JOIN ANSWERS a ON a.AS_ID = a_set.ANSWER_SET_ID
WHERE a_set.SURVEY_ID = {SURVEY_ID}
  AND a.PA_ID = 0
  AND a.VALUE = '{EMAIL}'
ORDER BY a_set.ANSWER_SET_DATE DESC
LIMIT 1;
```

### Get all answers for a specific answer set
```sql
SELECT a.ANSWER_ID, a.PA_ID, a.VALUE, a.QUESTION_ID, e.ETITLE
FROM ANSWERS a
LEFT JOIN ELEMENTS e ON e.ID = a.VALUE
WHERE a.AS_ID = {ANSWER_SET_ID}
ORDER BY a.ANSWER_ID;
```

### Get survey element structure
```sql
SELECT e.ID, e.ETITLE, e.ETYPE
FROM SURVEYS_ELEMENTS se
JOIN ELEMENTS e ON e.ID = se.elements_ID
WHERE se.SURVEYS_SURVEY_ID = {SURVEY_ID}
ORDER BY se.elements_ORDER;
```

## Known Survey IDs

| Survey ID | Name | Language | Elements | Type |
|-----------|------|----------|----------|------|
| 4578 | Check4Skills | Romanian (RO) | 80 | RIASEC |
| 4584 | Check4Skills | English (EN) | 77 | RIASEC |
| 4580 | Check4Skills | Dutch (NL) | 76 | RIASEC |
| 4583 | Check4Skills | Italian (IT) | 77 | RIASEC |
| 4579 | Check4Skills | German (DM) | 81 | RIASEC |
| 4609 | Check4TechnicalSkills | Romanian | 63 | Technical Skills |

**Note**: Survey 4609 (Check4TechnicalSkills) is a separate instrument from the Check4Skills RIASEC surveys. It has 2 identity + 52 Likert items + no demographics. It cannot be processed through the RIASEC pipeline.
