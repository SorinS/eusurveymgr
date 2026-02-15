# EUSurvey API Reference

## Instance Details

- **URL**: https://eusurvey.escoaladevalori.ro/eusurvey
- **Platform**: EUSurvey (Java/Spring/Tomcat 8.0.15, Java 8u281)
- **Source**: https://github.com/EUSurvey/EUSURVEY

## Authentication

### HTTP Basic Auth (WebService API)
Used for all `/webservice/*` endpoints.
- **User**: root
- **Password**: Jc9CA76b3Y5TQ7fS

```bash
curl -u root:Jc9CA76b3Y5TQ7fS "$URL/webservice/getMySurveys"
```

### Form-based Session Auth (Web UI + PDF endpoints)
Used for `/worker/*`, `/pdf/*`, and other web endpoints. Requires CSRF token.

```bash
# 1. Get CSRF token from login page
CSRF=$(curl -s -c cookies.txt "$URL/auth/login" \
  | grep '_csrf.*content=' | head -1 \
  | sed 's/.*content="\([^"]*\)".*/\1/')

# 2. Login (returns 302 on success)
curl -s -b cookies.txt -c cookies.txt -o /dev/null -w "%{http_code}" \
  "$URL/login" \
  -d "username=root&password=Jc9CA76b3Y5TQ7fS&_csrf=$CSRF"

# 3. Use session cookies for subsequent requests
curl -s -b cookies.txt "$URL/worker/createanswerpdf/{UNIQUECODE}"
```

**Key fields**: `username`, `password`, `_csrf` (POST to `/login`)
**CSRF source**: `<meta name="_csrf" content="TOKEN"/>` in the HTML of `/auth/login`
**Success indicator**: HTTP 302 redirect (or 200)

### Database Setting Required

The WebService API is gated by a database setting. It must be enabled:

```sql
-- Check current state
SELECT * FROM SETTINGS WHERE SETTINGS_KEY = 'DisableWebserviceAPI';

-- Enable the API (set to 'false')
INSERT INTO SETTINGS (SETTINGS_KEY, SETTINGS_VALUE, SETTINGS_FORMAT)
VALUES ('DisableWebserviceAPI', 'false', 'true / false');
-- or UPDATE if it already exists:
UPDATE SETTINGS SET SETTINGS_VALUE = 'false' WHERE SETTINGS_KEY = 'DisableWebserviceAPI';
```

Requires Tomcat restart after change. Rate limit controlled by `webservice.maxrequestsperday` in `spring.properties`.

---

## WebService API Endpoints (HTTP Basic Auth)

All under `/webservice/*`. Require HTTP Basic Auth.

### Surveys

| Method | Endpoint | Status | Description |
|--------|----------|--------|-------------|
| GET | `/webservice/getMySurveys` | 200 | List all surveys for the authenticated user |
| GET | `/webservice/getSurveyMetadata/{alias}` | 200 | Get survey metadata by alias/shortname |
| GET | `/webservice/getSurveyPDF/{alias}` | 200 | Download survey form as PDF (binary) |

**`getMySurveys` response** (XML):
```xml
<Surveys user='root'>
  <Survey uid='UUID' alias='ShortName'>
    <Title>Survey Title</Title>
  </Survey>
  ...
</Surveys>
```

**`getSurveyMetadata` response** (XML):
```xml
<Survey id='4537' alias='Check4SkillsInEnglish'>
  <SurveyType>Quiz</SurveyType>
  <Title>Check4Skills in English</Title>
  <PivotLanguage>EN</PivotLanguage>
  <Contact>email@example.com</Contact>
  <Status>published</Status>
  <Start>Unset</Start>
  <End>Unset</End>
  <Results>0</Results>
  <Security>open</Security>
  <Visibility>private</Visibility>
  ...
</Survey>
```

**Note**: XML may contain invalid UTF-8 (Latin-1 Romanian diacritics). Client must sanitize before parsing.

**Note**: `getMySurveys` was historically documented as `getSurveys` — that endpoint does **not** exist.

### Results Export (Async)

| Method | Endpoint | Status | Description |
|--------|----------|--------|-------------|
| GET | `/webservice/prepareResults/{formid}/{showids}` | **201** | Start async export. Body = task ID (plain text) |
| GET | `/webservice/getResults/{taskid}` | 200 or **204** | Fetch export results |

**`{formid}`** accepts both numeric survey ID (e.g. `4537`) and alias (e.g. `Check4SkillsInEnglish`).

**`{showids}`** is `true` or `false` — whether to include answer set IDs.

**Polling flow**:
1. `prepareResults` → HTTP **201 Created**, body is the task ID (e.g. `3`)
2. Poll `getResults/{taskid}`:
   - HTTP **204 No Content** → export still in progress, retry after delay
   - HTTP **200 OK** → export complete, body is XML data
   - HTTP **412 Precondition Failed** → survey has no results or does not exist

**Important**: `prepareResults` returns 201 (not 200). `getResults` returns 204 while processing.

There are also PDF result variants (not currently used by eusurveymgr):

| Method | Endpoint | Status | Description |
|--------|----------|--------|-------------|
| GET | `/webservice/prepareResultsPDF/{formid}` | 201 | Start async PDF results export |

PDF results are retrieved via the same `getResults/{taskid}` endpoint.

### Token Management

Tokens are managed through token *groups* (lists), not individually by survey name.

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/webservice/createNewTokenList/{shortname}/{active}` | Create a new token group for a survey |
| GET | `/webservice/createTokens/{groupid}/{number}` | Batch-create tokens in a group |
| GET | `/webservice/activateToken/{groupid}/{token}` | Activate a token |
| GET | `/webservice/deactivateToken/{groupid}/{token}` | Deactivate a token |
| GET | `/webservice/deleteToken/{groupid}/{token}` | Delete a token |

**Note**: There is **no** `getTokens` or `createToken` (singular) endpoint in the upstream EUSurvey source. The token commands in eusurveymgr currently use non-existent endpoints (`getTokens/{surveyname}`, `createToken/{surveyname}`) and will return 404. These need to be reworked to use the group-based API above.

---

## Web Endpoints (Session Auth)

These require form-based login (see Authentication above). Used by the `pdf answer` command.

### Session Login

| Method | Endpoint | Status | Description |
|--------|----------|--------|-------------|
| GET | `/auth/login` | 200 | Get login page (extract CSRF from `<meta name="_csrf">`) |
| POST | `/login` | 302/200 | Submit login form (`username`, `password`, `_csrf`) |

### Answer PDF Generation (3-step flow)

```
Step 1: Login (POST /login with CSRF)
Step 2: GET /worker/createanswerpdf/{UNIQUECODE}  → returns "OK"
Step 3: GET /pdf/answer/{UNIQUECODE}               → returns PDF binary
```

| Method | Endpoint | Status | Description |
|--------|----------|--------|-------------|
| GET | `/worker/createanswerpdf/{uniquecode}` | 200 | **Trigger** PDF generation. Body = `"OK"` on success |
| GET | `/pdf/answer/{uniquecode}` | 200 | **Download** generated answer PDF (binary) |
| GET | `/pdf/answerready/{uniquecode}` | 200 | **Check** if PDF exists. Body = `"exists"` if ready, other values if not |

**Important distinctions**:
- `/pdf/answer/` only SERVES existing PDFs — does NOT generate them
- `/pdf/answerready/` only CHECKS existence — does NOT generate them
- `/worker/createanswerpdf/` is what actually TRIGGERS generation

### Other Useful Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/pdf/survey/{id}` | Download survey form PDF (by numeric ID) |
| GET | `/management/list` | List surveys in management console |
| GET | `/forms/{shortname}` | Access survey by shortname |
| GET | `/preparecontribution/{uniquecode}` | Get answer HTML (rendered survey response) |

---

## eusurveymgr Client → API Mapping

| CLI Command | Client Method | API Endpoint | Auth |
|-------------|--------------|--------------|------|
| `surveys list` | `GetSurveys()` | `GET /webservice/getMySurveys` | Basic |
| `surveys info --alias X` | `GetSurveyMetadata(alias)` | `GET /webservice/getSurveyMetadata/{alias}` | Basic |
| `results export --id X` | `PrepareResults(id, showIDs)` + `GetResults(taskID)` | `GET /webservice/prepareResults/{id}/{showids}` → poll `getResults/{taskid}` | Basic |
| `pdf survey --alias X` | `GetSurveyPDF(alias)` | `GET /webservice/getSurveyPDF/{alias}` | Basic |
| `pdf answer --code X` | `CreateAnswerPDF(code)` + `DownloadAnswerPDF(code)` | `GET /worker/createanswerpdf/{code}` → `GET /pdf/answer/{code}` | Session |
| `pdf answer --email X --survey Y` | DB lookup → same as `--code` | DB query → same flow | Session + DB |
| `tokens list --survey X` | `GetTokens(name)` | `GET /webservice/getTokens/{name}` (**BROKEN** — endpoint doesn't exist) | Basic |
| `tokens create --survey X` | `CreateToken(name)` | `GET /webservice/createToken/{name}` (**BROKEN** — endpoint doesn't exist) | Basic |
| `db surveys` | Direct MySQL | `SELECT` from `SURVEYS` (latest version per UID) | DB |
| `db answers --survey X` | Direct MySQL | `SELECT` from `ANSWERS_SET` + PA_ID=0 identity | DB |
| `db lookup --email X --survey Y` | Direct MySQL | `SELECT` from `ANSWERS_SET` + `ANSWERS` | DB |
| `db responses --email X --survey Y` | Direct MySQL | Lookup → `SELECT` from `ANSWERS` + `ELEMENTS` | DB |

---

## Database Queries (used by `db` commands)

### List Surveys
```sql
SELECT SURVEY_ID, SURVEY_TITLE, SURVEY_SHORTNAME, SURVEY_ISPUBLISHED,
       (SELECT COUNT(*) FROM ANSWERS_SET WHERE SURVEY_ID = s.SURVEY_ID) AS num_answers,
       SURVEY_CREATED
FROM SURVEYS s
ORDER BY SURVEY_ID;
```

### List Answer Sets
```sql
SELECT a_set.ANSWER_SET_ID, a_set.UNIQUECODE, a_set.ANSWER_SET_DATE, a.VALUE AS email
FROM ANSWERS_SET a_set
LEFT JOIN ANSWERS a ON a.AS_ID = a_set.ANSWER_SET_ID AND a.PA_ID = 0
WHERE a_set.SURVEY_ID = {SURVEY_ID}
ORDER BY a_set.ANSWER_SET_DATE DESC;
```

### Lookup UNIQUECODE by Email
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

### Get Responses for a Respondent
```sql
SELECT a.PA_ID, e.ETITLE as question, a.VALUE
FROM ANSWERS a
LEFT JOIN ELEMENTS e ON e.ELEM_UID = a.PA_UID
WHERE a.AS_ID = {ANSWER_SET_ID}
ORDER BY a.PA_ID, a.ANSWER_ID;
```

---

## PDF File Storage

Generated PDFs are stored on the server at:
```
/home/eusurvey/eusurveytemp/surveys/{first-char-of-uid}/{survey-uid}/EXPORTS/answer{uniquecode}.pdf
```

Example:
```
/home/eusurvey/eusurveytemp/surveys/f/f1234.../EXPORTS/answerae8d5fec-daaf-4aba-b860-544d1f717d8a.pdf
```

## Server Tuning

### Export Thread Pools

Config file: `/home/eusurvey/apache-tomcat-8.0.15/webapps/eusurvey/WEB-INF/spring.properties`

```properties
export.poolSize=2        # default; increase to 4-5 for better throughput
```

This controls two separate `Executors.newFixedThreadPool()` pools in `BasicService.java`:
- **PDF export pool** (`getPDFPool()`) — bulk PDF result exports
- **General export pool** (`getPool()`) — XML/CSV result exports

Both are lazy-initialized and sized to `export.poolSize`. Requires Tomcat restart.

**Upper bound**: Each thread uses a `PDFRenderer` from a pool capped at 15 instances (`PDFService.java`), so setting `export.poolSize` above ~8 provides diminishing returns.

Other thread pools (not controlled by this property):

| Bean | Threads | Purpose |
|------|---------|---------|
| `taskExecutor` | 10 core / 10 max | Individual answer PDF generation (`/worker/createanswerpdf`) |
| `taskExecutorLong` | 1 core / 5 max | Long-running operations |
| `executorWithPoolSizeRange` | 2 core / 4 max | Spring `@Async` methods |

These are configured in `WEB-INF/spring/mvc-dispatcher-servlet.xml`.

### Rate Limiting

```properties
webservice.maxrequestsperday=100
```

Controls how many WebService API requests a user can make per day.

## JAXB Fix (Java 8)

The EUSurvey PDF generation uses Flying Saucer + iText which depends on JAXB. On this instance (Java 8u281), the webapp had `jaxb-core-2.2.11.jar` and `jaxb-impl-2.2.11.jar` but was missing `jaxb-api-2.2.11.jar`, causing `NoClassDefFoundError: javax/xml/bind/JAXBException`.

**Fix applied**: Added `jaxb-api-2.2.11.jar` to `WEB-INF/lib` (must match the 2.2.11 version, NOT 2.3.x which requires Java 9+).
